package completions

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"hop.top/uri/scheme"
)

// CobraCompleter provides Cobra-compatible completion logic for URI types.
type CobraCompleter struct {
	Registry *scheme.Registry
}

// NewCobraCompleter creates a new Cobra completion helper.
func NewCobraCompleter(r *scheme.Registry) *CobraCompleter {
	return &CobraCompleter{Registry: r}
}

// Complete returns a ValidArgsFunction for the given URI type.
func (c *CobraCompleter) Complete(typeName string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		prefix := toComplete
		scheme := ""
		if strings.Contains(toComplete, "://") {
			parts := strings.SplitN(toComplete, "://", 2)
			scheme = parts[0]
			prefix = parts[1]
		}

		if scheme != "" {
			candidates := c.Registry.CompleteVanity(toComplete)
			if len(candidates) > 1 {
				suggestions := make([]string, 0, len(candidates))
				for _, candidate := range candidates {
					suggestions = append(suggestions, candidate.From+"\tcanonical: "+candidate.To)
				}
				return suggestions, cobra.ShellCompDirectiveNoFileComp
			}
		}

		suggestions, err := c.Registry.Complete(context.Background(), typeName, prefix)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		if scheme != "" {
			for i, s := range suggestions {
				suggestions[i] = scheme + "://" + s
			}
		}

		return suggestions, cobra.ShellCompDirectiveNoFileComp
	}
}
