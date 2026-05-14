package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"hop.top/uri/handle/generate"
	"hop.top/uri/scheme"
)

type uriFixture struct {
	Version         int `json:"version"`
	NamespacePolicy struct {
		DefaultNamespaceSegments int            `json:"defaultNamespaceSegments"`
		SchemeNamespaceSegments  map[string]int `json:"schemeNamespaceSegments"`
	} `json:"namespacePolicy"`
	ValidCases   []simpleURICase `json:"validCases"`
	InvalidCases []simpleURICase `json:"invalidCases"`
	VanityCases []struct {
		Name    string               `json:"name"`
		Input   string               `json:"input"`
		Aliases []scheme.VanityAlias `json:"aliases"`
		Options struct {
			Strict        bool `json:"strict"`
			JSONAmbiguity bool `json:"jsonAmbiguity"`
		} `json:"options"`
	} `json:"vanityCases"`
	ActionCases []struct {
		Name         string                        `json:"name"`
		Input        string                        `json:"input"`
		ActionRoutes map[string]scheme.ActionRoute `json:"actionRoutes"`
	} `json:"actionCases"`
}

type simpleURICase struct {
	Name  string `json:"name"`
	Input string `json:"input"`
}

type handlerFixture struct {
	Version int `json:"version"`
	Cases []struct {
		Name     string               `json:"name"`
		Platform string               `json:"platform"`
		Spec     generate.HandlerSpec `json:"spec"`
	} `json:"cases"`
	InvalidCases []struct {
		Name string               `json:"name"`
		Spec generate.HandlerSpec `json:"spec"`
	} `json:"invalidCases"`
}

type uriOut struct {
	Scheme    string `json:"scheme"`
	Namespace string `json:"namespace"`
	ID        string `json:"id"`
	Query     string `json:"query"`
	Fragment  string `json:"fragment"`
	Original  string `json:"original"`
	Action    string `json:"action"`
	Canonical string `json:"canonical"`
	Vanity    string `json:"vanity"`
}

