package scheme

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type contractFixture struct {
	NamespacePolicy struct {
		DefaultNamespaceSegments int            `json:"defaultNamespaceSegments"`
		SchemeNamespaceSegments  map[string]int `json:"schemeNamespaceSegments"`
	} `json:"namespacePolicy"`
	ValidCases []struct {
		Name     string `json:"name"`
		Input    string `json:"input"`
		Expected struct {
			Scheme    string `json:"scheme"`
			Namespace string `json:"namespace"`
			ID        string `json:"id"`
			Query     string `json:"query"`
			Fragment  string `json:"fragment"`
			Original  string `json:"original"`
			Action    string `json:"action"`
		} `json:"expected"`
		Canonical string `json:"canonical"`
	} `json:"validCases"`
	InvalidCases []struct {
		Name  string `json:"name"`
		Input string `json:"input"`
	} `json:"invalidCases"`
	VanityCases []struct {
		Name    string `json:"name"`
		Input   string `json:"input"`
		Aliases []struct {
			From           string `json:"from"`
			To             string `json:"to"`
			Prefix         bool   `json:"prefix"`
			PreserveSuffix bool   `json:"preserveSuffix"`
		} `json:"aliases"`
		Options struct {
			Strict        bool `json:"strict"`
			JSONAmbiguity bool `json:"jsonAmbiguity"`
		} `json:"options"`
		Valid    bool `json:"valid"`
		Expected struct {
			Scheme    string `json:"scheme"`
			Namespace string `json:"namespace"`
			ID        string `json:"id"`
			Query     string `json:"query"`
			Fragment  string `json:"fragment"`
			Original  string `json:"original"`
			Action    string `json:"action"`
		} `json:"expected"`
		Canonical string `json:"canonical"`
	} `json:"vanityCases"`
	ActionCases []struct {
		Name         string                 `json:"name"`
		Input        string                 `json:"input"`
		ActionRoutes map[string]ActionRoute `json:"actionRoutes"`
		Expected     struct {
			Scheme    string `json:"scheme"`
			Namespace string `json:"namespace"`
			ID        string `json:"id"`
			Query     string `json:"query"`
			Fragment  string `json:"fragment"`
			Original  string `json:"original"`
			Action    string `json:"action"`
		} `json:"expected"`
		Canonical string   `json:"canonical"`
		Command   string   `json:"command"`
		Args      []string `json:"args"`
	} `json:"actionCases"`
}

func loadContract(t *testing.T) contractFixture {
	t.Helper()

	raw, err := os.ReadFile("../../spec/fixtures/uri-contract.json")
	require.NoError(t, err)

	var fixture contractFixture
	require.NoError(t, json.Unmarshal(raw, &fixture))
	return fixture
}

func fixturePolicy(fixture contractFixture) Policy {
	return Policy{
		DefaultNamespaceSegments: fixture.NamespacePolicy.DefaultNamespaceSegments,
		SchemeNamespaceSegments:  fixture.NamespacePolicy.SchemeNamespaceSegments,
	}
}

func TestParse_ContractValid(t *testing.T) {
	fixture := loadContract(t)

	policy := fixturePolicy(fixture)
	for _, tt := range fixture.ValidCases {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := ParseWithPolicy(tt.Input, policy)
			require.NoError(t, err)
			assert.Equal(t, tt.Expected.Scheme, got.Scheme)
			assert.Equal(t, tt.Expected.Namespace, got.Namespace)
			assert.Equal(t, tt.Expected.ID, got.ID)
			assert.Equal(t, tt.Expected.Query, got.Query)
			assert.Equal(t, tt.Expected.Fragment, got.Fragment)
			assert.Equal(t, tt.Expected.Original, got.Original)
			assert.Equal(t, tt.Expected.Action, got.Action)
			assert.Equal(t, tt.Canonical, got.Canonical())
			assert.Equal(t, tt.Canonical, got.String())
		})
	}
}

func TestParse_ContractInvalid(t *testing.T) {
	fixture := loadContract(t)

	policy := fixturePolicy(fixture)
	for _, tt := range fixture.InvalidCases {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := ParseWithPolicy(tt.Input, policy)
			assert.Error(t, err)
		})
	}
}

func TestParse_ContractVanityCases(t *testing.T) {
	fixture := loadContract(t)
	policy := fixturePolicy(fixture)

	for _, tt := range fixture.VanityCases {
		t.Run(tt.Name, func(t *testing.T) {
			policy.VanityAliases = nil
			for _, alias := range tt.Aliases {
				policy.VanityAliases = append(policy.VanityAliases, VanityAlias{
					From:           alias.From,
					To:             alias.To,
					Prefix:         alias.Prefix,
					PreserveSuffix: alias.PreserveSuffix,
				})
			}

			var opts []ParseOption
			if tt.Options.Strict {
				opts = append(opts, WithStrict())
			}
			if tt.Options.JSONAmbiguity {
				opts = append(opts, WithJSONAmbiguity())
			}

			got, err := ParseWithPolicyOptions(tt.Input, policy, opts...)
			if !tt.Valid {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.Expected.Scheme, got.Scheme)
			assert.Equal(t, tt.Expected.Namespace, got.Namespace)
			assert.Equal(t, tt.Expected.ID, got.ID)
			assert.Equal(t, tt.Expected.Query, got.Query)
			assert.Equal(t, tt.Expected.Fragment, got.Fragment)
			assert.Equal(t, tt.Expected.Original, got.Original)
			assert.Equal(t, tt.Expected.Action, got.Action)
			assert.Equal(t, tt.Canonical, got.Canonical())
			assert.Equal(t, tt.Expected.Original, got.Vanity())
		})
	}
}

