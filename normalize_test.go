package errors

import (
	"fmt"
	"regexp"
	"testing"
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
