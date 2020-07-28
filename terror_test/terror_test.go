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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	. "github.com/pingcap/check"
	"github.com/pingcap/errors"
)

// Error classes.
// Those fields below are copied from the original version of terror,
// so that we can reuse those test cases.
var (
	reg            = errors.NewRegistry("DB")
	ClassExecutor  = reg.RegisterErrorClass(5, "executor")
	ClassKV        = reg.RegisterErrorClass(8, "kv")
	ClassOptimizer = reg.RegisterErrorClass(10, "planner")
	ClassParser    = reg.RegisterErrorClass(11, "parser")
	ClassServer    = reg.RegisterErrorClass(15, "server")
	ClassTable     = reg.RegisterErrorClass(19, "table")
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

func (s *testTErrorSuite) TestTError(c *C) {
	c.Assert(ClassParser.String(), Not(Equals), "")
	c.Assert(ClassOptimizer.String(), Not(Equals), "")
	c.Assert(ClassKV.String(), Not(Equals), "")
	c.Assert(ClassServer.String(), Not(Equals), "")

	parserErr := ClassParser.New(errors.ErrCode(100), "error 100")
	c.Assert(parserErr.Error(), Not(Equals), "")
	c.Assert(ClassParser.EqualClass(parserErr), IsTrue)
	c.Assert(ClassParser.NotEqualClass(parserErr), IsFalse)

	c.Assert(ClassOptimizer.EqualClass(parserErr), IsFalse)
	optimizerErr := ClassOptimizer.New(errors.ErrCode(2), "abc")
	c.Assert(ClassOptimizer.EqualClass(errors.New("abc")), IsFalse)
	c.Assert(ClassOptimizer.EqualClass(nil), IsFalse)
	c.Assert(optimizerErr.Equal(optimizerErr.GenWithStack("def")), IsTrue)
	c.Assert(optimizerErr.Equal(errors.Trace(optimizerErr.GenWithStack("def"))), IsTrue)
	c.Assert(optimizerErr.Equal(nil), IsFalse)
	c.Assert(optimizerErr.Equal(errors.New("abc")), IsFalse)

	// Test case for FastGen.
	c.Assert(optimizerErr.Equal(optimizerErr.FastGen("def")), IsTrue)
	c.Assert(optimizerErr.Equal(optimizerErr.FastGen("def: %s", "def")), IsTrue)
	kvErr := ClassKV.New(1062, "key already exist")
	e := kvErr.FastGen("Duplicate entry '%d' for key 'PRIMARY'", 1)
	c.Assert(e, NotNil)
	c.Assert(e.Error(), Equals, "[DB:kv:1062] Duplicate entry '1' for key 'PRIMARY'")
}

func (s *testTErrorSuite) TestJson(c *C) {
	prevTErr := ClassTable.New(CodeExecResultIsEmpty, "json test")
	buf, err := json.Marshal(prevTErr)
	c.Assert(err, IsNil)
	var curTErr errors.Error
	err = json.Unmarshal(buf, &curTErr)
	c.Assert(err, IsNil)
	isEqual := prevTErr.Equal(&curTErr)
	c.Assert(isEqual, IsTrue)
}

var predefinedErr = ClassExecutor.New(errors.ErrCode(123), "predefiend error")
var predefinedTextualErr = ClassExecutor.DefineError().
	TextualCode("ExecutorAbsent").
	MessageTemplate("executor is taking vacation at %s").
	Build()

func example() error {
	err := call()
	return errors.Trace(err)
}

func call() error {
	return predefinedErr.GenWithStack("error message:%s", "abc")
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
	code1 := errors.ErrCode(9001)
	code2 := errors.ErrCode(9002)
	te1 := ClassParser.Synthesize(code1, "abc")
	te3 := ClassKV.New(code1, "abc")
	te4 := ClassKV.New(code2, "abc")
	c.Assert(errors.ErrorEqual(te1, te3), IsFalse)
	c.Assert(errors.ErrorEqual(te3, te4), IsFalse)
}

func (s *testTErrorSuite) TestNewError(c *C) {
	today := time.Now().Weekday().String()
	err := predefinedTextualErr.GenWithStackByArgs(today)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "[DB:executor:ExecutorAbsent] executor is taking vacation at "+today)
}

func (s *testTErrorSuite) TestAllErrClasses(c *C) {
	items := []errors.ErrClass{
		ClassExecutor, ClassKV, ClassOptimizer, ClassParser, ClassServer, ClassTable,
	}
	registered := reg.AllErrorClasses()

	// sort it to align them.
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	sort.Slice(registered, func(i, j int) bool {
		return registered[i].ID < registered[j].ID
	})

	for i := range items {
		c.Assert(items[i].ID, Equals, registered[i].ID)
	}
}

