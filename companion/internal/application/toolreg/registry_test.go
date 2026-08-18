package toolreg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func listTool(t *testing.T, items int) Tool {
	t.Helper()
	return Tool{
		Name: "search-cards", Description: "pure query over the snapshot",
		Schema: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"}}}`),
		Handler: func(_ context.Context, call Call) ([]json.RawMessage, error) {
			if call.Correlation == "" {
				t.Fatal("handler must receive the correlation ID")
			}
			var out []json.RawMessage
			for i := range items {
				out = append(out, json.RawMessage(fmt.Sprintf(`{"card":%d}`, i)))
			}
			return out, nil
		},
	}
}

func TestInvokeAppliesBudgetsAndPagination(t *testing.T) {
	registry := New(64, 10)
	if err := registry.Register(listTool(t, 25)); err != nil {
		t.Fatalf("register: %v", err)
	}
	first, err := registry.Invoke(t.Context(), "search-cards", "corr-1",
		json.RawMessage(`{"q":"red"}`), Page{Limit: 100}) // clamped to 10
	if err != nil || len(first.Items) != 10 || first.NextOffset != 10 {
		t.Fatalf("first page unexpected: %d items, next %d (%v)",
			len(first.Items), first.NextOffset, err)
	}
	last, err := registry.Invoke(t.Context(), "search-cards", "corr-1", nil,
		Page{Offset: 20, Limit: 10})
	if err != nil || len(last.Items) != 5 || last.NextOffset != 0 {
		t.Fatalf("last page unexpected: %d items, next %d (%v)",
			len(last.Items), last.NextOffset, err)
	}
}

func TestInvokeRejectsBadCalls(t *testing.T) {
	registry := New(32, 10)
	if err := registry.Register(listTool(t, 1)); err != nil {
		t.Fatalf("register: %v", err)
	}
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	cases := map[string]func() error{
		"unknown tool denied": func() error {
			_, err := registry.Invoke(t.Context(), "ghost", "c", nil, Page{})
			return err
		},
		"missing correlation": func() error {
			_, err := registry.Invoke(t.Context(), "search-cards", "", nil, Page{})
			return err
		},
		"oversized arguments": func() error {
			args := json.RawMessage(`{"q":"` + strings.Repeat("x", 64) + `"}`)
			_, err := registry.Invoke(t.Context(), "search-cards", "c", args, Page{})
			return err
		},
		"malformed arguments": func() error {
			_, err := registry.Invoke(t.Context(), "search-cards", "c",
				json.RawMessage(`[1,2]`), Page{})
			return err
		},
		"cancellation honored": func() error {
			_, err := registry.Invoke(cancelled, "search-cards", "c", nil, Page{})
			return err
		},
	}
	for name, call := range cases {
		if call() == nil {
			t.Errorf("%s: should fail", name)
		}
	}
}