func TestParse_ContractActionCases(t *testing.T) {
	fixture := loadContract(t)
	policy := fixturePolicy(fixture)

	for _, tt := range fixture.ActionCases {
		t.Run(tt.Name, func(t *testing.T) {
			policy.ActionRoutes = tt.ActionRoutes

			got, err := ParseWithPolicy(tt.Input, policy)
			require.NoError(t, err)
			assert.Equal(t, tt.Expected.Scheme, got.Scheme)
			assert.Equal(t, tt.Expected.Namespace, got.Namespace)
			assert.Equal(t, tt.Expected.ID, got.ID)
			assert.Equal(t, tt.Expected.Action, got.Action)
			assert.Equal(t, tt.Canonical, got.Canonical())

			resolved, err := policy.ResolveAction(got)
			require.NoError(t, err)
			assert.Equal(t, got.Action, resolved.Action)
			assert.Equal(t, tt.Command, resolved.Command)
			assert.Equal(t, tt.Args, resolved.Args)
		})
	}
}

func TestParseWithPolicy_CustomNamespaceSegments(t *testing.T) {
	got, err := ParseWithPolicy("item://tenant/project/object/42", Policy{
		DefaultNamespaceSegments: 2,
	})
	require.NoError(t, err)

	assert.Equal(t, "item", got.Scheme)
	assert.Equal(t, "tenant/project", got.Namespace)
	assert.Equal(t, "object/42", got.ID)
	assert.Equal(t, "item://tenant/project/object/42", got.Canonical())
}

func TestParseWithPolicy_VanityExactAlias(t *testing.T) {
	got, err := ParseWithPolicy("task://my-custom-slug/path-optional", Policy{
		DefaultNamespaceSegments: 1,
		SchemeNamespaceSegments: map[string]int{
			"task": 2,
		},
		VanityAliases: []VanityAlias{
			{
				From: "task://my-custom-slug/path-optional",
				To:   "task://hop-top/uri/T-0001",
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "task", got.Scheme)
	assert.Equal(t, "hop-top/uri", got.Namespace)
	assert.Equal(t, "T-0001", got.ID)
	assert.Equal(t, "task://hop-top/uri/T-0001", got.Canonical())
	assert.Equal(t, "task://my-custom-slug/path-optional", got.Vanity())
}

func TestParseWithPolicy_VanityPrefixAlias(t *testing.T) {
	got, err := ParseWithPolicy("task://shortcut/child", Policy{
		DefaultNamespaceSegments: 1,
		SchemeNamespaceSegments: map[string]int{
			"task": 2,
		},
		VanityAliases: []VanityAlias{
			{
				From:           "task://shortcut",
				To:             "task://hop-top/uri/T-0001",
				Prefix:         true,
				PreserveSuffix: true,
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "task://hop-top/uri/T-0001/child", got.Canonical())
	assert.Equal(t, "task://shortcut/child", got.Vanity())
}

func TestParseWithPolicyOptions_VanityFuzzyClosest(t *testing.T) {
	got, err := ParseWithPolicyOptions("task://my-custom-slg/path-optional", Policy{
		DefaultNamespaceSegments: 1,
		SchemeNamespaceSegments: map[string]int{
			"task": 2,
		},
		VanityAliases: []VanityAlias{
			{
				From: "task://my-custom-slug/path-optional",
				To:   "task://hop-top/uri/T-0001",
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "task://hop-top/uri/T-0001", got.Canonical())
	assert.Equal(t, "task://my-custom-slg/path-optional", got.Vanity())
}

func TestParseWithPolicyOptions_StrictSkipsFuzzyVanity(t *testing.T) {
	_, err := ParseWithPolicyOptions("task://my-custom-slg/path-optional", Policy{
		DefaultNamespaceSegments: 1,
		SchemeNamespaceSegments: map[string]int{
			"task": 2,
		},
		VanityAliases: []VanityAlias{
			{
				From: "task://my-custom-slug/path-optional",
				To:   "task://hop-top/uri/T-0001",
			},
		},
	}, WithStrict())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestParseWithPolicyOptions_AmbiguousFuzzyVanityJSON(t *testing.T) {
	_, err := ParseWithPolicyOptions("task://shortcut", Policy{
		DefaultNamespaceSegments: 1,
		SchemeNamespaceSegments: map[string]int{
			"task": 2,
		},
		VanityAliases: []VanityAlias{
			{
				From: "task://shortcuta",
				To:   "task://hop-top/uri/T-0001",
			},
			{
				From: "task://shortcutb",
				To:   "task://hop-top/uri/T-0002",
			},
		},
	}, WithJSONAmbiguity())
	require.Error(t, err)

	var ambiguous AmbiguousVanityError
	require.True(t, errors.As(err, &ambiguous))
	assert.Len(t, ambiguous.Candidates, 2)
	assert.JSONEq(t, `{
		"input": "task://shortcut",
		"candidates": [
			{"from": "task://shortcuta", "to": "task://hop-top/uri/T-0001", "distance": 1},
			{"from": "task://shortcutb", "to": "task://hop-top/uri/T-0002", "distance": 1}
		]
	}`, err.Error())
}
