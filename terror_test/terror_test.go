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

	. "github.com/pingcap/check"
	"github.com/pingcap/errors"
)

const (
	CodeExecResultIsEmpty  errors.ErrCode = 3
	CodeMissConnectionID   errors.ErrCode = 1
	CodeResultUndetermined errors.ErrCode = 2
)

func TestT(t *testing.T) {
	CustomVerboseFlag = true
	TestingT(t)
}

var _ = Suite(&testTErrorSuite{})

type testTErrorSuite struct {
}

func (s *testTErrorSuite) TestErrCode(c *C) {
	c.Assert(CodeMissConnectionID, Equals, errors.ErrCode(1))
	c.Assert(CodeResultUndetermined, Equals, errors.ErrCode(2))
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

func (s *testTErrorSuite) TestJson(c *C) {
	tmpErr := errors.Normalize("this is a test error", errors.RFCCodeText("ddl:-1"), errors.MySQLErrorCode(-1))
	buf, err := json.Marshal(tmpErr)
	c.Assert(err, IsNil)
	var curTErr errors.Error
	err = json.Unmarshal(buf, &curTErr)
	c.Assert(err, IsNil)
	isEqual := tmpErr.Equal(&curTErr)
	c.Assert(curTErr.Error(), Equals, tmpErr.Error())
	c.Assert(isEqual, IsTrue)
}

func (s *testTErrorSuite) TestTraceAndLocation(c *C) {
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
	c.Assert(len(lines)-(2*sysStack), Equals, 15, Commentf("stack =\n%s", stack))
	var containTerr bool
	for _, v := range lines {
		if strings.Contains(v, "terror_test.go") {
			containTerr = true
			break
		}
	}
	c.Assert(containTerr, IsTrue)
}

func (s *testTErrorSuite) TestErrorEqual(c *C) {
	e1 := errors.New("test error")
	c.Assert(e1, NotNil)

	e2 := errors.Trace(e1)
	c.Assert(e2, NotNil)

	e3 := errors.Trace(e2)
	c.Assert(e3, NotNil)

	c.Assert(errors.Cause(e2), Equals, e1)
	c.Assert(errors.Cause(e3), Equals, e1)
	c.Assert(errors.Cause(e2), Equals, errors.Cause(e3))

	e4 := errors.New("test error")
	c.Assert(errors.Cause(e4), Not(Equals), e1)

	e5 := errors.Errorf("test error")
	c.Assert(errors.Cause(e5), Not(Equals), e1)

	c.Assert(errors.ErrorEqual(e1, e2), IsTrue)
	c.Assert(errors.ErrorEqual(e1, e3), IsTrue)
	c.Assert(errors.ErrorEqual(e1, e4), IsTrue)
	c.Assert(errors.ErrorEqual(e1, e5), IsTrue)

	var e6 error

	c.Assert(errors.ErrorEqual(nil, nil), IsTrue)
	c.Assert(errors.ErrorNotEqual(e1, e6), IsTrue)
}

func (s *testTErrorSuite) TestNewError(c *C) {
	today := time.Now().Weekday().String()
	err := predefinedTextualErr.GenWithStackByArgs(today)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "[executor:ExecutorAbsent]executor is taking vacation at "+today)
}

func (s *testTErrorSuite) TestRFCCode(c *C) {
	c1err1 := errors.Normalize("nothing", errors.RFCCodeText("TestErr1:Err1"))
	c2err2 := errors.Normalize("nothing", errors.RFCCodeText("TestErr2:Err2"))
	c.Assert(c1err1.RFCCode(), Equals, errors.RFCErrorCode("TestErr1:Err1"))
	c.Assert(c2err2.RFCCode(), Equals, errors.RFCErrorCode("TestErr2:Err2"))
	berr := errors.Normalize("nothing", errors.RFCCodeText("Blank:B1"), errors.Workaround(`Do nothing`))
	c.Assert(berr.RFCCode(), Equals, errors.RFCErrorCode("Blank:B1"))
}

func (*testTErrorSuite) TestLineAndFile(c *C) {
	err := predefinedTextualErr.GenWithStackByArgs("everyday")
	_, f, l, _ := runtime.Caller(0)
	terr, ok := errors.Cause(err).(*errors.Error)
	c.Assert(ok, IsTrue)
	file, line := terr.Location()
	c.Assert(file, Equals, f)
	c.Assert(line, Equals, l-1)

	err2 := predefinedTextualErr.GenWithStackByArgs("everyday and everywhere")
	_, f2, l2, _ := runtime.Caller(0)
	terr2, ok2 := errors.Cause(err2).(*errors.Error)
	c.Assert(ok2, IsTrue)
	file2, line2 := terr2.Location()
	c.Assert(file2, Equals, f2)
	c.Assert(line2, Equals, l2-1)
}

func (*testTErrorSuite) TestWarpAndField(c *C) {
	causeErr := errors.New("load from etcd meet error")
	ErrGetLeader := errors.Normalize("fail to get leader", errors.RFCCodeText("member:ErrGetLeader"))
	errWithWarpedCause := errors.Annotate(ErrGetLeader, causeErr.Error())
	c.Assert(errWithWarpedCause.Error(), Equals, "load from etcd meet error: [member:ErrGetLeader]fail to get leader")
}
