package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
	"testing"
)

func noErrors(at, depth int) error {
	if at >= depth {
		return stderrors.New("no error")
	}
	return noErrors(at+1, depth)
}

func yesErrors(at, depth int) error {
	if at >= depth {
		return New("ye error")
	}
	return yesErrors(at+1, depth)
}

// GlobalE is an exported global to store the result of benchmark results,
// preventing the compiler from optimising the benchmark functions away.
var GlobalE interface{}

func BenchmarkErrors(b *testing.B) {
	type run struct {
		stack int
		std   bool
	}
	runs := []run{
		{10, false},
		{10, true},
		{100, false},
		{100, true},
		{1000, false},
		{1000, true},
	}
	for _, r := range runs {
		part := "pkg/errors"
		if r.std {
			part = "errors"
		}
		name := fmt.Sprintf("%s-stack-%d", part, r.stack)
		b.Run(name, func(b *testing.B) {
			var err error
			f := yesErrors
			if r.std {
				f = noErrors
			}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				err = f(0, r.stack)
			}
			b.StopTimer()
			GlobalE = err
		})
	}
}

func BenchmarkStackFormatting(b *testing.B) {
	type run struct {
		stack  int
		format string
	}
	runs := []run{
		{10, "%s"},
		{10, "%v"},
		{10, "%+v"},
		{30, "%s"},
		{30, "%v"},
		{30, "%+v"},
		{60, "%s"},
		{60, "%v"},
		{60, "%+v"},
	}

	var stackStr string
	for _, r := range runs {
		name := fmt.Sprintf("%s-stack-%d", r.format, r.stack)
		b.Run(name, func(b *testing.B) {
			err := yesErrors(0, r.stack)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				stackStr = fmt.Sprintf(r.format, err)
			}
			b.StopTimer()
		})
	}

	for _, r := range runs {
		name := fmt.Sprintf("%s-stacktrace-%d", r.format, r.stack)
		b.Run(name, func(b *testing.B) {
			err := yesErrors(0, r.stack)
			st := err.(*fundamental).stack.StackTrace()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				stackStr = fmt.Sprintf(r.format, st)
			}
			b.StopTimer()
		})
	}
	GlobalE = stackStr
}

type argsProfile struct {
	name           string
	containsString bool
	build          func(count, stringLen int) []interface{}
}

type benchmarkHackedStr string

func (s benchmarkHackedStr) FreezeStr() string {
	return string(append([]byte(nil), s...))
}

func buildHackedStringArgs(count, stringLen int) []interface{} {
	arg := benchmarkHackedStr(strings.Repeat("x", stringLen))
	args := make([]interface{}, count)
	for i := range args {
		args[i] = arg
	}
	return args
}

func buildPlainStringArgs(count, stringLen int) []interface{} {
	arg := strings.Repeat("x", stringLen)
	args := make([]interface{}, count)
	for i := range args {
		args[i] = arg
	}
	return args
}

func buildIntArgs(count, _ int) []interface{} {
	args := make([]interface{}, count)
	for i := range args {
		args[i] = i
	}
	return args
}

func BenchmarkByArgsHackedStrFreeze(b *testing.B) {
	errPrototype := Normalize("bench", RFCCodeText("Internal:Bench"))

	apiCases := []struct {
		name string
		call func(errPrototype *Error, args []interface{}) error
	}{
		{
			name: "FastGenByArgs",
			call: func(errPrototype *Error, args []interface{}) error {
				return errPrototype.FastGenByArgs(args...)
			},
		},
	}
	profiles := []argsProfile{
		{name: "plain", containsString: true, build: buildPlainStringArgs},
		{name: "hacked", containsString: true, build: buildHackedStringArgs},
	}
	argCounts := []int{1, 4, 8}
	stringLens := []int{16, 1024}

	for _, apiCase := range apiCases {
		apiCase := apiCase
		b.Run(apiCase.name, func(b *testing.B) {
			for _, profile := range profiles {
				profile := profile
				lens := []int{0}
				if profile.containsString {
					lens = stringLens
				}
				for _, argCount := range argCounts {
					argCount := argCount
					for _, strLen := range lens {
						strLen := strLen
						templateArgs := profile.build(argCount, strLen)
						caseName := fmt.Sprintf("type-%s/count-%d", profile.name, argCount)
						if profile.containsString {
							caseName = fmt.Sprintf("%s/strlen-%d", caseName, strLen)
						}

						b.Run(caseName, func(b *testing.B) {
							var err error
							args := make([]interface{}, len(templateArgs))
							b.ReportAllocs()
							for i := 0; i < b.N; i++ {
								copy(args, templateArgs)
								err = apiCase.call(errPrototype, args)
							}
							GlobalE = err
						})
					}
				}
			}
		})
	}
}