func (s *testTErrorSuite) TestErrorExists(c *C) {
	origin := ClassParser.DefineError().
		TextualCode("EverythingAlright").
		MessageTemplate("that was a joke, hoo!").
		Build()

	c.Assert(func() {
		_ = ClassParser.DefineError().
			TextualCode("EverythingAlright").
			MessageTemplate("that was another joke, hoo!").
			Build()
	}, Panics, "replicated error prototype created")

	// difference at either code or text should be different error
	changeCode := ClassParser.DefineError().
		NumericCode(4399).
		MessageTemplate("that was a joke, hoo!").
		Build()
	changeText := ClassParser.DefineError().
		TextualCode("EverythingBad").
		MessageTemplate("that was not a joke, folks!").
		Build()
	containsErr := func(e error) bool {
		for _, err := range ClassParser.AllErrors() {
			if err.Equal(e) {
				return true
			}
		}
		return false
	}
	c.Assert(containsErr(origin), IsTrue)
	c.Assert(containsErr(changeCode), IsTrue)
	c.Assert(containsErr(changeText), IsTrue)
}

func (s *testTErrorSuite) TestRFCCode(c *C) {
	reg := errors.NewRegistry("TEST")
	errc1 := reg.RegisterErrorClass(1, "TestErr1")
	errc2 := reg.RegisterErrorClass(2, "TestErr2")
	c1err1 := errc1.DefineError().
		TextualCode("Err1").
		MessageTemplate("nothing").
		Build()
	c2err2 := errc2.DefineError().
		TextualCode("Err2").
		MessageTemplate("nothing").
		Build()
	c.Assert(c1err1.RFCCode(), Equals, errors.RFCErrorCode("TEST:TestErr1:Err1"))
	c.Assert(c2err2.RFCCode(), Equals, errors.RFCErrorCode("TEST:TestErr2:Err2"))
	blankReg := errors.NewRegistry("")
	errb := blankReg.RegisterErrorClass(1, "Blank")
	berr := errb.DefineError().
		TextualCode("B1").
		MessageTemplate("nothing").
		Workaround(`Do nothing`).
		Build()
	c.Assert(berr.RFCCode(), Equals, errors.RFCErrorCode("Blank:B1"))
}

const (
	somewhatErrorTOML = `[error.KV:Somewhat:Foo]
error = '''some {placeholder} thing happened, and some {placeholder} goes verbose. I'm {placeholder} percent confusing...
Maybe only {placeholder} peaces of placeholders can save me... Oh my {placeholder}.{placeholder}!'''
description = '''N/A'''
workaround = '''N/A'''
`
	err8005TOML = `[error.KV:2PC:8005]
error = '''Write Conflict, txnStartTS is stale'''
description = '''A certain Raft Group is not available, such as the number of replicas is not enough.
This error usually occurs when the TiKV server is busy or the TiKV node is down.'''
` + "workaround = '''Check whether `tidb_disable_txn_auto_retry` is set to `on`. If so, set it to `off`; " +
		"if it is already `off`, increase the value of `tidb_retry_limit` until the error no longer occurs.'''\n"
	errUnavailableTOML = `[error.KV:Region:Unavailable]
error = '''Region is unavailable'''
description = '''A certain Raft Group is not available, such as the number of replicas is not enough.
This error usually occurs when the TiKV server is busy or the TiKV node is down.'''
workaround = '''Check the status, monitoring data and log of the TiKV server.'''
`
)

func (*testTErrorSuite) TestExport(c *C) {
	RegKV := errors.NewRegistry("KV")
	Class2PC := RegKV.RegisterErrorClass(1, "2PC")
	_ = Class2PC.DefineError().
		NumericCode(8005).
		Description("A certain Raft Group is not available, such as the number of replicas is not enough.\n" +
			"This error usually occurs when the TiKV server is busy or the TiKV node is down.").
		Workaround("Check whether `tidb_disable_txn_auto_retry` is set to `on`. If so, set it to `off`; " +
			"if it is already `off`, increase the value of `tidb_retry_limit` until the error no longer occurs.").
		MessageTemplate("Write Conflict, txnStartTS is stale").
		Build()

	ClassRegion := RegKV.RegisterErrorClass(2, "Region")
	_ = ClassRegion.DefineError().
		TextualCode("Unavailable").
		Description("A certain Raft Group is not available, such as the number of replicas is not enough.\n" +
			"This error usually occurs when the TiKV server is busy or the TiKV node is down.").
		Workaround("Check the status, monitoring data and log of the TiKV server.").
		MessageTemplate("Region is unavailable").
		Build()

	ClassSomewhat := RegKV.RegisterErrorClass(3, "Somewhat")
	_ = ClassSomewhat.DefineError().
		TextualCode("Foo").
		MessageTemplate("some %.6s thing happened, and some %#v goes verbose. I'm %6.3f percent confusing...\n" +
			"Maybe only %[3]*.[2]*[1]f peaces of placeholders can save me... Oh my %s.%d!").
		Build()

	result := bytes.NewBuffer([]byte{})
	err := RegKV.ExportTo(result)
	c.Assert(err, IsNil)
	resultStr := result.String()
	fmt.Println("Result: ")
	fmt.Print(resultStr)
	c.Assert(strings.Contains(resultStr, somewhatErrorTOML), IsTrue)
	c.Assert(strings.Contains(resultStr, err8005TOML), IsTrue)
	c.Assert(strings.Contains(resultStr, errUnavailableTOML), IsTrue)
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
