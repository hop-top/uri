package completions_test

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"hop.top/uri/scheme"
	"hop.top/uri/scheme/completions"
)

func TestCobraCompleter_Complete(t *testing.T) {
	reg := scheme.NewRegistry()
	_ = reg.Register(scheme.TypeRegistration{
		Name: "task",
		Completer: func(_ context.Context, prefix string) ([]string, error) {
			all := []string{"T-0001", "T-0002", "T-0099"}
			var out []string
			for _, id := range all {
				if strings.HasPrefix(id, prefix) {
					out = append(out, id)
				}
			}
			return out, nil
		},
	})

	c := completions.NewCobraCompleter(reg)
	fn := c.Complete("task")

	got, directive := fn(&cobra.Command{}, nil, "T-000")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive: %v", directive)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 suggestions, got %d: %v", len(got), got)
	}
}

func TestCobraCompleter_Complete_URIScheme(t *testing.T) {
	reg := scheme.NewRegistry()
	_ = reg.Register(scheme.TypeRegistration{
		Name: "task",
		Completer: func(_ context.Context, _ string) ([]string, error) {
			return []string{"T-0001"}, nil
		},
	})

	c := completions.NewCobraCompleter(reg)
	fn := c.Complete("task")

	got, _ := fn(&cobra.Command{}, nil, "task://T-")
	if len(got) != 1 || got[0] != "task://T-0001" {
		t.Errorf("expected task://T-0001, got %v", got)
	}
}

func TestCobraCompleter_Complete_FuzzyVanityCandidates(t *testing.T) {
	reg := scheme.NewRegistryWithPolicy(scheme.Policy{
		VanityAliases: []scheme.VanityAlias{
			{From: "task://shortcuta", To: "task://hop-top/uri/T-0001"},
			{From: "task://shortcutb", To: "task://hop-top/uri/T-0002"},
		},
	})
	_ = reg.Register(scheme.TypeRegistration{
		Name: "task",
		Completer: func(_ context.Context, _ string) ([]string, error) {
			return nil, nil
		},
	})

	c := completions.NewCobraCompleter(reg)
	fn := c.Complete("task")

	got, directive := fn(&cobra.Command{}, nil, "task://shortcut")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive: %v", directive)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 fuzzy vanity suggestions, got %d: %v", len(got), got)
	}
	if !strings.Contains(got[0], "task://shortcuta") || !strings.Contains(got[0], "canonical: task://hop-top/uri/T-0001") {
		t.Fatalf("unexpected first suggestion: %q", got[0])
	}
}
