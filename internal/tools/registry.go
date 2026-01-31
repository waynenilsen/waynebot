package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// ToolFunc is the signature for all tool implementations.
type ToolFunc func(ctx context.Context, args json.RawMessage) (string, error)

// Registry maps tool names to their implementations.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]ToolFunc
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]ToolFunc)}
}

// Register adds a tool to the registry. It returns an error if the name is
// already registered.
func (r *Registry) Register(name string, fn ToolFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tools[name]; ok {
		return fmt.Errorf("tool %q already registered", name)
	}
	r.tools[name] = fn
	return nil
}

// Call invokes the named tool with the given arguments.
func (r *Registry) Call(ctx context.Context, name string, args json.RawMessage) (string, error) {
	r.mu.RLock()
	fn, ok := r.tools[name]
	r.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("unknown tool %q", name)
	}
	return fn(ctx, args)
}

// RegisterDefaults registers all built-in tools using the given sandbox config.
func (r *Registry) RegisterDefaults(cfg *SandboxConfig) {
	r.Register("shell_exec", ShellExec(cfg))
	r.Register("file_read", FileRead(cfg))
	r.Register("file_write", FileWrite(cfg))
	r.Register("http_fetch", HTTPFetch(cfg))
}

// Names returns the sorted list of registered tool names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for n := range r.tools {
		names = append(names, n)
	}
	return names
}
