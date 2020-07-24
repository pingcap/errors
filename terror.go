// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"encoding/json"
	"fmt"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"strconv"
)

// Registry is a set of errors for a component.
type Registry struct {
	Name       string
	errClasses map[ErrClassID]ErrClass
}

// ErrCode represents a specific error type in a error class.
// Same error code can be used in different error classes.
type ErrCode int

// ErrCodeText is a textual error code that represents a specific error type in a error class.
type ErrCodeText string

// ErrClass represents a class of errors.
// You can create error 'prototypes' of this class.
type ErrClass struct {
	ID          ErrClassID
	Description string
	errors      map[ErrorID]*Error
	registry    *Registry
}

type ErrorID = string
type ErrClassID = int
type RFCErrorCode = string

// NewRegistry make a registry, where ErrClasses register to.
func NewRegistry(name string) *Registry {
	return &Registry{Name: name, errClasses: map[ErrClassID]ErrClass{}}
}

// RegisterErrorClass registers new error class for terror.
func (r *Registry) RegisterErrorClass(classCode int, desc string) ErrClass {
	if _, exists := r.errClasses[classCode]; exists {
		panic(fmt.Sprintf("duplicate register ClassCode %d - %s", classCode, desc))
	}
	errClass := ErrClass{
		ID:          classCode,
		Description: desc,
		errors:      map[ErrorID]*Error{},
		registry:    r,
	}
	r.errClasses[classCode] = errClass
	return errClass
}

// String implements fmt.Stringer interface.
func (ec *ErrClass) String() string {
	return ec.Description
}

// Equal tests whether the other error is in this class.
func (ec *ErrClass) Equal(other *ErrClass) bool {
	if ec == nil || other == nil {
		return ec == other
	}
	return ec.ID == other.ID
}

// EqualClass returns true if err is *Error with the same class.
func (ec *ErrClass) EqualClass(err error) bool {
	e := Cause(err)
	if e == nil {
		return false
	}
	if te, ok := e.(*Error); ok {
		return te.class.Equal(ec)
	}
	return false
}

// NotEqualClass returns true if err is not *Error with the same class.
func (ec *ErrClass) NotEqualClass(err error) bool {
	return !ec.EqualClass(err)
}

// New defines an *Error with an error code and an error message.
// Usually used to create base *Error.
// This function is reserved for compatibility, if possible, use DefineError instead.
func (ec *ErrClass) New(code ErrCode, message string) *Error {
	return ec.DefineError().
		NumericCode(code).
		MessageTemplate(message).
		Done()
}

// DefineError is the entrance of the define error DSL,
// simple usage:
// ```
// ClassExecutor.DefineError().
//	TextualCode("ExecutorAbsent").
//	MessageTemplate("executor is taking vacation at %s").
//	Done()
// ```
func (ec *ErrClass) DefineError() *Builder {
	return &Builder{
		err:   &Error{},
		class: ec,
	}
}

// RegisterError try to register an error to a class.
// return true if success.
func (ec *ErrClass) RegisterError(err *Error) bool {
	if _, ok := ec.errors[err.ID()]; ok {
		return false
	}
	err.class = ec
	ec.errors[err.ID()] = err
	return true
}

// AllErrors returns all errors of this ErrClass
// Note this isn't thread-safe.
// You shouldn't modify the returned slice without copying.
func (ec *ErrClass) AllErrors() []*Error {
	all := make([]*Error, 0, len(ec.errors))
	for _, err := range ec.errors {
		all = append(all, err)
	}
	return all
}

// AllErrorClasses returns all errClasses that has been registered.
// Note this isn't thread-safe.
func (r *Registry) AllErrorClasses() []ErrClass {
	all := make([]ErrClass, 0, len(r.errClasses))
	for _, errClass := range r.errClasses {
		all = append(all, errClass)
	}
	return all
}

// Synthesize synthesizes an *Error in the air
// it didn't register error into ErrClass
// so it's goroutine-safe
// and often be used to create Error came from other systems like TiKV.
func (ec *ErrClass) Synthesize(code ErrCode, message string) *Error {
	return &Error{
		class:   ec,
		code:    code,
		message: message,
	}
}

