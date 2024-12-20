package errors

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		err  string
		want error
	}{
		{"", fmt.Errorf("")},
		{"foo", fmt.Errorf("foo")},
		{"foo", New("foo")},
		{"string with format specifiers: %v", errors.New("string with format specifiers: %v")},
	}

	for _, tt := range tests {
		got := New(tt.err)
		if got.Error() != tt.want.Error() {
			t.Errorf("New.Error(): got: %q, want %q", got, tt.want)
		}
	}
}

func TestWrapNil(t *testing.T) {
	got := Annotate(nil, "no error")
	if got != nil {
		t.Errorf("Wrap(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{Annotate(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := Annotate(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("Wrap(%v, %q): got: %v, want %v", tt.err, tt.message, got, tt.want)
		}
	}
}

type nilError struct{}

func (nilError) Error() string { return "nil error" }

func TestCause(t *testing.T) {
	x := New("error")
	tests := []struct {
		err  error
		want error
	}{{
		// nil error is nil
		err:  nil,
		want: nil,
	}, {
		// explicit nil error is nil
		err:  (error)(nil),
		want: nil,
	}, {
		// typed nil is nil
		err:  (*nilError)(nil),
		want: (*nilError)(nil),
	}, {
		// uncaused error is unaffected
		err:  io.EOF,
		want: io.EOF,
	}, {
		// caused error returns cause
		err:  Annotate(io.EOF, "ignored"),
		want: io.EOF,
	}, {
		err:  x, // return from errors.New
		want: x,
	}, {
		WithMessage(nil, "whoops"),
		nil,
	}, {
		WithMessage(io.EOF, "whoops"),
		io.EOF,
	}, {
		WithStack(nil),
		nil,
	}, {
		WithStack(io.EOF),
		io.EOF,
	}, {
		AddStack(nil),
		nil,
	}, {
		AddStack(io.EOF),
		io.EOF,
	}}

	for i, tt := range tests {
		got := Cause(tt.err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("test %d: got %#v, want %#v", i+1, got, tt.want)
		}
	}
}

