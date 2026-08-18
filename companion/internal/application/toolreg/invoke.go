package toolreg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Handler executes a tool call and returns its full item list; the
// registry applies pagination on top, so handlers stay simple.
type Handler func(ctx context.Context, call Call) ([]json.RawMessage, error)

// Call is the validated invocation the handler receives.
type Call struct {
	Correlation string
	Args        json.RawMessage
}

// Page bounds the result window requested by the client.
type Page struct {
	Offset int
	Limit  int
}

// Result is the paginated response: NextOffset is 0 when exhausted.
type Result struct {
	Items      []json.RawMessage
	NextOffset int
}

// Invoke runs a registered tool with all budgets applied: unknown tools
// are denied, arguments are size-capped strict JSON objects, cancellation
// is honored before dispatch, correlation is mandatory, and results are
// paginated within the registry's window.
func (r *Registry) Invoke(
	ctx context.Context,
	name, correlation string,
	args json.RawMessage,
	page Page,
) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if correlation == "" {
		return Result{}, fmt.Errorf("toolreg: correlation ID is mandatory")
	}
	tool, known := r.tools[name]
	if !known {
		return Result{}, fmt.Errorf("toolreg: tool %q denied (not registered)", name)
	}
	if len(args) > r.maxArgBytes {
		return Result{}, fmt.Errorf("toolreg: arguments exceed %d bytes", r.maxArgBytes)
	}
	if len(args) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(args))
		var parsed map[string]json.RawMessage
		if err := decoder.Decode(&parsed); err != nil || decoder.More() {
			return Result{}, fmt.Errorf("toolreg: arguments must be one JSON object")
		}
	}
	items, err := tool.Handler(ctx, Call{Correlation: correlation, Args: args})
	if err != nil {
		return Result{}, err
	}
	return paginate(items, page, r.maxPageSize), nil
}

// paginate clamps the window to the registry budget and reports the next
// offset (0 when the listing is exhausted).
func paginate(items []json.RawMessage, page Page, maxPage int) Result {
	limit := page.Limit
	if limit <= 0 || limit > maxPage {
		limit = maxPage
	}
	start := min(max(page.Offset, 0), len(items))
	end := min(start+limit, len(items))
	result := Result{Items: items[start:end]}
	if end < len(items) {
		result.NextOffset = end
	}
	return result
}
