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
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"io"
	"regexp"
)

const tomlTemplate = `[error.%s]
error = '''%s'''
description = '''%s'''
workaround = '''%s'''
`

func (e *Error) exportTo(writer io.Writer) error {
	desc := e.Description
	if e.Description == "" {
		log.Warn("error description missed", zap.String("error", string(e.RFCCode())))
		desc = "N/A"
	}
	workaround := e.Workaround
	if e.Workaround == "" {
		log.Warn("error workaround missed", zap.String("error", string(e.RFCCode())))
		workaround = "N/A"
	}
	_, err := fmt.Fprintf(writer, tomlTemplate, e.RFCCode(), replaceFlags(e.MessageTemplate()), desc, workaround)
	return err
}

func (ec *ErrClass) exportTo(writer io.Writer) error {
	for _, e := range ec.AllErrors() {
		if err := e.exportTo(writer); err != nil {
			return err
		}
	}
	return nil
}

// ExportTo export the registry to a writer, as toml format from the RFC.
func (r *Registry) ExportTo(writer io.Writer) error {
	for _, ec := range r.errClasses {
		if err := ec.exportTo(writer); err != nil {
			return err
		}
	}
	return nil
}

// CPP printf flag with minimal support for explicit argument indexes.
// introductory % character: %
// (optional) one or more flags that modify the behavior of the conversion: [+\-# 0]?
// (optional) integer value or * that specifies minimum field width: ([0-9]+|(\[[0-9]+])?\*)?
// (optional) . followed by integer number or *, or neither that specifies precision of the conversion: (\.([0-9]+|(\[[0-9]+])?\*))?
//     the prepending (\[[0-9]+])? is for golang explicit argument indexes:
//     The same notation before a '*' for a width or precision selects the argument index holding the value.
// (optional) the notation [n] immediately before the verb indicates
//     that the nth one-indexed argument is to be formatted instead: (\[[0-9]+])?
// conversion format specifier: [vTtbcdoOqxXUeEfFgGsp]
// %% shouldn't be replaced.
var flagRe, _ = regexp.Compile(`%[+\-# 0]?([0-9]+|(\[[0-9]+])?\*)?(\.([0-9]+|(\[[0-9]+])?\*))?(\[[0-9]+])?[vTtbcdoOqxXUeEfFgGsp]`)

func replaceFlags(origin string) string {
	return string(flagRe.ReplaceAll([]byte(origin), []byte("{placeholder}")))
}
