package scheme

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	err := r.Register(TypeRegistration{})
	assert.Error(t, err)

	err = r.Register(TypeRegistration{Name: "task"})
	assert.NoError(t, err)

	// Re-registering the same name should error.
	err = r.Register(TypeRegistration{Name: "task"})
	assert.Error(t, err)
}

func TestRegistry_Parse(t *testing.T) {
	r := NewRegistry()

	_, err := r.Parse("task://hop-top/uri/T-0001")
	assert.Error(t, err)

	assert.NoError(t, r.Register(TypeRegistration{Name: "task"}))
	got, err := r.Parse("task://hop-top/uri/T-0001")
	require.NoError(t, err)
	assert.Equal(t, "task", got.Scheme)
	assert.Equal(t, "hop-top/uri", got.Namespace)
	assert.Equal(t, "T-0001", got.ID)

	assert.NoError(t, r.Register(TypeRegistration{
		Name: "repo",
		Parser: func(input string) (*URI, error) {
			assert.Equal(t, "repo://hop-top/uri", input)
			return &URI{Scheme: "repo", Namespace: "custom", ID: "parsed"}, nil
		},
	}))
	got, err = r.Parse("repo://hop-top/uri")
	require.NoError(t, err)
	assert.Equal(t, "custom", got.Namespace)
	assert.Equal(t, "parsed", got.ID)
}

func TestRegistry_Parse_UsesPolicyVanity(t *testing.T) {
	r := NewRegistryWithPolicy(Policy{
		DefaultNamespaceSegments: 1,
		SchemeNamespaceSegments: map[string]int{
			"task": 2,
		},
		VanityAliases: []VanityAlias{
			{
				From: "task://shortcut",
				To:   "task://hop-top/uri/T-0001",
			},
		},
	})
	require.NoError(t, r.Register(TypeRegistration{Name: "task"}))

	got, err := r.Parse("task://shortcut")
	require.NoError(t, err)
	assert.Equal(t, "task://hop-top/uri/T-0001", got.Canonical())
	assert.Equal(t, "task://shortcut", got.Vanity())
}

func TestRegistry_CompleteVanity(t *testing.T) {
	r := NewRegistryWithPolicy(Policy{
		VanityAliases: []VanityAlias{
			{From: "task://shortcuta", To: "task://hop-top/uri/T-0001"},
			{From: "task://shortcutb", To: "task://hop-top/uri/T-0002"},
		},
	})

	got := r.CompleteVanity("task://shortcut")
	require.Len(t, got, 2)
	assert.Equal(t, "task://shortcuta", got[0].From)
	assert.Equal(t, "task://shortcutb", got[1].From)
}

func TestRegistry_Complete(t *testing.T) {
	r := NewRegistry()

	called := false
	err := r.Register(TypeRegistration{
		Name: "task",
		Completer: func(ctx context.Context, prefix string) ([]string, error) {
			called = true
			assert.Equal(t, "T-00", prefix)
			return []string{"T-0001", "T-0002"}, nil
		},
	})
	assert.NoError(t, err)

	// Unknown type returns an error.
	_, err = r.Complete(context.Background(), "unknown", "")
	assert.Error(t, err)

	// Registered type with completer returns suggestions.
	got, err := r.Complete(context.Background(), "task", "T-00")
	assert.NoError(t, err)
	assert.Equal(t, []string{"T-0001", "T-0002"}, got)
	assert.True(t, called)

	// Registered type without completer returns (nil, nil).
	err = r.Register(TypeRegistration{Name: "aps"})
	assert.NoError(t, err)
	got, err = r.Complete(context.Background(), "aps", "")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestRegistry_Types(t *testing.T) {
	r := NewRegistry()
	assert.Empty(t, r.Types())

	assert.NoError(t, r.Register(TypeRegistration{Name: "task"}))
	assert.NoError(t, r.Register(TypeRegistration{Name: "repo"}))
	assert.NoError(t, r.Register(TypeRegistration{Name: "doc"}))

	got := r.Types()
	assert.True(t, sort.StringsAreSorted(got))
	assert.Equal(t, []string{"doc", "repo", "task"}, got)
}
