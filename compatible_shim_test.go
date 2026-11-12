package errors

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalJSONClearsLocalState(t *testing.T) {
	e := Error{
		message:       "old %s",
		args:          []any{"secret"},
		cause:         New("old cause"),
		redactArgsPos: []int{0},
		file:          "old.go",
		line:          42,
	}
	const message = "literal %s, %1000000[1]d, 95% full"
	require.NoError(t, e.UnmarshalJSON([]byte(`{"class":6,"code":123,"message":"`+message+`","rfccode":"executor:123"}`)))
	require.Empty(t, e.Args())
	require.Equal(t, message, e.GetMsg())
	require.Equal(t, Error{
		code:     123,
		codeText: "executor:123",
		message:  message,
	}, e)

	data, err := json.Marshal(&e)
	require.NoError(t, err)
	var decoded jsonError
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, message, decoded.Msg)
}

func TestUnmarshalJSONInvalidInputPreservesError(t *testing.T) {
	for _, input := range []string{
		`{"message":"new message",`,
		`{"message":"new message","code":"invalid"}`,
	} {
		t.Run(input, func(t *testing.T) {
			e := Error{
				code:          123,
				codeText:      "executor:123",
				message:       "old %s",
				args:          []any{"secret"},
				cause:         New("old cause"),
				redactArgsPos: []int{0},
				file:          "old.go",
				line:          42,
			}
			before := e
			require.Error(t, e.UnmarshalJSON([]byte(input)))
			require.Equal(t, before, e)
			require.Equal(t, "old secret", e.GetMsg())
		})
	}
}
