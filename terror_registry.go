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
	"fmt"
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

type ErrorID string
type ErrClassID int
type RFCErrorCode string

// NewRegistry make a registry, where ErrClasses register to.
// One component should create only one registry, named by the RFC.
// For TiDB ecosystem components, when creating registry,
// please use the component name to name the registry, see below example:
//
// TiKV: KV
// PD: PD
// DM: DM
// BR: BR
// TiCDC: CDC
// Lightning: LN
// TiFlash: FLASH
// Dumpling: DP
func NewRegistry(name string) *Registry {
	return &Registry{Name: name, errClasses: map[ErrClassID]ErrClass{}}
}

// RegisterErrorClass registers new error class for terror.
func (r *Registry) RegisterErrorClass(classCode int, desc string) ErrClass {
	code := ErrClassID(classCode)
	if _, exists := r.errClasses[code]; exists {
		panic(fmt.Sprintf("duplicate register ClassCode %d - %s", code, desc))
	}
	errClass := ErrClass{
		ID:          code,
		Description: desc,
		errors:      map[ErrorID]*Error{},
		registry:    r,
	}
	r.errClasses[code] = errClass
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
		Build()
}

// DefineError is the entrance of the define error DSL,
// simple usage:
// ```
// ClassExecutor.DefineError().
//	TextualCode("ExecutorAbsent").
//	MessageTemplate("executor is taking vacation at %s").
//	Build()
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
