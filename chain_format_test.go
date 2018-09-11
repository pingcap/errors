package errors

import (
	"fmt"
	"testing"
	"time"
)

func resetErrDataTooLong(colName string, rowIdx int, err error) error {
	newErr := Errorf("Data too long for column '%v' at row %v", colName, rowIdx)
	return BuildChain(newErr, err)
}
func TestChainFormat(t *testing.T) {
	_, errTime := time.ParseDuration("dfasdf")
	err1 := resetErrDataTooLong("c", 1, errTime)
	err2 := resetErrDataTooLong("b", 2, err1)

	expectedV := "time: invalid duration dfasdf\n" +
		"Data too long for column 'c' at row 1\n" +
		"Data too long for column 'b' at row 2"
	if fmt.Sprintf("%v", err2) != expectedV {
		t.Errorf("expected %v", expectedV)
	}
	if fmt.Sprintf("%s", err2) != expectedV {
		t.Errorf("expected %v", expectedV)
	}

	expectedS := "\"time: invalid duration dfasdf\"\n" +
		"\"Data too long for column 'c' at row 1\"\n" +
		"\"Data too long for column 'b' at row 2\""
	gotQ := fmt.Sprintf("%q", err2)
	if gotQ != expectedS {
		t.Errorf("expected %v, got %v", expectedS, gotQ)
	}

	expectedPV := `time: invalid duration dfasdf
Data too long for column 'c' at row 1
github.com/pkg/errors.resetErrDataTooLong
	.+/github.com/pkg/errors/chain_format_test.go:10
github.com/pkg/errors.TestChainFormat
	.+/github.com/pkg/errors/chain_format_test.go:15
testing.tRunner
	.+testing.go:777
runtime.goexit
	.+asm_amd64.s:2361
.*Data too long for column 'b' at row 2`

	testFormatRegexp(t, 0, err2, "%+v", expectedPV)
}
