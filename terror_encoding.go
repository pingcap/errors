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
	"bytes"
	"encoding"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type errorRender struct {
	// Deprecated field, please use `RFCCode` instead.
	Class       int    `json:"class" toml:"-"`
	Code        int    `json:"code" toml:"-"`
	Msg         string `json:"message" toml:"message"`
	RFCCode     string `json:"rfccode" toml:"code"`
	Description string `json:"description,omitempty" toml:"description"`
	Workaround  string `json:"workaround,omitempty" toml:"workaround"`
}

// RenderTOML implements MarshaText and UnmarshaText for Error.
// Outputs text in TOML format.
type RenderTOML Error

var _ encoding.TextMarshaler = RenderTOML{}
var _ encoding.TextUnmarshaler = &RenderTOML{}

// MarshalText implements encoding.TextMarshaler interface.
func (rd RenderTOML) MarshalText() ([]byte, error) {
	e := Error(rd)
	render := &errorRender{
		Msg:         e.GetMsg(),
		Description: e.description,
		Workaround:  e.workaround,
		RFCCode:     string(e.codeText),
	}
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(render); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (rd *RenderTOML) UnmarshalText(data []byte) error {
	e := (*Error)(rd)
	render := &errorRender{}

	if _, err := toml.Decode(string(data), &render); err != nil {
		return Trace(err)
	}
	codes := strings.Split(string(render.RFCCode), ":")
	innerCode := codes[len(codes)-1]
	if i, errAtoi := strconv.Atoi(innerCode); errAtoi == nil {
		e.code = ErrCode(i)
	}
	e.codeText = ErrCodeText(render.RFCCode)
	e.message = render.Msg
	e.workaround = render.Workaround
	e.description = render.Description
	return nil
}

// RenderJSON implements MarshaJSON and UnmarshaJSON for Error.
type RenderJSON Error

var _ json.Marshaler = RenderJSON{}
var _ json.Unmarshaler = &RenderJSON{}

// MarshalJSON implements json.Marshaler interface.
func (rd RenderJSON) MarshalJSON() ([]byte, error) {
	e := Error(rd)
	ec := strings.Split(string(e.codeText), ":")[0]
	return json.Marshal(&errorRender{
		Class:       rfcCode2class[ec],
		Code:        int(e.code),
		Msg:         e.GetMsg(),
		RFCCode:     string(e.codeText),
		Description: e.description,
		Workaround:  e.workaround,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (rd *RenderJSON) UnmarshalJSON(data []byte) error {
	e := (*Error)(rd)
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
	e.description = tErr.Description
	e.workaround = tErr.Workaround
	return nil
}

// MarshalJSON implements json.Marshaler interface.
// aware that this function cannot save a 'registered' status,
// since we cannot access the registry when unmarshaling,
// and the original global registry would be removed here.
// This function is reserved for compatibility.
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal((*RenderJSON)(e))
}

// UnmarshalJSON implements json.Unmarshaler interface.
// aware that this function cannot create a 'registered' error,
// since we cannot access the registry in this context,
// and the original global registry is removed.
// This function is reserved for compatibility.
func (e *Error) UnmarshalJSON(data []byte) error {
	rd := (*RenderJSON)(e)
	return rd.UnmarshalJSON(data)
}
