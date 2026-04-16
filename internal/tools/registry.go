package tools

import (
	"context"
	"fmt"
)

type Tool interface {
	Name() string
	Run(ctx context.Context, input map[string]any) (map[string]any, error)
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}
func (r *Registry) Get(name string) (Tool, error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not fount: %s", name)
	}
	return t, nil
}
