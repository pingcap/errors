// Copyright 2015 PingCAP, Inc.
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
	"fmt"
	"github.com/pingcap/errors/terror"
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
	ClassExecutor  = terror.RegisterErrorClass(5, "executor")
	ClassKV        = terror.RegisterErrorClass(8, "kv")
	ClassOptimizer = terror.RegisterErrorClass(10, "planner")
	ClassParser    = terror.RegisterErrorClass(11, "parser")
	ClassServer    = terror.RegisterErrorClass(15, "server")
	ClassTable     = terror.RegisterErrorClass(19, "table")
)

const (
	CodeExecResultIsEmpty  terror.ErrCode = 3
	CodeMissConnectionID   terror.ErrCode = 1
	CodeResultUndetermined terror.ErrCode = 2
)

func TestT(t *testing.T) {
	CustomVerboseFlag = true
	TestingT(t)
}

var _ = Suite(&testTErrorSuite{})

type testTErrorSuite struct {
}

func (s *testTErrorSuite) TestErrCode(c *C) {
	c.Assert(CodeMissConnectionID, Equals, terror.ErrCode(1))
	c.Assert(CodeResultUndetermined, Equals, terror.ErrCode(2))
}

func (s *testTErrorSuite) TestTError(c *C) {
	c.Assert(ClassParser.String(), Not(Equals), "")
	c.Assert(ClassOptimizer.String(), Not(Equals), "")
	c.Assert(ClassKV.String(), Not(Equals), "")
	c.Assert(ClassServer.String(), Not(Equals), "")

	parserErr := ClassParser.New(terror.ErrCode(100), "error 100")
	c.Assert(parserErr.Error(), Not(Equals), "")
	c.Assert(ClassParser.EqualClass(parserErr), IsTrue)
	c.Assert(ClassParser.NotEqualClass(parserErr), IsFalse)

	c.Assert(ClassOptimizer.EqualClass(parserErr), IsFalse)
	optimizerErr := ClassOptimizer.New(terror.ErrCode(2), "abc")
	c.Assert(ClassOptimizer.EqualClass(errors.New("abc")), IsFalse)
	c.Assert(ClassOptimizer.EqualClass(nil), IsFalse)
	c.Assert(optimizerErr.Equal(optimizerErr.GenWithStack("def")), IsTrue)
	c.Assert(optimizerErr.Equal(nil), IsFalse)
	c.Assert(optimizerErr.Equal(errors.New("abc")), IsFalse)

	// Test case for FastGen.
	c.Assert(optimizerErr.Equal(optimizerErr.FastGen("def")), IsTrue)
	c.Assert(optimizerErr.Equal(optimizerErr.FastGen("def: %s", "def")), IsTrue)
	kvErr := ClassKV.New(1062, "key already exist")
	e := kvErr.FastGen("Duplicate entry '%d' for key 'PRIMARY'", 1)
	c.Assert(e, NotNil)
	c.Assert(e.Error(), Equals, "[kv:1062]Duplicate entry '1' for key 'PRIMARY'")
}

func (s *testTErrorSuite) TestJson(c *C) {
	prevTErr := ClassTable.New(CodeExecResultIsEmpty, "json test")
	buf, err := json.Marshal(prevTErr)
	c.Assert(err, IsNil)
	var curTErr terror.Error
	err = json.Unmarshal(buf, &curTErr)
	c.Assert(err, IsNil)
	isEqual := prevTErr.Equal(&curTErr)
	c.Assert(isEqual, IsTrue)
}

var predefinedErr = ClassExecutor.New(terror.ErrCode(123), "predefiend error")
var predefinedTextualErr = ClassExecutor.NewError(terror.ErrCode(124), "Executor is absent",
	"executor is taking vacation at %s")

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

	c.Assert(terror.ErrorEqual(e1, e2), IsTrue)
	c.Assert(terror.ErrorEqual(e1, e3), IsTrue)
	c.Assert(terror.ErrorEqual(e1, e4), IsTrue)
	c.Assert(terror.ErrorEqual(e1, e5), IsTrue)

	var e6 error

	c.Assert(terror.ErrorEqual(nil, nil), IsTrue)
	c.Assert(terror.ErrorNotEqual(e1, e6), IsTrue)
	code1 := terror.ErrCode(9001)
	code2 := terror.ErrCode(9002)
	te1 := ClassParser.Synthesize(code1, "abc")
	te3 := ClassKV.New(code1, "abc")
	te4 := ClassKV.New(code2, "abc")
	c.Assert(terror.ErrorEqual(te1, te3), IsFalse)
	c.Assert(terror.ErrorEqual(te3, te4), IsFalse)
}

func (s *testTErrorSuite) TestLog(_ *C) {
	err := fmt.Errorf("xxx")
	terror.Log(err)
}

func (s *testTErrorSuite) TestNewError(c *C) {
	today := time.Now().Weekday().String()
	err := predefinedTextualErr.GenWithStackByArgs(today)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "[executor:Executor is absent]executor is taking vacation at "+today)
}

func (s *testTErrorSuite) TestAllErrClasses(c *C) {
	items := []terror.ErrClass{
		ClassExecutor, ClassKV, ClassOptimizer, ClassParser, ClassServer, ClassTable,
	}
	registered := terror.AllErrorClasses()

	// sort it to align them.
	sort.Slice(items, func(i, j int) bool {
		return items[i] < items[j]
	})
	sort.Slice(registered, func(i, j int) bool {
		return registered[i] < registered[j]
	})

	for i := range items {
		c.Assert(items[i], Equals, registered[i])
	}
}

func (s *testTErrorSuite) TestErrorExists(c *C) {
	origin := ClassParser.NewError(114, "everything is alright", "that was a joke, hoo!")
	c.Assert(func() {
		_ = ClassParser.NewError(114, "everything is alright", "that was a joke, hoo!")
	}, Panics, "replicated error prototype created")

	// difference at either code or text should be different error
	changeCode := ClassParser.NewError(1145, "everything is alright", "that was a joke, hoo!")
	changeText := ClassParser.NewError(114, "everything goes bad", "that was a joke, hoo!")
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
