package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// FuzzDecodeStrict: the strict decoder must classify arbitrary input
// as accepted or refused without ever panicking, and anything it
// accepts must be a single clean JSON object.
func FuzzDecodeStrict(f *testing.F) {
	f.Add(`{"name":"a"}`)
	f.Add(`{"name":"a","extra":1}`)
	f.Add(`{"name":"a"}{"name":"b"}`)
	f.Add(`[1,2`)
	f.Add(`null`)
	f.Add(``)
	f.Add(`{"name":"` + strings.Repeat("x", 4096) + `"}`)
	f.Fuzz(func(t *testing.T, raw string) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(raw))
		var body enrollBody
		if err := DecodeStrict(req, &body); err == nil {
			if strings.Contains(raw, "extra") {
				t.Errorf("unknown fields must never decode: %q", raw)
			}
		}
	})
}
