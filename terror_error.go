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
	"runtime"
	"strconv"
	"strings"
)

// Error is the 'prototype' of a type of errors.
// Use DefineError to make a *Error:
// var ErrUnavailable = ClassRegion.DefineError().
//		TextualCode("Unavailable").
//		Description("A certain Raft Group is not available, such as the number of replicas is not enough.\n" +
//			"This error usually occurs when the TiKV server is busy or the TiKV node is down.").
//		Workaround("Check the status, monitoring data and log of the TiKV server.").
//		MessageTemplate("Region %d is unavailable").
//		Build()
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

// Code returns ErrorCode, by the RFC:
//
// The error code is a 3-tuple of abbreviated component name, error class and error code,
// joined by a colon like {Component}:{ErrorClass}:{InnerErrorCode}.
func (e *Error) RFCCode() RFCErrorCode {
	ec := e.Class()
	if ec == nil {
		return RFCErrorCode(e.ID())
	}
	reg := ec.registry
	// Maybe top-level errors.
	if reg.Name == "" {
		return RFCErrorCode(fmt.Sprintf("%s:%s",
			ec.Description,
			e.ID(),
		))
	}
	return RFCErrorCode(fmt.Sprintf("%s:%s:%s",
		reg.Name,
		ec.Description,
		e.ID(),
	))
}

// ID returns the ID of this error.
func (e *Error) ID() ErrorID {
	if e.codeText != "" {
		return ErrorID(e.codeText)
	}
	return ErrorID(strconv.Itoa(int(e.code)))
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
	return fmt.Sprintf("[%s] %s", e.RFCCode(), e.getMsg())
}

func (e *Error) getMsg() string {
	if len(e.args) > 0 {
		return fmt.Sprintf(e.message, e.args...)
	}
	return e.message
}

func (e *Error) fillLineAndFile(skip int) {
	// skip this
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		e.file = "<unknown>"
		e.line = -1
		return
	}
	e.file = file
	e.line = line
}

// GenWithStack generates a new *Error with the same class and code, and a new formatted message.
func (e *Error) GenWithStack(format string, args ...interface{}) error {
	err := *e
	err.message = format
	err.args = args
	err.fillLineAndFile(1)
	return AddStack(&err)
}

// GenWithStackByArgs generates a new *Error with the same class and code, and new arguments.
func (e *Error) GenWithStackByArgs(args ...interface{}) error {
	err := *e
	err.args = args
	err.fillLineAndFile(1)
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

type jsonError struct {
	RFCCode     RFCErrorCode `json:"code"`
	Error       string       `json:"message"`
	Description string       `json:"description,omitempty"`
	Workaround  string       `json:"workaround,omitempty"`
	Class       ErrClassID   `json:"classID"`
	File        string       `json:"file"`
	Line        int          `json:"line"`
}

// MarshalJSON implements json.Marshaler interface.
// aware that this function cannot save a 'registered' status,
// since we cannot access the registry when unmarshaling,
// and the original global registry would be removed here.
// This function is reserved for compatibility.
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonError{
		Error:       e.getMsg(),
		Description: e.Description,
		Workaround:  e.Workaround,
		RFCCode:     e.RFCCode(),
		Class:       e.class.ID,
		Line:        e.line,
		File:        e.file,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
// aware that this function cannot create a 'registered' error,
// since we cannot access the registry in this context,
// and the original global registry is removed.
// This function is reserved for compatibility.
func (e *Error) UnmarshalJSON(data []byte) error {
	err := &jsonError{}

	if err := json.Unmarshal(data, &err); err != nil {
		return Trace(err)
	}
	codes := strings.Split(string(err.RFCCode), ":")
	regName := codes[0]
	className := codes[1]
	innerCode := codes[2]
	if i, errAtoi := strconv.Atoi(innerCode); errAtoi == nil {
		e.code = ErrCode(i)
	} else {
		e.codeText = ErrCodeText(innerCode)
	}
	e.line = err.Line
	e.file = err.File
	e.message = err.Error
	e.class = &ErrClass{
		Description: className,
		ID:          err.Class,
		registry:    &Registry{Name: regName},
	}
	return nil
}
