package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func ToJSON(t *testing.T, v any) string {
	return string(ToJSONBytes(t, v))
}

func ToJSONBytes(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func ToJSONReader(t *testing.T, v any) io.Reader {
	return bytes.NewReader([]byte(ToJSON(t, v)))
}
