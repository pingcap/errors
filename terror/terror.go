// Copyright 2015 PingCAP, Inc.
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

package terror

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"
)

// ErrCode represents a specific error type in a error class.
// Same error code can be used in different error classes.
type ErrCode int

// ErrCodeText is a textual error code that represents a specific error type in a error class.
// This string should include component name, see
// https://github.com/pingcap/tidb/blob/master/docs/design/2020-05-08-standardize-error-codes-and-messages.md#the-error-code-range
type ErrCodeText string

// ErrClass represents a class of errors.
type ErrClass int

var (
	errClass2Desc = make(map[ErrClass]string)
	// errsOfClass maps some errClass to all errors belongs to it.
	errsOfClass = make(map[ErrClass][]*Error)
)

// RegisterErrorClass registers new error class for terror.
func RegisterErrorClass(classCode int, desc string) ErrClass {
	errClass := ErrClass(classCode)
	if _, exists := errClass2Desc[errClass]; exists {
		panic(fmt.Sprintf("duplicate register ClassCode %d - %s", classCode, desc))
	}
	errClass2Desc[errClass] = desc
	return errClass
}

// String implements fmt.Stringer interface.
func (ec ErrClass) String() string {
	if s, exists := errClass2Desc[ec]; exists {
		return s
	}
	return strconv.Itoa(int(ec))
}

// EqualClass returns true if err is *Error with the same class.
func (ec ErrClass) EqualClass(err error) bool {
	e := errors.Cause(err)
	if e == nil {
		return false
	}
	if te, ok := e.(*Error); ok {
		return te.class == ec
	}
	return false
}

// NotEqualClass returns true if err is not *Error with the same class.
func (ec ErrClass) NotEqualClass(err error) bool {
	return !ec.EqualClass(err)
}

// New defines an *Error with an error code and an error message.
// Usually used to create base *Error.
// Currently, this method is same as Synthesize after extracted from parser.
// deprecated: textual error codes is more readable than numeric error codes, use NewError instead.
func (ec ErrClass) New(code ErrCode, message string) *Error {
	return ec.NewError(code, "", message)
}

// NewError defines an *Error with an error code, its code message and an error message.
// Note this isn't thread-safe.
func (ec ErrClass) NewError(code ErrCode, text ErrCodeText, message string) *Error {
	err := &Error{
		class:    ec,
		code:     code,
		codeText: text,
		message:  message,
	}
	// We might can use a map as a set here to speed up this.
	for _, e := range errsOfClass[ec] {
		if e.Equal(err) {
			log.Panic("replicated error prototype created",
				zap.Int("code", int(code)),
				zap.String("codeText", string(text)))
		}
	}
	errsOfClass[ec] = append(errsOfClass[ec], err)
	return err
}

// AllErrors returns all errors of this ErrClass
// Note this isn't thread-safe.
// You shouldn't modify the returned slice without copying.
func (ec ErrClass) AllErrors() []*Error {
	return errsOfClass[ec]
}

// AllErrorClasses returns all errClasses that has been registered.
// Note this isn't thread-safe.
func AllErrorClasses() []ErrClass {
	all := make([]ErrClass, 0, len(errClass2Desc))
	for errClass := range errClass2Desc {
		all = append(all, errClass)
	}
	return all
}

// Synthesize synthesizes an *Error in the air
// it didn't register error into ErrClassToMySQLCodes
// so it's goroutine-safe
// and often be used to create Error came from other systems like TiKV.
func (ec ErrClass) Synthesize(code ErrCode, message string) *Error {
	return &Error{
		class:   ec,
		code:    code,
		message: message,
	}
}

// Error implements error interface and adds integer Class and Code, so
// errors with different message can be compared.
type Error struct {
	class ErrClass
	code  ErrCode
	// codeText is the textual describe of the error code
	codeText ErrCodeText
	message  string
	args     []interface{}
	file     string
	line     int
}