func TestWrapfNil(t *testing.T) {
	got := Annotate(nil, "no error")
	if got != nil {
		t.Errorf("Wrapf(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrapf(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{Annotatef(io.EOF, "read error without format specifiers"), "client error", "client error: read error without format specifiers: EOF"},
		{Annotatef(io.EOF, "read error with %d format specifier", 1), "client error", "client error: read error with 1 format specifier: EOF"},
	}

	for _, tt := range tests {
		got := Annotatef(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("Wrapf(%v, %q): got: %v, want %v", tt.err, tt.message, got, tt.want)
		}
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{Errorf("read error without format specifiers"), "read error without format specifiers"},
		{Errorf("read error with %d format specifier", 1), "read error with 1 format specifier"},
	}

	for _, tt := range tests {
		got := tt.err.Error()
		if got != tt.want {
			t.Errorf("Errorf(%v): got: %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestWithStackNil(t *testing.T) {
	got := WithStack(nil)
	if got != nil {
		t.Errorf("WithStack(nil): got %#v, expected nil", got)
	}
	got = AddStack(nil)
	if got != nil {
		t.Errorf("AddStack(nil): got %#v, expected nil", got)
	}
}

func TestWithStack(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{io.EOF, "EOF"},
		{WithStack(io.EOF), "EOF"},
	}

	for _, tt := range tests {
		got := WithStack(tt.err).Error()
		if got != tt.want {
			t.Errorf("WithStack(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestAddStack(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{io.EOF, "EOF"},
		{AddStack(io.EOF), "EOF"},
	}

	for _, tt := range tests {
		got := AddStack(tt.err).Error()
		if got != tt.want {
			t.Errorf("AddStack(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestGetStackTracer(t *testing.T) {
	orig := io.EOF
	if GetStackTracer(orig) != nil {
		t.Errorf("GetStackTracer: got: %v, want %v", GetStackTracer(orig), nil)
	}
	stacked := AddStack(orig)
	if GetStackTracer(stacked).(error) != stacked {
		t.Errorf("GetStackTracer(stacked): got: %v, want %v", GetStackTracer(stacked), stacked)
	}
	final := AddStack(stacked)
	if GetStackTracer(final).(error) != stacked {
		t.Errorf("GetStackTracer(final): got: %v, want %v", GetStackTracer(final), stacked)
	}
}

func TestAddStackDedup(t *testing.T) {
	stacked := WithStack(io.EOF)
	err := AddStack(stacked)
	if err != stacked {
		t.Errorf("AddStack: got: %+v, want %+v", err, stacked)
	}
	err = WithStack(stacked)
	if err == stacked {
		t.Errorf("WithStack: got: %v, don't want %v", err, stacked)
	}
}

func TestWithMessageNil(t *testing.T) {
	got := WithMessage(nil, "no error")
	if got != nil {
		t.Errorf("WithMessage(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWithMessage(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{WithMessage(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := WithMessage(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("WithMessage(%v, %q): got: %q, want %q", tt.err, tt.message, got, tt.want)
		}
	}
}

// errors.New, etc values are not expected to be compared by value
// but the change in errors#27 made them incomparable. Assert that
// various kinds of errors have a functional equality operator, even
// if the result of that equality is always false.
func TestErrorEquality(t *testing.T) {
	vals := []error{
		nil,
		io.EOF,
		errors.New("EOF"),
		New("EOF"),
		Errorf("EOF"),
		Annotate(io.EOF, "EOF"),
		Annotatef(io.EOF, "EOF%d", 2),
		WithMessage(nil, "whoops"),
		WithMessage(io.EOF, "whoops"),
		WithStack(io.EOF),
		WithStack(nil),
		AddStack(io.EOF),
		AddStack(nil),
	}

	for i := range vals {
		for j := range vals {
			_ = vals[i] == vals[j] // mustn't panic
		}
	}
}

func TestFind(t *testing.T) {
	eNew := errors.New("error")
	wrapped := Annotate(nilError{}, "nil")
	tests := []struct {
		err    error
		finder func(error) bool
		found  error
	}{
		{io.EOF, func(_ error) bool { return true }, io.EOF},
		{io.EOF, func(_ error) bool { return false }, nil},
		{io.EOF, func(err error) bool { return err == io.EOF }, io.EOF},
		{io.EOF, func(err error) bool { return err != io.EOF }, nil},

		{eNew, func(err error) bool { return true }, eNew},
		{eNew, func(err error) bool { return false }, nil},

		{nilError{}, func(err error) bool { return true }, nilError{}},
		{nilError{}, func(err error) bool { return false }, nil},
		{nilError{}, func(err error) bool { _, ok := err.(nilError); return ok }, nilError{}},

		{wrapped, func(err error) bool { return true }, wrapped},
		{wrapped, func(err error) bool { return false }, nil},
		{wrapped, func(err error) bool { _, ok := err.(nilError); return ok }, nilError{}},
	}

	for _, tt := range tests {
		got := Find(tt.err, tt.finder)
		if got != tt.found {
			t.Errorf("WithMessage(%v): got: %q, want %q", tt.err, got, tt.found)
		}
	}
}

type errWalkTest struct {
	cause error
	sub   []error
	v     int
}

func (e *errWalkTest) Error() string {
	return strconv.Itoa(e.v)
}

func (e *errWalkTest) Cause() error {
	return e.cause
}

func (e *errWalkTest) Errors() []error {
	return e.sub
}

func testFind(err error, v int) bool {
	return WalkDeep(err, func(err error) bool {
		e := err.(*errWalkTest)
		return e.v == v
	})
}

func TestWalkDeep(t *testing.T) {
	err := &errWalkTest{
		sub: []error{
			&errWalkTest{
				v:     10,
				cause: &errWalkTest{v: 11},
			},
			&errWalkTest{
				v:     20,
				cause: &errWalkTest{v: 21, cause: &errWalkTest{v: 22}},
			},
			&errWalkTest{
				v:     30,
				cause: &errWalkTest{v: 31},
			},
		},
	}

	if !testFind(err, 11) {
		t.Errorf("not found in first cause chain")
	}

	if !testFind(err, 22) {
		t.Errorf("not found in siblings")
	}

	if testFind(err, 32) {
		t.Errorf("found not exists")
	}
}

func TestWalkDeepNil(t *testing.T) {
	require.False(t, WalkDeep(nil, func(err error) bool { return true }))
}

func TestWalkDeepComplexTree(t *testing.T) {
	err := &errWalkTest{v: 1, cause: &errWalkTest{
		sub: []error{
			&errWalkTest{
				v:     10,
				cause: &errWalkTest{v: 11},
			},
			&errWalkTest{
				v: 20,
				sub: []error{
					&errWalkTest{v: 21},
					&errWalkTest{v: 22},
				},
			},
			&errWalkTest{
				v:     30,
				cause: &errWalkTest{v: 31},
			},
		},
	}}

	assertFind := func(v int, comment string) {
		if !testFind(err, v) {
			t.Errorf("%d not found in the error: %s", v, comment)
		}
	}
	assertNotFind := func(v int, comment string) {
		if testFind(err, v) {
			t.Errorf("%d found in the error, but not expected: %s", v, comment)
		}
	}

	assertFind(1, "shallow search")
	assertFind(11, "deep search A1")
	assertFind(21, "deep search A2")
	assertFind(22, "deep search B1")
	assertNotFind(23, "deep search Neg")
	assertFind(31, "deep search B2")
	assertNotFind(32, "deep search Neg")
	assertFind(30, "Tree node A")
	assertFind(20, "Tree node with many children")
}

type fooError int

func (fooError) Error() string {
	return "foo"
}

func TestWorkWithStdErrors(t *testing.T) {
	e1 := fooError(100)
	e2 := Normalize("e2", RFCCodeText("e2"))
	e3 := Normalize("e3", RFCCodeText("e3"))
	e21 := e2.Wrap(e1)
	e31 := e3.Wrap(e1)
	e32 := e3.Wrap(e2)
	e321 := e3.Wrap(e21)

	unwrapTbl := []struct {
		x *Error // x.Unwrap() == y
		y error
	}{{e2, nil}, {e3, nil}, {e21, e1}, {e31, e1}, {e32, e2}, {e321, e21}}
	for _, c := range unwrapTbl {
		if c.x.Unwrap() != c.y {
			t.Errorf("`%s`.Unwrap() != `%s`", c.x, c.y)
		}
	}

	isTbl := []struct {
		x, y error // errors.Is(x, y) == b
		b    bool
	}{
		{e1, e1, true}, {e2, e1, false}, {e3, e1, false}, {e21, e1, true}, {e321, e1, true},
		{e1, e2, false}, {e2, e2, true}, {e3, e2, false}, {e21, e2, true}, {e31, e2, false}, {e321, e2, true},
		{e2, e21, true}, {e21, e21, true}, {e31, e21, false}, {e321, e21, true},
		{e321, e321, true}, {e3, e321, true}, {e21, e321, false},
	}
	for _, c := range isTbl {
		if c.b && !errors.Is(c.x, c.y) {
			t.Errorf("`%s` is not `%s`", c.x, c.y)
		}
		if !c.b && errors.Is(c.x, c.y) {
			t.Errorf("`%s` is `%s`", c.x, c.y)
		}
	}

	var e1x fooError
	if ok := errors.As(e21, &e1x); !ok {
		t.Error("e21 cannot convert to e1")
	}
	if int(e1x) != 100 {
		t.Error("e1x is not 100")
	}

	var e2x *Error
	if ok := errors.As(e21, &e2x); !ok {
		t.Error("e21 cannot convert to e2")
	}
	if e2x.ID() != "e2" {
		t.Error("err is not e2")
	}

	e3x := e3.Wrap(e1)
	if ok := errors.As(e21, &e3x); !ok {
		t.Error("e21 cannot convert to e3")
	}
	if e3x.ID() != "e2" {
		t.Error("err is not e2")
	}
}

func TestHasTrace(t *testing.T) {
	targetErr := Normalize("test err")
	require.False(t, HasStack(targetErr))
	require.False(t, HasStack(targetErr.FastGen("fast gen")))
	require.False(t, HasStack(targetErr.FastGenByArgs("fast gen arg")))
	require.True(t, HasStack(Trace(targetErr.FastGen("fast gen"))))
	require.True(t, HasStack(targetErr.GenWithStack("gen")))
}
