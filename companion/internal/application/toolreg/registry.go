// Package toolreg implements the typed assistant tool registry (OpenSpec
// introduce-agentic-go-companion, task 6.1): JSON-schema-described tools,
// argument size limits, bounded result pagination, cancellation and
// correlation IDs. Unknown tools are denied, and dangerous tool names can
// never be registered at all (task 6.6, defense in depth).
package toolreg

import (
	"encoding/json"
	"fmt"
	"strings"
)

// forbiddenNames can never be registered: these capabilities must be
// absent from the product by construction, not merely unregistered.
var forbiddenNames = []string{
	"memory", "input", "gameplay", "play", "purchase", "buy",
	"account", "credential", "password", "shell", "exec", "filesystem", "file-write",
}

// Tool is a registered, typed assistant tool.
type Tool struct {
	Name        string
	Description string
	Schema      json.RawMessage // JSON schema of the arguments, for clients
	Handler     Handler
}

// Registry holds the tools and the global invocation budgets.
type Registry struct {
	tools       map[string]Tool
	maxArgBytes int
	maxPageSize int
}

// New builds the registry with its budgets; zero values get safe defaults.
func New(maxArgBytes, maxPageSize int) *Registry {
	if maxArgBytes <= 0 {
		maxArgBytes = 16 << 10
	}
	if maxPageSize <= 0 {
		maxPageSize = 50
	}
	return &Registry{tools: make(map[string]Tool),
		maxArgBytes: maxArgBytes, maxPageSize: maxPageSize}
}

// Register validates and adds a tool; forbidden capability names are
// rejected permanently and duplicates are errors.
func (r *Registry) Register(tool Tool) error {
	if tool.Name == "" || tool.Handler == nil || len(tool.Schema) == 0 {
		return fmt.Errorf("toolreg: tool %q incomplete (name, schema, handler)", tool.Name)
	}
	lower := strings.ToLower(tool.Name)
	for _, forbidden := range forbiddenNames {
		if strings.Contains(lower, forbidden) {
			return fmt.Errorf("toolreg: capability %q is forbidden by policy", tool.Name)
		}
	}
	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("toolreg: tool %q already registered", tool.Name)
	}
	r.tools[tool.Name] = tool
	return nil
}

// List returns the registered tool names and schemas for clients.
func (r *Registry) List() []Tool {
	list := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		list = append(list, Tool{Name: tool.Name, Description: tool.Description,
			Schema: tool.Schema})
	}
	return list
}
