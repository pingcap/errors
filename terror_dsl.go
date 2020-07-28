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
	"github.com/pingcap/log"
	"go.uber.org/zap"
)

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

// Build ends the define of the error.
func (b *Builder) Build() *Error {
	if ok := b.class.RegisterError(b.err); !ok {
		log.Panic("replicated error prototype created",
			zap.String("ID", string(b.err.ID())),
			zap.String("RFCCode", string(b.err.RFCCode())))
	}
	return b.err
}
