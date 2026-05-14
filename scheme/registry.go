package scheme

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Parser parses a URI string for a registered URI type.
type Parser func(input string) (*URI, error)

// Completer returns a list of suggested values for a URI type.
type Completer func(ctx context.Context, prefix string) ([]string, error)

// TypeRegistration defines how a specific URI type is handled.
type TypeRegistration struct {
	Name      string
	Parser    Parser
	Completer Completer
}

// Registry manages URI types and their parsers/completion logic.
type Registry struct {
	mu     sync.RWMutex
	types  map[string]TypeRegistration
	policy Policy
}

// NewRegistry creates a new URI type registry.
func NewRegistry() *Registry {
	return NewRegistryWithPolicy(DefaultPolicy)
}

// NewRegistryWithPolicy creates a new URI type registry with a parser policy.
func NewRegistryWithPolicy(policy Policy) *Registry {
	return &Registry{
		types:  make(map[string]TypeRegistration),
		policy: policy,
	}
}

// Register adds a new URI type to the registry.
func (r *Registry) Register(reg TypeRegistration) error {
	if reg.Name == "" {
		return fmt.Errorf("uri: registration name is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.types[reg.Name]; exists {
		return fmt.Errorf("uri: type %q already registered", reg.Name)
	}

	r.types[reg.Name] = reg
	return nil
}

// Parse parses input and dispatches to a registered parser when one exists.
func (r *Registry) Parse(input string) (*URI, error) {
	parsed, err := ParseWithPolicy(input, r.policy)
	if err != nil {
		return nil, err
	}

	r.mu.RLock()
	reg, exists := r.types[parsed.Scheme]
	r.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("uri: unknown type %q", parsed.Scheme)
	}
	if reg.Parser != nil {
		return reg.Parser(input)
	}
	return parsed, nil
}

// CompleteVanity returns fuzzy vanity aliases that may match input.
func (r *Registry) CompleteVanity(input string) []VanityCandidate {
	r.mu.RLock()
	policy := r.policy
	r.mu.RUnlock()
	return policy.VanityCandidates(input)
}

// Complete returns suggestions for a given URI type.
func (r *Registry) Complete(ctx context.Context, typeName, prefix string) ([]string, error) {
	r.mu.RLock()
	reg, exists := r.types[typeName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("uri: unknown type %q", typeName)
	}
	if reg.Completer == nil {
		return nil, nil
	}
	return reg.Completer(ctx, prefix)
}

// Types returns registered type names in deterministic order.
func (r *Registry) Types() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.types))
	for name := range r.types {
		types = append(types, name)
	}
	sort.Strings(types)
	return types
}
