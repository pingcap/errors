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
	"strconv"
	"strings"
)

type errorRender struct {
	// Deprecated field, please use `RFCCode` instead.
	Class       int    `json:"class" toml:"-"`
	Code        int    `json:"code" toml:"-"`
	Msg         string `json:"message" toml:"error"`
	RFCCode     string `json:"rfccode" toml:"code"`
	Description string `json:"description,omitempty" toml:"description"`
	Workaround  string `json:"workaround,omitempty" toml:"workaround"`
}

// RenderTOML implements MarshaText and UnmarshaText for Error.
// Outputs text in TOML format.
func RenderTOML(e Error) interface{} {
	return errorRender{
		Msg:         e.GetMsg(),
		Description: e.description,
		Workaround:  e.workaround,
		RFCCode:     string(e.codeText),
	}
}

// MarshalJSON implements json.Marshaler interface.
// aware that this function cannot save a 'registered' status,
// since we cannot access the registry when unmarshaling,
// and the original global registry would be removed here.
// This function is reserved for compatibility.
func (e *Error) MarshalJSON() ([]byte, error) {
	ec := strings.Split(string(e.codeText), ":")[0]
	return json.Marshal(&errorRender{
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
	tErr := &errorRender{}
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
