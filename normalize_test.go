package errors

import (
	"fmt"
	"regexp"
	"testing"
	"unsafe"
)

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

func TestGenWithStackByArgsFreezeStringArg(t *testing.T) {
	errTest := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Test"))

	origin := []byte("120120519090607")
	arg := *(*string)(unsafe.Pointer(&origin))
	err := errTest.GenWithStackByArgs(arg)

	copy(origin, "1 1:1:1.0000027")
	got := err.(*withStack).error.(*Error).GetMsg()
	want := "Incorrect time value: '120120519090607'"
	if got != want {
		t.Fatalf("message changed after source bytes mutated, got %q, want %q", got, want)
	}
}

func TestFastGenByArgsFreezeStringArg(t *testing.T) {
	errTest := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Test"))

	origin := []byte("120120519090607")
	arg := *(*string)(unsafe.Pointer(&origin))
	err := errTest.FastGenByArgs(arg)

	copy(origin, "1 1:1:1.0000027")
	got := err.(*withStack).error.(*Error).GetMsg()
	want := "Incorrect time value: '120120519090607'"
	if got != want {
		t.Fatalf("message changed after source bytes mutated, got %q, want %q", got, want)
	}
}
