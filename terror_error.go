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

// RedactLogEnabled defines whether the arguments of Error need to be redacted.
var RedactLogEnabled bool = false

// ErrCode represents a specific error type in a error class.
// Same error code can be used in different error classes.
type ErrCode int

// ErrCodeText is a textual error code that represents a specific error type in a error class.
type ErrCodeText string

type ErrorID string
type RFCErrorCode string

// class2RFCCode is used for compatible with old version of TiDB. When
// marshal Error to json, old version of TiDB contain a 'class' field
// which is represented for error class. In order to parser and convert
// json to errors.Error, using this map to convert error class to RFC
// error code text. here is reference:
// https://github.com/pingcap/parser/blob/release-3.0/terror/terror.go#L58
var class2RFCCode = map[int]string{
	1:  "autoid",
	2:  "ddl",
	3:  "domain",
	4:  "evaluator",
	5:  "executor",
	6:  "expression",
	7:  "admin",
	8:  "kv",
	9:  "meta",
	10: "planner",
	11: "parser",
	12: "perfschema",
	13: "privilege",
	14: "schema",
	15: "server",
	16: "struct",
	17: "variable",
	18: "xeval",
	19: "table",
	20: "types",
	21: "global",
	22: "mocktikv",
	23: "json",
	24: "tikv",
	25: "session",
	26: "plugin",
	27: "util",
}
var rfcCode2class map[string]int

func init() {
	rfcCode2class = make(map[string]int)
	for k, v := range class2RFCCode {
		rfcCode2class[v] = k
	}
}

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
	code ErrCode
	// codeText is the textual describe of the error code
	codeText ErrCodeText
	// message is a template of the description of this error.
	// printf-style formatting is enabled.
	message string
	// The workaround field: how to work around this error.
	// It's used to teach the users how to solve the error if occurring in the real environment.
	workaround string
	// description is the expanded detail of why this error occurred.
	// This could be written by developer at a static env,
	// and the more detail this field explaining the better,
	// even some guess of the cause could be included.
	description string
	// redactArgsPos defines the positions of arguments in message that need to be redacted.
	// And it is controlled by the global var RedactLogEnabled.
	// For example, an original error is `Duplicate entry 'PRIMARY' for key 'key'`,
	// when RedactLogEnabled is ON and redactArgsPos is [0, 1], the error is `Duplicate entry '?' for key '?'`.
	redactArgsPos []int
	// Cause is used to warp some third party error.
	cause error
	args  []interface{}
	file  string
	line  int
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
	return RFCErrorCode(e.ID())
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
	if e == nil {
		return "<nil>"
	}
	describe := e.codeText
	if len(describe) == 0 {
		describe = ErrCodeText(strconv.Itoa(int(e.code)))
	}
	return fmt.Sprintf("[%s]%s", e.RFCCode(), e.GetMsg())
}

func (e *Error) GetMsg() string {
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
	// TODO: RedactErrorArg
	err := *e
	err.message = format
	err.args = args
	err.fillLineAndFile(1)
	return AddStack(&err)
}

// GenWithStackByArgs generates a new *Error with the same class and code, and new arguments.
func (e *Error) GenWithStackByArgs(args ...interface{}) error {
	RedactErrorArg(args, e.redactArgsPos)
	err := *e
	err.args = args
	err.fillLineAndFile(1)
	return AddStack(&err)
}

// FastGen generates a new *Error with the same class and code, and a new formatted message.
// This will not call runtime.Caller to get file and line.
func (e *Error) FastGen(format string, args ...interface{}) error {
	// TODO: RedactErrorArg
	err := *e
	err.message = format
	err.args = args
	return SuspendStack(&err)
}

// FastGen generates a new *Error with the same class and code, and a new arguments.
// This will not call runtime.Caller to get file and line.
func (e *Error) FastGenByArgs(args ...interface{}) error {
	RedactErrorArg(args, e.redactArgsPos)
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
	idEquals := e.ID() == inErr.ID()
	return idEquals
}

// NotEqual checks if err is not equal to e.
func (e *Error) NotEqual(err error) bool {
	return !e.Equal(err)
}

// RedactErrorArg redacts the args by position if RedactLogEnabled is enabled.
func RedactErrorArg(args []interface{}, position []int) {
	if RedactLogEnabled {
		for _, pos := range position {
			if len(args) > pos {
				args[pos] = "?"
			}
		}
	}
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
	// Deprecated field, please use `RFCCode` instead.
	Class   int    `json:"class"`
	Code    int    `json:"code"`
	Msg     string `json:"message"`
	RFCCode string `json:"rfccode"`
}

// MarshalJSON implements json.Marshaler interface.
// aware that this function cannot save a 'registered' status,
// since we cannot access the registry when unmarshaling,
// and the original global registry would be removed here.
// This function is reserved for compatibility.
func (e *Error) MarshalJSON() ([]byte, error) {
	ec := strings.Split(string(e.codeText), ":")[0]
	return json.Marshal(&jsonError{
		Class:   rfcCode2class[ec],
		Code:    int(e.code),
		Msg:     e.GetMsg(),
		RFCCode: string(e.codeText),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
// aware that this function cannot create a 'registered' error,
// since we cannot access the registry in this context,
// and the original global registry is removed.
// This function is reserved for compatibility.
func (e *Error) UnmarshalJSON(data []byte) error {
	tErr := &jsonError{}
	if err := json.Unmarshal(data, &tErr); err != nil {
		return Trace(err)
	}
	e.codeText = ErrCodeText(tErr.RFCCode)
	if tErr.RFCCode == "" && tErr.Class > 0 {
		e.codeText = ErrCodeText(class2RFCCode[tErr.Class] + ":" + strconv.Itoa(tErr.Code))
	}
	e.code = ErrCode(tErr.Code)
	e.message = tErr.Msg
	return nil
}

func (e *Error) Wrap(err error) *Error {
	if err != nil {
		newErr := *e
		newErr.cause = err
		return &newErr
	}
	return nil
}

func (e *Error) Cause() error {
	root := Unwrap(e.cause)
	if root == nil {
		return e.cause
	}
	return root
}

func (e *Error) FastGenWithCause(args ...interface{}) error {
	err := *e
	if e.cause != nil {
		err.message = e.cause.Error()
	}
	err.args = args
	return SuspendStack(&err)
}

func (e *Error) GenWithStackByCause(args ...interface{}) error {
	err := *e
	if e.cause != nil {
		err.message = e.cause.Error()
	}
	err.args = args
	err.fillLineAndFile(1)
	return AddStack(&err)
}

type NormalizeOption func(*Error)

// Description returns a NormalizeOption to set description.
func Description(desc string) NormalizeOption {
	return func(e *Error) {
		e.description = desc
	}
}

// Workaround returns a NormalizeOption to set workaround.
func Workaround(wr string) NormalizeOption {
	return func(e *Error) {
		e.workaround = wr
	}
}

func RedactArgs(pos []int) NormalizeOption {
	return func(e *Error) {
		e.redactArgsPos = pos
	}
}

// RFCCodeText returns a NormalizeOption to set RFC error code.
func RFCCodeText(codeText string) NormalizeOption {
	return func(e *Error) {
		e.codeText = ErrCodeText(codeText)
	}
}

// MySQLErrorCode returns a NormalizeOption to set error code.
func MySQLErrorCode(code int) NormalizeOption {
	return func(e *Error) {
		e.code = ErrCode(code)
	}
}

// Normalize creates a new Error object.
func Normalize(message string, opts ...NormalizeOption) *Error {
	e := &Error{
		message: message,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}
