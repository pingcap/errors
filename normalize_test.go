package errors

import (
	stderrors "errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
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

func TestRedactedArgFormatter(t *testing.T) {
	previousMode := RedactLogEnabled.Load()
	t.Cleanup(func() { RedactLogEnabled.Store(previousMode) })
	RedactLogEnabled.Store(RedactLogEnable)
	args := []any{42}
	RedactErrorArg(args, []int{0})

	for _, format := range []string{
		"%v", "%+v", "%#v", "%s", "%q", "%d", "%04d", "%+8.4d",
		"%b", "%o", "%O", "%x", "%X", "%c", "%U", "%t",
		"%e", "%E", "%f", "%F", "%g", "%G", "%.0s", "%z",
	} {
		t.Run(format, func(t *testing.T) {
			require.Equal(t, "?", fmt.Sprintf(format, args...))
		})
	}
}

func TestGenWithStackRedaction(t *testing.T) {
	previousMode := RedactLogEnabled.Load()
	t.Cleanup(func() { RedactLogEnabled.Store(previousMode) })
	prototype := Normalize("unused template", RFCCodeText("Internal:Test"), RedactArgs([]int{0, 5}))

	for _, tc := range []struct {
		mode string
		want string
	}{
		{RedactLogDisable, "value secret‹›, count 0042, public visible"},
		{RedactLogEnable, "value ?, count 0042, public visible"},
		{RedactLogMarker, "value ‹secret‹‹›››, count 0042, public visible"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			RedactLogEnabled.Store(tc.mode)
			origin := []byte("secret‹›")
			err := prototype.GenWithStack("value %s, count %04d, public %s", hackedStringArg{raw: origin}, 42, "visible")
			copy(origin, "modified")
			got := err.(*withStack).error.(*Error)
			require.Equal(t, tc.want, got.GetMsg())
			require.Equal(t, prototype.ID(), got.ID())
			require.True(t, HasStack(err))
			file, line := got.Location()
			require.True(t, strings.HasSuffix(file, "normalize_test.go"))
			require.Positive(t, line)
			require.Equal(t, "unused template", prototype.message)
			require.Empty(t, prototype.args)
		})
	}
}

func TestFastGenRedaction(t *testing.T) {
	previousMode := RedactLogEnabled.Load()
	t.Cleanup(func() { RedactLogEnabled.Store(previousMode) })
	prototype := Normalize("unused template", RFCCodeText("Internal:Test"), RedactArgs([]int{0, 1, 5}))
	const format = "value %s, count %04d, public %s"

	for _, tc := range []struct {
		mode string
		want string
	}{
		{RedactLogDisable, "value secret‹›, count 0042, public visible"},
		{RedactLogEnable, "value ?, count ?, public visible"},
		{RedactLogMarker, "value ‹secret‹‹›››, count ‹0042›, public visible"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			RedactLogEnabled.Store(tc.mode)
			origin := []byte("secret‹›")
			err := prototype.FastGen(format, hackedStringArg{raw: origin}, 42, "visible")
			copy(origin, "modified")
			got := err.(*withStack).error.(*Error)
			require.Equal(t, tc.want, got.GetMsg())
			require.Equal(t, prototype.ID(), got.ID())
			require.False(t, HasStack(err))
			file, line := got.Location()
			require.Empty(t, file)
			require.Zero(t, line)
			require.Equal(t, "unused template", prototype.message)
			require.Empty(t, prototype.args)
		})
	}
}

func TestGenWithStackByArgsRedaction(t *testing.T) {
	previousMode := RedactLogEnabled.Load()
	t.Cleanup(func() { RedactLogEnabled.Store(previousMode) })
	prototype := Normalize("value %s, count %04d, public %s", RFCCodeText("Internal:Test"), RedactArgs([]int{0, 5}))

	for _, tc := range []struct {
		mode string
		want string
	}{
		{RedactLogDisable, "value secret‹›, count 0042, public visible"},
		{RedactLogEnable, "value ?, count 0042, public visible"},
		{RedactLogMarker, "value ‹secret‹‹›››, count 0042, public visible"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			RedactLogEnabled.Store(tc.mode)
			origin := []byte("secret‹›")
			err := prototype.GenWithStackByArgs(hackedStringArg{raw: origin}, 42, "visible")
			copy(origin, "modified")
			got := err.(*withStack).error.(*Error)
			require.Equal(t, tc.want, got.GetMsg())
			require.Equal(t, prototype.ID(), got.ID())
			require.True(t, HasStack(err))
			file, line := got.Location()
			require.True(t, strings.HasSuffix(file, "normalize_test.go"))
			require.Positive(t, line)
			require.Equal(t, "value %s, count %04d, public %s", prototype.message)
			require.Empty(t, prototype.args)
		})
	}
}