// Error is the 'prototype' of all errors you defined.
// Use DefineError to make a *Error:
// var ErrUnavailable = ClassRegion.DefineError().
//		TextualCode("Unavailable").
//		Description("A certain Raft Group is not available, such as the number of replicas is not enough.\n" +
//			"This error usually occurs when the TiKV server is busy or the TiKV node is down.").
//		Workaround("Check the status, monitoring data and log of the TiKV server.").
//		MessageTemplate("Region %d is unavailable").
//		Done()
//
// "throw" it at runtime:
// func Somewhat() error {
//     ...
//     if err != nil {
//         // generate a stackful error use the message template at defining,
//         // also see FastGen(it's stackless), GenWithStack(it uses custom message template).
//         return ErrUnavailable.GenWithStackByArgs(region.ID)
//     }
// }
//
// testing whether an error belongs to a prototype:
// if ErrUnavailable.Equal(err) {
//     // handle this error.
// }
type Error struct {
	class *ErrClass
	code  ErrCode
	// codeText is the textual describe of the error code
	codeText ErrCodeText
	// message is a template of the description of this error.
	// printf-style formatting is enabled.
	message string
	// The workaround field: how to work around this error.
	// It's used to teach the users how to solve the error if occurring in the real environment.
	Workaround string
	// Description is the expanded detail of why this error occurred.
	// This could be written by developer at a static env,
	// and the more detail this field explaining the better,
	// even some guess of the cause could be included.
	Description string
	args        []interface{}
	file        string
	line        int
}

// Class returns ErrClass
func (e *Error) Class() *ErrClass {
	return e.class
}

// Code returns the numeric code of this error.
// ID() will return textual error if there it is,
// when you just want to get the purely numeric error
// (e.g., for mysql protocol transmission.), this would be useful.
func (e *Error) Code() ErrCode {
	return e.code
}

// Code returns InnerErrorCode, by the RFC:
// https://github.com/pingcap/tidb/blob/master/docs/design/2020-05-08-standardize-error-codes-and-messages.md#the-error-code-range
func (e *Error) RFCCode() RFCErrorCode {
	ec := e.Class()
	if ec == nil {
		return e.ID()
	}
	reg := ec.registry
	// Maybe top-level errors.
	if reg.Name == "" {
		return fmt.Sprintf("%s:%s",
			ec.Description,
			e.ID(),
		)
	}
	return fmt.Sprintf("%s:%s:%s",
		reg.Name,
		ec.Description,
		e.ID(),
	)
}

// ID returns the ID of this error.
func (e *Error) ID() ErrorID {
	if e.codeText != "" {
		return string(e.codeText)
	}
	return strconv.Itoa(int(e.code))
}

