package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type enrollBody struct {
	Name string `json:"name"`
}

func decodeInto(t *testing.T, raw string) error {
	t.Helper()
	req := httptest.NewRequest("POST", "/", strings.NewReader(raw))
	var body enrollBody
	return DecodeStrict(req, &body)
}

func TestDecodeStrictAcceptsExactPayloads(t *testing.T) {
	if err := decodeInto(t, `{"name":"a"}`); err != nil {
		t.Fatalf("exact payload must decode: %v", err)
	}
}

func TestDecodeStrictRefusesUnknownFields(t *testing.T) {
	if decodeInto(t, `{"name":"a","extra":1}`) == nil {
		t.Fatal("unknown fields must be refused")
	}
}

func TestDecodeStrictRefusesTrailingData(t *testing.T) {
	for _, raw := range []string{`{"name":"a"}{"name":"b"}`, `{"name":"a"} x`} {
		if decodeInto(t, raw) == nil {
			t.Fatalf("trailing data must be refused: %s", raw)
		}
	}
}
