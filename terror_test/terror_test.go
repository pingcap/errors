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

package terror_test

import (
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/pingcap/errors"
)

const (
	CodeMissConnectionID   errors.ErrCode = 1
	CodeResultUndetermined errors.ErrCode = 2
	CodeExecResultIsEmpty  errors.ErrCode = 3
)

type TErrorTestSuite struct {
	suite.Suite
}

func (s *TErrorTestSuite) TestErrCode() {
	s.Equal(CodeMissConnectionID, errors.ErrCode(1))
	s.Equal(CodeResultUndetermined, errors.ErrCode(2))
	s.Equal(CodeExecResultIsEmpty, errors.ErrCode(3))
}

var predefinedErr = errors.Normalize("predefiend error", errors.MySQLErrorCode(123))
var predefinedTextualErr = errors.Normalize("executor is taking vacation at %s", errors.RFCCodeText("executor:ExecutorAbsent"))

func example() error {
	err := call()
	return errors.Trace(err)
}

func call() error {
	return predefinedErr.GenWithStack("error message:%s", "abc")
}

func (s *TErrorTestSuite) TestJson() {
	tmpErr := errors.Normalize("this is a test error", errors.RFCCodeText("ddl:-1"), errors.MySQLErrorCode(-1))
	buf, err := json.Marshal(tmpErr)
	s.Nil(err)
	var curTErr errors.Error
	err = json.Unmarshal(buf, &curTErr)
	s.Nil(err)
	isEqual := tmpErr.Equal(&curTErr)
	s.Equal(curTErr.Error(), tmpErr.Error())
	s.True(isEqual)
}

func (s *TErrorTestSuite) TestTraceAndLocation() {
	err := example()
	stack := errors.ErrorStack(err)
	lines := strings.Split(stack, "\n")
	goroot := strings.ReplaceAll(runtime.GOROOT(), string(os.PathSeparator), "/")
	var sysStack = 0
	for _, line := range lines {
		if strings.Contains(line, goroot) {
			sysStack++
		}
	}
	s.Equalf(11, len(lines)-(2*sysStack), "stack = \n%s", stack)
	var containTerr bool
	for _, v := range lines {
		if strings.Contains(v, "terror_test.go") {
			containTerr = true
			break
		}
	}
	s.True(containTerr)
}

func (s *TErrorTestSuite) TestErrorEqual() {
	e1 := errors.New("test error")
	s.NotNil(e1)

	e2 := errors.Trace(e1)
	s.NotNil(e2)

	e3 := errors.Trace(e2)
	s.NotNil(e3)

	s.Equal(e1, errors.Cause(e2))
	s.Equal(e1, errors.Cause(e3))
	s.Equal(errors.Cause(e3), errors.Cause(e2))

	e4 := errors.New("test error")
	s.NotEqual(e1, errors.Cause(e4))

	e5 := errors.Errorf("test error")
	s.NotEqual(e1, errors.Cause(e5))

	s.True(errors.ErrorEqual(e1, e2))
	s.True(errors.ErrorEqual(e1, e3))
	s.True(errors.ErrorEqual(e1, e4))
	s.True(errors.ErrorEqual(e1, e5))

	var e6 error

	s.True(errors.ErrorEqual(nil, nil))
	s.True(errors.ErrorNotEqual(e1, e6))
}

func (s *TErrorTestSuite) TestNewError() {
	today := time.Now().Weekday().String()
	err := predefinedTextualErr.GenWithStackByArgs(today)
	s.NotNil(err)
	s.Equal("[executor:ExecutorAbsent]executor is taking vacation at "+today, err.Error())
}

func (s *TErrorTestSuite) TestRFCCode() {
	c1err1 := errors.Normalize("nothing", errors.RFCCodeText("TestErr1:Err1"))
	c2err2 := errors.Normalize("nothing", errors.RFCCodeText("TestErr2:Err2"))
	s.Equal(errors.RFCErrorCode("TestErr1:Err1"), c1err1.RFCCode())
	s.Equal(errors.RFCErrorCode("TestErr2:Err2"), c2err2.RFCCode())

	berr := errors.Normalize("nothing", errors.RFCCodeText("Blank:B1"))
	s.Equal(errors.RFCErrorCode("Blank:B1"), berr.RFCCode())
}

func (s *TErrorTestSuite) TestLineAndFile() {
	err := predefinedTextualErr.GenWithStackByArgs("everyday")
	_, f, l, _ := runtime.Caller(0)
	terr, ok := errors.Cause(err).(*errors.Error)
	s.True(ok)

	file, line := terr.Location()
	s.Equal(f, file)
	s.Equal(l-1, line)

	err2 := predefinedTextualErr.GenWithStackByArgs("everyday and everywhere")
	_, f2, l2, _ := runtime.Caller(0)
	terr2, ok2 := errors.Cause(err2).(*errors.Error)
	s.True(ok2)
	file2, line2 := terr2.Location()
	s.Equal(f2, file2)
	s.Equal(l2-1, line2)
}

func (s *TErrorTestSuite) TestWarpAndField() {
	cause := errors.New("load from etcd meet error")
	s.NotNil(cause)

	err := errors.Normalize("fail to get leader", errors.RFCCodeText("member:ErrGetLeader"))
	errWithCause := errors.Annotate(err, cause.Error())
	s.NotNil(errWithCause)

	s.Equal("load from etcd meet error: [member:ErrGetLeader]fail to get leader", errWithCause.Error())
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(TErrorTestSuite))
}
