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
)

const tomlTemplate = `[error.%s]
error = '''%s'''
description = '''%s'''
workaround = '''%s'''

`

func (e *Error) exportTo(writer io.Writer) error {
	if e.Description == "" {
		log.Warn("error description missed", zap.String("error", e.RFCCode()))
		e.Description = "N/A"
	}
	if e.Workaround == "" {
		log.Warn("error workaround missed", zap.String("error", e.RFCCode()))
		e.Workaround = "N/A"
	}
	_, err := fmt.Fprintf(writer, tomlTemplate, e.RFCCode(), e.MessageTemplate(), e.Description, e.Workaround)
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
