package errors

import (
	stderrors "errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unsafe"
)

type hackedStringArg struct {
	raw []byte
}

func (h hackedStringArg) FreezeStr() string {
	return string(append([]byte(nil), h.raw...))
}

func errorMatches(t *testing.T, err error, re string) {
	if err == nil && re != "" {
		t.Errorf("nil error doesn't match %s", re)
		return
	}
	match, reErr := regexp.MatchString(re, err.Error())
	if reErr != nil {
		t.Errorf("invalid regexp %s (%s)", re, reErr.Error())
		return
	}
	if !match {
		t.Errorf("error %s doesn't match %s", err.Error(), re)
		return
	}
	t.Logf("passed: %s ~= %s", err.Error(), re)
}

func TestCauseInErrorMessage(t *testing.T) {
	errTest := Normalize("this error just for testing", RFCCodeText("Internal:Test"))

	wrapped := errTest.Wrap(New("everything is alright :)"))
	errorMatches(t, wrapped, `\[Internal:Test\]this error just for testing: everything is alright :\)`)

	notWrapped := errTest.GenWithStack("everything is alright")
	errorMatches(t, notWrapped, `^\[Internal:Test\]everything is alright$`)
}

func TestWrappedNamedErrorGenWithStackByArgsFormatsCauseStack(t *testing.T) {
	errTest := Normalize("named error: %s", RFCCodeText("Internal:Test"))

	err := errTest.Wrap(New("cause error")).GenWithStackByArgs("wrapped")

	if _, ok := err.(fmt.Formatter); !ok {
		t.Fatalf("zap requires fmt.Formatter to emit errorVerbose for stackful errors, got %T", err)
	}

	formatted := fmt.Sprintf("%+v", err)
	if !strings.Contains(formatted, "github.com/pingcap/errors.TestWrappedNamedErrorGenWithStackByArgsFormatsCauseStack") {
		t.Fatalf("formatted error does not contain the wrapped cause stack:\n%s", formatted)
	}
	if !strings.Contains(formatted, "[Internal:Test]named error: wrapped") {
		t.Fatalf("formatted error does not contain named error context:\n%s", formatted)
	}
}

func TestWrappedNamedErrorFormatsStacklessCause(t *testing.T) {
	errTest := Normalize("named error: %s", RFCCodeText("Internal:Test"))

	err := errTest.Wrap(stderrors.New("plain cause")).GenWithStackByArgs("wrapped")

	formatted := fmt.Sprintf("%+v", err)
	wantPrefix := "plain cause\n[Internal:Test]named error: wrapped\n"
	if !strings.HasPrefix(formatted, wantPrefix) {
		t.Fatalf("unexpected formatted error prefix:\ngot:  %q\nwant prefix: %q", formatted, wantPrefix)
	}
	if !strings.Contains(formatted, "github.com/pingcap/errors.TestWrappedNamedErrorFormatsStacklessCause") {
		t.Fatalf("formatted error does not contain the generated stack:\n%s", formatted)
	}
}

func TestRedactFormatter(t *testing.T) {
	rv := 34.03498
	v := &redactFormatter{rv}
	for _, f := range []string{"%d", "%.2d"} {
		a := fmt.Sprintf(f, v)
		b := fmt.Sprintf("‹"+f+"›", rv)
		if a != b {
			t.Errorf("%s != %s", a, b)
		}
	}

	v = &redactFormatter{"‹"}
	if a := fmt.Sprintf("%s", v); a != "‹‹‹›" {
		t.Errorf("%s != <<<>", a)
	}
}

func TestGenWithStackByArgsNoCloneByDefault(t *testing.T) {
	errTest := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Test"))

	origin := []byte("120120519090607")
	arg := *(*string)(unsafe.Pointer(&origin))
	err := errTest.GenWithStackByArgs(arg)

	copy(origin, "1 1:1:1.0000027")
	got := err.(*withStack).error.(*Error).GetMsg()
	want := "Incorrect time value: '1 1:1:1.0000027'"
	if got != want {
		t.Fatalf("message should track source bytes by default, got %q, want %q", got, want)
	}
}

func TestGenWithStackByArgsFreezeHackedStringArg(t *testing.T) {
	errTest := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Test"))

	origin := []byte("120120519090607")
	arg := hackedStringArg{raw: origin}
	err := errTest.GenWithStackByArgs(arg)

	copy(origin, "1 1:1:1.0000027")
	got := err.(*withStack).error.(*Error).GetMsg()
	want := "Incorrect time value: '120120519090607'"
	if got != want {
		t.Fatalf("message changed after source bytes mutated, got %q, want %q", got, want)
	}
}

func TestFastGenByArgsFreezeHackedStringArg(t *testing.T) {
	errTest := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Test"))

	origin := []byte("120120519090607")
	arg := hackedStringArg{raw: origin}
	err := errTest.FastGenByArgs(arg)

	copy(origin, "1 1:1:1.0000027")
	got := err.(*withStack).error.(*Error).GetMsg()
	want := "Incorrect time value: '120120519090607'"
	if got != want {
		t.Fatalf("message changed after source bytes mutated, got %q, want %q", got, want)
	}
}

func TestGenWithStackFreezeHackedStringArg(t *testing.T) {
	errTest := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Test"))

	origin := []byte("120120519090607")
	arg := hackedStringArg{raw: origin}
	err := errTest.GenWithStack("Incorrect time value: '%s'", arg)

	copy(origin, "1 1:1:1.0000027")
	got := err.(*withStack).error.(*Error).GetMsg()
	want := "Incorrect time value: '120120519090607'"
	if got != want {
		t.Fatalf("message changed after source bytes mutated, got %q, want %q", got, want)
	}
}

func TestFastGenFreezeHackedStringArg(t *testing.T) {
	errTest := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Test"))

	origin := []byte("120120519090607")
	arg := hackedStringArg{raw: origin}
	err := errTest.FastGen("Incorrect time value: '%s'", arg)

	copy(origin, "1 1:1:1.0000027")
	got := err.(*withStack).error.(*Error).GetMsg()
	want := "Incorrect time value: '120120519090607'"
	if got != want {
		t.Fatalf("message changed after source bytes mutated, got %q, want %q", got, want)
	}
}

func TestGenWithStackByCauseFreezeHackedStringArg(t *testing.T) {
	errTest := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Test"))

	origin := []byte("120120519090607")
	arg := hackedStringArg{raw: origin}
	err := errTest.GenWithStackByCause(arg)

	copy(origin, "1 1:1:1.0000027")
	got := err.(*withStack).error.(*Error).GetMsg()
	want := "Incorrect time value: '120120519090607'"
	if got != want {
		t.Fatalf("message changed after source bytes mutated, got %q, want %q", got, want)
	}
}

func TestFastGenWithCauseFreezeHackedStringArg(t *testing.T) {
	errTest := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Test"))

	origin := []byte("120120519090607")
	arg := hackedStringArg{raw: origin}
	err := errTest.FastGenWithCause(arg)

	copy(origin, "1 1:1:1.0000027")
	got := err.(*withStack).error.(*Error).GetMsg()
	want := "Incorrect time value: '120120519090607'"
	if got != want {
		t.Fatalf("message changed after source bytes mutated, got %q, want %q", got, want)
	}
}