// Class returns ErrClass
func (e *Error) Class() ErrClass {
	return e.class
}

// Code returns ErrCode
func (e *Error) Code() ErrCode {
	return e.code
}

// MarshalJSON implements json.Marshaler interface.
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Class    ErrClass    `json:"class"`
		Code     ErrCode     `json:"code"`
		CodeText ErrCodeText `json:"codeText"`
		Msg      string      `json:"message"`
	}{
		Class:    e.class,
		Code:     e.code,
		Msg:      e.getMsg(),
		CodeText: e.codeText,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (e *Error) UnmarshalJSON(data []byte) error {
	err := &struct {
		Class    ErrClass    `json:"class"`
		Code     ErrCode     `json:"code"`
		Msg      string      `json:"message"`
		CodeText ErrCodeText `json:"codeText"`
	}{}

	if err := json.Unmarshal(data, &err); err != nil {
		return errors.Trace(err)
	}

	e.class = err.Class
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

// Error implements error interface.
func (e *Error) Error() string {
	describe := e.codeText
	if len(describe) == 0 {
		describe = ErrCodeText(strconv.Itoa(int(e.code)))
	}
	return fmt.Sprintf("[%s:%s]%s", e.class, describe, e.getMsg())
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
	return errors.AddStack(&err)
}

// GenWithStackByArgs generates a new *Error with the same class and code, and new arguments.
func (e *Error) GenWithStackByArgs(args ...interface{}) error {
	err := *e
	err.args = args
	return errors.AddStack(&err)
}

// FastGen generates a new *Error with the same class and code, and a new formatted message.
// This will not call runtime.Caller to get file and line.
func (e *Error) FastGen(format string, args ...interface{}) error {
	err := *e
	err.message = format
	err.args = args
	return errors.SuspendStack(&err)
}

// FastGen generates a new *Error with the same class and code, and a new arguments.
// This will not call runtime.Caller to get file and line.
func (e *Error) FastGenByArgs(args ...interface{}) error {
	err := *e
	err.args = args
	return errors.SuspendStack(&err)
}

// Equal checks if err is equal to e.
func (e *Error) Equal(err error) bool {
	originErr := errors.Cause(err)
	if originErr == nil {
		return false
	}

	if error(e) == originErr {
		return true
	}
	inErr, ok := originErr.(*Error)
	return ok && e.class == inErr.class && e.code == inErr.code && e.codeText == inErr.codeText
}

// NotEqual checks if err is not equal to e.
func (e *Error) NotEqual(err error) bool {
	return !e.Equal(err)
}

// ErrorEqual returns a boolean indicating whether err1 is equal to err2.
func ErrorEqual(err1, err2 error) bool {
	e1 := errors.Cause(err1)
	e2 := errors.Cause(err2)

	if e1 == e2 {
		return true
	}

	if e1 == nil || e2 == nil {
		return e1 == e2
	}

	te1, ok1 := e1.(*Error)
	te2, ok2 := e2.(*Error)
	if ok1 && ok2 {
		return te1.class == te2.class && te1.code == te2.code
	}

	return e1.Error() == e2.Error()
}

// ErrorNotEqual returns a boolean indicating whether err1 isn't equal to err2.
func ErrorNotEqual(err1, err2 error) bool {
	return !ErrorEqual(err1, err2)
}

// MustNil cleans up and fatals if err is not nil.
func MustNil(err error, closeFuns ...func()) {
	if err != nil {
		for _, f := range closeFuns {
			f()
		}
		log.Fatal("unexpected error", zap.Error(err))
	}
}

// Call executes a function and checks the returned err.
func Call(fn func() error) {
	err := fn()
	if err != nil {
		log.Error("function call errored", zap.Error(err))
	}
}

// Log logs the error if it is not nil.
func Log(err error) {
	if err != nil {
		log.Error("encountered error", zap.Error(errors.WithStack(err)))
	}
}
