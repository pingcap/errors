package errors

import (
	"fmt"
	"testing"

	stderrors "errors"
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

func BenchmarkByArgsArgFreeze(b *testing.B) {
	errPrototype := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Bench"))
	stringArg := "120120519090607"

	b.Run("FastGenByArgs-string", func(b *testing.B) {
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err = errPrototype.FastGenByArgs(stringArg)
		}
		GlobalE = err
	})

	b.Run("FastGenByArgs-int", func(b *testing.B) {
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err = errPrototype.FastGenByArgs(120120519090607)
		}
		GlobalE = err
	})

	b.Run("GenWithStackByArgs-string", func(b *testing.B) {
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err = errPrototype.GenWithStackByArgs(stringArg)
		}
		GlobalE = err
	})

	b.Run("GenWithStackByArgs-int", func(b *testing.B) {
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err = errPrototype.GenWithStackByArgs(120120519090607)
		}
		GlobalE = err
	})
}

func BenchmarkFormatArgFreeze(b *testing.B) {
	errPrototype := Normalize("Incorrect time value: '%s'", RFCCodeText("Internal:Bench"))
	stringArg := "120120519090607"

	b.Run("FastGen-string", func(b *testing.B) {
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err = errPrototype.FastGen("Incorrect time value: '%s'", stringArg)
		}
		GlobalE = err
	})

	b.Run("FastGen-int", func(b *testing.B) {
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err = errPrototype.FastGen("Incorrect time value: '%d'", 120120519090607)
		}
		GlobalE = err
	})

	b.Run("GenWithStack-string", func(b *testing.B) {
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err = errPrototype.GenWithStack("Incorrect time value: '%s'", stringArg)
		}
		GlobalE = err
	})

	b.Run("GenWithStack-int", func(b *testing.B) {
		var err error
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err = errPrototype.GenWithStack("Incorrect time value: '%d'", 120120519090607)
		}
		GlobalE = err
	})
}