type resolvedActionOut struct {
	Action  string   `json:"action"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type uriCaseOut struct {
	Name   string             `json:"name"`
	OK     bool               `json:"ok"`
	URI    *uriOut            `json:"uri,omitempty"`
	Action *resolvedActionOut `json:"action,omitempty"`
}

type handlerCaseOut struct {
	Name            string `json:"name"`
	OK              bool   `json:"ok"`
	HandlerID       string `json:"handlerId"`
	DesktopFilename string `json:"desktopFilename"`
	Rendered        string `json:"rendered"`
}

func main() {
	fixtures := flag.String("fixtures", "../spec/fixtures", "fixture directory")
	flag.Parse()

	uriFx := mustReadJSON[uriFixture](filepath.Join(*fixtures, "uri-contract.json"))
	handlerFx := mustReadJSON[handlerFixture](filepath.Join(*fixtures, "handler-contract.json"))
	policy := scheme.Policy{
		DefaultNamespaceSegments: uriFx.NamespacePolicy.DefaultNamespaceSegments,
		SchemeNamespaceSegments:  uriFx.NamespacePolicy.SchemeNamespaceSegments,
	}

	out := map[string]any{
		"version": 1,
		"uri": map[string]any{
			"valid":   emitURIValid(uriFx.ValidCases, policy),
			"invalid": emitURIInvalid(uriFx.InvalidCases, policy),
			"vanity":  emitURIVanity(uriFx.VanityCases, policy),
			"action":  emitURIAction(uriFx.ActionCases, policy),
		},
		"handler": map[string]any{
			"valid":   emitHandlerValid(handlerFx.Cases),
			"invalid": emitHandlerInvalid(handlerFx.InvalidCases),
		},
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		panic(err)
	}
}

func mustReadJSON[T any](path string) T {
	var out T
	raw, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		panic(fmt.Errorf("%s: %w", path, err))
	}
	return out
}

func emitURIValid(cases []simpleURICase, policy scheme.Policy) []uriCaseOut {
	out := make([]uriCaseOut, 0, len(cases))
	for _, tc := range cases {
		out = append(out, parseCase(tc.Name, tc.Input, policy))
	}
	return out
}

func emitURIInvalid(cases []simpleURICase, policy scheme.Policy) []uriCaseOut {
	out := make([]uriCaseOut, 0, len(cases))
	for _, tc := range cases {
		got := parseCase(tc.Name, tc.Input, policy)
		got.OK = got.OK == false
		got.URI = nil
		out = append(out, got)
	}
	return out
}

func emitURIVanity(cases []struct {
	Name    string               `json:"name"`
	Input   string               `json:"input"`
	Aliases []scheme.VanityAlias `json:"aliases"`
	Options struct {
		Strict        bool `json:"strict"`
		JSONAmbiguity bool `json:"jsonAmbiguity"`
	} `json:"options"`
}, base scheme.Policy) []uriCaseOut {
	out := make([]uriCaseOut, 0, len(cases))
	for _, tc := range cases {
		policy := base
		policy.VanityAliases = tc.Aliases
		opts := make([]scheme.ParseOption, 0, 2)
		if tc.Options.Strict {
			opts = append(opts, scheme.WithStrict())
		}
		if tc.Options.JSONAmbiguity {
			opts = append(opts, scheme.WithJSONAmbiguity())
		}
		uri, err := scheme.ParseWithPolicyOptions(tc.Input, policy, opts...)
		out = append(out, uriResult(tc.Name, uri, err))
	}
	return out
}

func emitURIAction(cases []struct {
	Name         string                        `json:"name"`
	Input        string                        `json:"input"`
	ActionRoutes map[string]scheme.ActionRoute `json:"actionRoutes"`
}, base scheme.Policy) []uriCaseOut {
	out := make([]uriCaseOut, 0, len(cases))
	for _, tc := range cases {
		policy := base
		policy.ActionRoutes = tc.ActionRoutes
		uri, err := scheme.ParseWithPolicy(tc.Input, policy)
		result := uriResult(tc.Name, uri, err)
		if err == nil {
			plan, planErr := policy.ResolveAction(uri)
			if planErr != nil {
				result.OK = false
				result.URI = nil
			} else {
				result.Action = &resolvedActionOut{Action: plan.Action, Command: plan.Command, Args: plan.Args}
			}
		}
		out = append(out, result)
	}
	return out
}

func parseCase(name, input string, policy scheme.Policy) uriCaseOut {
	uri, err := scheme.ParseWithPolicy(input, policy)
	return uriResult(name, uri, err)
}

func uriResult(name string, uri *scheme.URI, err error) uriCaseOut {
	if err != nil {
		return uriCaseOut{Name: name, OK: false}
	}
	return uriCaseOut{Name: name, OK: true, URI: &uriOut{
		Scheme: uri.Scheme, Namespace: uri.Namespace, ID: uri.ID, Query: uri.Query,
		Fragment: uri.Fragment, Original: uri.Original, Action: uri.Action,
		Canonical: uri.Canonical(), Vanity: uri.Vanity(),
	}}
}

func emitHandlerValid(cases []struct {
	Name     string               `json:"name"`
	Platform string               `json:"platform"`
	Spec     generate.HandlerSpec `json:"spec"`
}) []handlerCaseOut {
	out := make([]handlerCaseOut, 0, len(cases))
	for _, tc := range cases {
		id, idErr := tc.Spec.HandlerID()
		rendered, renderErr := generate.Snippet(tc.Platform, tc.Spec)
		desktop := ""
		if tc.Platform == "linux" {
			desktop, _ = generate.DesktopFilename(tc.Spec)
		}
		if idErr != nil || renderErr != nil {
			out = append(out, handlerCaseOut{Name: tc.Name, OK: false})
			continue
		}
		out = append(out, handlerCaseOut{Name: tc.Name, OK: true, HandlerID: id, DesktopFilename: desktop, Rendered: rendered})
	}
	return out
}

func emitHandlerInvalid(cases []struct {
	Name string               `json:"name"`
	Spec generate.HandlerSpec `json:"spec"`
}) []handlerCaseOut {
	out := make([]handlerCaseOut, 0, len(cases))
	for _, tc := range cases {
		err := tc.Spec.Validate()
		out = append(out, handlerCaseOut{Name: tc.Name, OK: err != nil})
	}
	return out
}