func TestFastGenByArgsRedaction(t *testing.T) {
	previousMode := RedactLogEnabled.Load()
	t.Cleanup(func() { RedactLogEnabled.Store(previousMode) })
	prototype := Normalize("value %s, count %04d, public %s", RFCCodeText("Internal:Test"), RedactArgs([]int{0, 1, 5}))
	const format = "value %s, count %04d, public %s"

	for _, tc := range []struct {
		mode string
		want string
	}{
		{RedactLogDisable, "value secret‹›, count 0042, public visible"},
		{RedactLogEnable, "value ?, count ?, public visible"},
		{RedactLogMarker, "value ‹secret‹‹›››, count ‹0042›, public visible"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			RedactLogEnabled.Store(tc.mode)
			origin := []byte("secret‹›")
			err := prototype.FastGenByArgs(hackedStringArg{raw: origin}, 42, "visible")
			copy(origin, "modified")
			got := err.(*withStack).error.(*Error)
			require.Equal(t, tc.want, got.GetMsg())
			require.Equal(t, prototype.ID(), got.ID())
			require.False(t, HasStack(err))
			file, line := got.Location()
			require.Empty(t, file)
			require.Zero(t, line)
			require.Equal(t, format, prototype.message)
			require.Empty(t, prototype.args)
		})
	}
}

func TestGenWithStackByCauseRedaction(t *testing.T) {
	previousMode := RedactLogEnabled.Load()
	t.Cleanup(func() { RedactLogEnabled.Store(previousMode) })
	prototype := Normalize("unused template", RFCCodeText("Internal:Test"), RedactArgs([]int{0}))
	causeTemplate := Normalize("cause %s, public %s, 95%% full, %%1000000[1]d", RFCCodeText("Internal:Cause"), RedactArgs([]int{0}))

	for _, tc := range []struct {
		mode string
		arg  string
	}{
		{RedactLogDisable, "secret‹›"},
		{RedactLogEnable, "?"},
		{RedactLogMarker, "‹secret‹‹›››"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			RedactLogEnabled.Store(tc.mode)
			cause := causeTemplate.GenWithStackByArgs("secret‹›", "visible")
			err := prototype.Wrap(cause).GenWithStackByCause()
			want := "[Internal:Cause]cause " + tc.arg + ", public visible, 95% full, %1000000[1]d"
			got := err.(*withStack).error.(*Error)
			require.Equal(t, want, got.GetMsg())
			require.EqualError(t, err, "[Internal:Test]"+want+": "+want)
		})
	}
}

func TestFastGenWithCauseRedaction(t *testing.T) {
	previousMode := RedactLogEnabled.Load()
	t.Cleanup(func() { RedactLogEnabled.Store(previousMode) })
	prototype := Normalize("unused template", RFCCodeText("Internal:Test"), RedactArgs([]int{0}))
	causeTemplate := Normalize("cause %s, public %s, 95%% full, %%1000000[1]d", RFCCodeText("Internal:Cause"), RedactArgs([]int{0}))

	for _, tc := range []struct {
		mode string
		arg  string
	}{
		{RedactLogDisable, "secret‹›"},
		{RedactLogEnable, "?"},
		{RedactLogMarker, "‹secret‹‹›››"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			RedactLogEnabled.Store(tc.mode)
			cause := causeTemplate.GenWithStackByArgs("secret‹›", "visible")
			err := prototype.Wrap(cause).FastGenWithCause()
			want := "[Internal:Cause]cause " + tc.arg + ", public visible, 95% full, %1000000[1]d"
			got := err.(*withStack).error.(*Error)
			require.Equal(t, want, got.GetMsg())
			require.EqualError(t, err, "[Internal:Test]"+want+": "+want)
		})
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
	wrapped := errTest.Wrap(err).GenWithStackByCause()
	require.Equal(t, err.Error(), wrapped.(*withStack).error.(*Error).GetMsg())
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
	wrapped := errTest.Wrap(err).FastGenWithCause()
	require.Equal(t, err.Error(), wrapped.(*withStack).error.(*Error).GetMsg())
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