// MarshalJSON implements json.Marshaler interface.
// aware that this function cannot save a 'registered' status,
// since we cannot access the registry when unmarshaling,
// and the original global registry would be removed here.
// This function is reserved for compatibility.
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Class    ErrClassID  `json:"class"`
		Code     ErrCode     `json:"code"`
		CodeText ErrCodeText `json:"codeText"`
		Msg      string      `json:"message"`
	}{
		Class:    e.class.ID,
		Code:     e.code,
		Msg:      e.getMsg(),
		CodeText: e.codeText,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
// aware that this function cannot create a 'registered' error,
// since we cannot access the registry in this context,
// and the original global registry is removed.
// This function is reserved for compatibility.
func (e *Error) UnmarshalJSON(data []byte) error {
	err := &struct {
		Class    ErrClassID  `json:"class"`
		Code     ErrCode     `json:"code"`
		Msg      string      `json:"message"`
		CodeText ErrCodeText `json:"codeText"`
	}{}

	if err := json.Unmarshal(data, &err); err != nil {
		return Trace(err)
	}

	e.class = &ErrClass{ID: err.Class}
	e.code = err.Code
	e.message = err.Msg
	e.codeText = err.CodeText
	return nil
}

// Location returns the location where the error is created,
// implements juju/errors locationer interface.
func (e *Error) Location() (file string, line int) {
	return e.file, e.line
}

// MessageTemplate returns the error message template of this error.
func (e *Error) MessageTemplate() string {
	return e.message
}

// Error implements error interface.
func (e *Error) Error() string {
	describe := e.codeText
	if len(describe) == 0 {
		describe = ErrCodeText(strconv.Itoa(int(e.code)))
	}
	return fmt.Sprintf("[%s:%s] %s", e.class, describe, e.getMsg())
}

func (e *Error) getMsg() string {
	if len(e.args) > 0 {
		return fmt.Sprintf(e.message, e.args...)
	}
	return e.message
}

// GenWithStack generates a new *Error with the same class and code, and a new formatted message.
func (e *Error) GenWithStack(format string, args ...interface{}) error {
	err := *e
	err.message = format
	err.args = args
	return AddStack(&err)
}

// GenWithStackByArgs generates a new *Error with the same class and code, and new arguments.
func (e *Error) GenWithStackByArgs(args ...interface{}) error {
	err := *e
	err.args = args
	return AddStack(&err)
}

// FastGen generates a new *Error with the same class and code, and a new formatted message.
// This will not call runtime.Caller to get file and line.
func (e *Error) FastGen(format string, args ...interface{}) error {
	err := *e
	err.message = format
	err.args = args
	return SuspendStack(&err)
}

// FastGen generates a new *Error with the same class and code, and a new arguments.
// This will not call runtime.Caller to get file and line.
func (e *Error) FastGenByArgs(args ...interface{}) error {
	err := *e
	err.args = args
	return SuspendStack(&err)
}

// Equal checks if err is equal to e.
func (e *Error) Equal(err error) bool {
	originErr := Cause(err)
	if originErr == nil {
		return false
	}
	if error(e) == originErr {
		return true
	}
	inErr, ok := originErr.(*Error)
	if !ok {
		return false
	}
	classEquals := e.class.Equal(inErr.class)
	idEquals := e.ID() == inErr.ID()
	return classEquals && idEquals
}

// NotEqual checks if err is not equal to e.
func (e *Error) NotEqual(err error) bool {
	return !e.Equal(err)
}

// ErrorEqual returns a boolean indicating whether err1 is equal to err2.
func ErrorEqual(err1, err2 error) bool {
	e1 := Cause(err1)
	e2 := Cause(err2)

	if e1 == e2 {
		return true
	}

	if e1 == nil || e2 == nil {
		return e1 == e2
	}

	te1, ok1 := e1.(*Error)
	te2, ok2 := e2.(*Error)
	if ok1 && ok2 {
		return te1.Equal(te2)
	}

	return e1.Error() == e2.Error()
}

// ErrorNotEqual returns a boolean indicating whether err1 isn't equal to err2.
func ErrorNotEqual(err1, err2 error) bool {
	return !ErrorEqual(err1, err2)
}

// Builder is the builder of Error.
type Builder struct {
	err   *Error
	class *ErrClass
}

// TextualCode is is the textual identity of this error internally,
// note that this error code can only be duplicated in different Registry or ErrorClass.
func (b *Builder) TextualCode(text ErrCodeText) *Builder {
	b.err.codeText = text
	return b
}

// NumericCode is is the numeric identity of this error internally,
// note that this error code can only be duplicated in different Registry or ErrorClass.
func (b *Builder) NumericCode(num ErrCode) *Builder {
	b.err.code = num
	return b
}

// Description is the expanded detail of why this error occurred.
// This could be written by developer at a static env,
// and the more detail this field explaining the better,
// even some guess of the cause could be included.
func (b *Builder) Description(desc string) *Builder {
	b.err.Description = desc
	return b
}

// Workaround shows how to work around this error.
// It's used to teach the users how to solve the error if occurring in the real environment.
func (b *Builder) Workaround(wd string) *Builder {
	b.err.Workaround = wd
	return b
}

// MessageTemplate is the template of the error string that can be formatted after
// calling `GenWithArgs` method.
// currently, printf style template is used.
func (b *Builder) MessageTemplate(template string) *Builder {
	b.err.message = template
	return b
}

// Done ends the define of the error.
func (b *Builder) Done() *Error {
	if ok := b.class.RegisterError(b.err); !ok {
		log.Panic("replicated error prototype created",
			zap.String("ID", b.err.ID()),
			zap.String("RFCCode", b.err.RFCCode()))
	}
	return b.err
}
