package scheme

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// URI represents a canonical custom URI identifier.
type URI struct {
	Scheme    string
	Namespace string
	ID        string
	Query     string
	Fragment  string
	Original  string
	Action    string
}

// Policy controls how leading URI path segments are assigned to Namespace.
type Policy struct {
	DefaultNamespaceSegments int
	SchemeNamespaceSegments  map[string]int
	VanityAliases            []VanityAlias
	ActionRoutes             map[string]ActionRoute
}

// VanityAlias maps a shorter URI to its canonical URI.
type VanityAlias struct {
	From           string
	To             string
	Prefix         bool
	PreserveSuffix bool
}

// ActionRoute maps a parsed URI action to a command argv template.
type ActionRoute struct {
	Command string
	Args    []string
}

// ResolvedAction is a command plan derived from a URI action route.
type ResolvedAction struct {
	Action  string
	Command string
	Args    []string
}

// ParseOptions controls optional parser behavior.
type ParseOptions struct {
	Strict        bool
	JSONAmbiguity bool
}

// ParseOption configures parsing behavior.
type ParseOption func(*ParseOptions)

// WithStrict disables fuzzy vanity alias matching.
func WithStrict() ParseOption {
	return func(options *ParseOptions) {
		options.Strict = true
	}
}

// WithJSONAmbiguity returns ambiguous fuzzy vanity matches as a JSON error.
func WithJSONAmbiguity() ParseOption {
	return func(options *ParseOptions) {
		options.JSONAmbiguity = true
	}
}

// VanityCandidate is one possible fuzzy vanity alias match.
type VanityCandidate struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Distance int    `json:"distance"`
}

// AmbiguousVanityError reports multiple equally close vanity aliases.
type AmbiguousVanityError struct {
	Input      string            `json:"input"`
	Candidates []VanityCandidate `json:"candidates"`
	AsJSON     bool              `json:"-"`
}

func (e AmbiguousVanityError) Error() string {
	if e.AsJSON {
		raw, err := json.Marshal(e)
		if err == nil {
			return string(raw)
		}
	}
	parts := make([]string, 0, len(e.Candidates))
	for _, candidate := range e.Candidates {
		parts = append(parts, candidate.From)
	}
	return fmt.Sprintf("uri: ambiguous vanity alias %q: %s", e.Input, strings.Join(parts, ", "))
}

// DefaultPolicy matches the shared URI contract fixture.
var DefaultPolicy = Policy{
	DefaultNamespaceSegments: 1,
	SchemeNamespaceSegments: map[string]int{
		"task":        2,
		"doc":         2,
		"repo":        1,
		"tlc":         2,
		"task-dev":    2,
		"task-stress": 2,
	},
}

// Parse converts input into a URI using DefaultPolicy.
func Parse(input string) (*URI, error) {
	return ParseWithPolicy(input, DefaultPolicy)
}

// ParseWithPolicy converts input into a URI using the provided namespace policy.
func ParseWithPolicy(input string, policy Policy) (*URI, error) {
	return ParseWithPolicyOptions(input, policy)
}

// ParseWithPolicyOptions converts input into a URI using the provided namespace policy and options.
func ParseWithPolicyOptions(input string, policy Policy, opts ...ParseOption) (*URI, error) {
	if input == "" {
		return nil, fmt.Errorf("uri: empty input")
	}
	options := ParseOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	parseInput, vanity, err := policy.resolveVanity(input, options)
	if err != nil {
		return nil, err
	}

	parsed, err := url.Parse(parseInput)
	if err != nil {
		return nil, fmt.Errorf("uri: invalid input: %w", err)
	}
	if parsed.Scheme == "" {
		return nil, fmt.Errorf("uri: scheme is required")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("uri: namespace is required")
	}

	segments := []string{parsed.Host}
	for _, segment := range strings.Split(strings.TrimPrefix(parsed.EscapedPath(), "/"), "/") {
		if segment != "" {
			decoded, err := url.PathUnescape(segment)
			if err != nil {
				return nil, fmt.Errorf("uri: invalid path segment: %w", err)
			}
			segments = append(segments, decoded)
		}
	}

	namespaceSegments := policy.namespaceSegments(parsed.Scheme)
	if namespaceSegments <= 0 {
		return nil, fmt.Errorf("uri: namespace segment count must be positive")
	}
	if len(segments) <= namespaceSegments {
		return nil, fmt.Errorf("uri: id is required")
	}

	namespace := strings.Join(segments[:namespaceSegments], "/")
	id := strings.Join(segments[namespaceSegments:], "/")
	if namespace == "" {
		return nil, fmt.Errorf("uri: namespace is required")
	}
	if id == "" {
		return nil, fmt.Errorf("uri: id is required")
	}

	action, err := actionFromQuery(parsed.RawQuery)
	if err != nil {
		return nil, err
	}

	return &URI{
		Scheme:    parsed.Scheme,
		Namespace: namespace,
		ID:        id,
		Query:     parsed.RawQuery,
		Fragment:  parsed.Fragment,
		Original:  vanity,
		Action:    action,
	}, nil
}

func actionFromQuery(rawQuery string) (string, error) {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", fmt.Errorf("uri: invalid query: %w", err)
	}

	candidates := make([]string, 0, 3)
	if action := values.Get("action"); action != "" && values.Get("name") == "" {
		candidates = append(candidates, action)
	}
	if cmd, verb := values.Get("cmd"), values.Get("verb"); cmd != "" || verb != "" {
		if cmd == "" || verb == "" {
			return "", fmt.Errorf("uri: cmd and verb must be provided together")
		}
		candidates = append(candidates, cmd+"."+verb)
	}
	if name, action := values.Get("name"), values.Get("action"); name != "" {
		if action == "" {
			return "", fmt.Errorf("uri: name and action must be provided together")
		}
		candidates = append(candidates, name+"."+action)
	}

	if len(candidates) == 0 {
		return "", nil
	}
	for _, candidate := range candidates[1:] {
		if candidate != candidates[0] {
			return "", fmt.Errorf("uri: conflicting action query parameters")
		}
	}
	return candidates[0], nil
}

// ResolveAction turns a parsed URI action into a command plan. It never executes.
func (p Policy) ResolveAction(uri *URI) (*ResolvedAction, error) {
	if uri == nil {
		return nil, fmt.Errorf("uri: nil URI")
	}
	if uri.Action == "" {
		return nil, fmt.Errorf("uri: action is required")
	}
	route, ok := p.ActionRoutes[uri.Action]
	if !ok {
		return nil, fmt.Errorf("uri: unknown action %q", uri.Action)
	}
	if route.Command == "" {
		return nil, fmt.Errorf("uri: action route command is required")
	}

	args := make([]string, len(route.Args))
	for i, arg := range route.Args {
		args[i] = expandActionTemplate(arg, uri)
	}
	return &ResolvedAction{
		Action:  uri.Action,
		Command: expandActionTemplate(route.Command, uri),
		Args:    args,
	}, nil
}

func expandActionTemplate(value string, uri *URI) string {
	replacements := map[string]string{
		"{scheme}":    uri.Scheme,
		"{namespace}": uri.Namespace,
		"{id}":        uri.ID,
		"{query}":     uri.Query,
		"{fragment}":  uri.Fragment,
	}
	for from, to := range replacements {
		value = strings.ReplaceAll(value, from, to)
	}
	return value
}

func (p Policy) namespaceSegments(scheme string) int {
	if p.SchemeNamespaceSegments != nil {
		if segments, ok := p.SchemeNamespaceSegments[scheme]; ok {
			return segments
		}
	}
	if p.DefaultNamespaceSegments != 0 {
		return p.DefaultNamespaceSegments
	}
	return 1
}

func (p Policy) resolveVanity(input string, options ParseOptions) (parseInput string, vanity string, err error) {
	best := VanityAlias{}
	bestLen := -1
	for _, candidate := range p.VanityAliases {
		if candidate.From == "" || candidate.To == "" {
			return "", "", fmt.Errorf("uri: vanity alias from and to are required")
		}
		if strings.Contains(candidate.To, "://") {
			if _, parseErr := url.Parse(candidate.To); parseErr != nil {
				return "", "", fmt.Errorf("uri: invalid vanity target: %w", parseErr)
			}
		}

		matched := input == candidate.From
		if !matched && candidate.Prefix {
			matched = strings.HasPrefix(input, candidate.From+"/")
		}
		if matched && len(candidate.From) > bestLen {
			best = candidate
			bestLen = len(candidate.From)
		}
	}
	if bestLen == -1 {
		if !options.Strict {
			if fuzzy, ok, err := p.closestVanity(input, options); err != nil || ok {
				return fuzzy.parseInput, fuzzy.vanity, err
			}
		}
		return input, "", nil
	}

	target := best.To
	if best.Prefix && best.PreserveSuffix && len(input) > len(best.From) {
		target = strings.TrimRight(target, "/") + input[len(best.From):]
	}
	return target, input, nil
}

type resolvedVanity struct {
	parseInput string
	vanity     string
}

func (p Policy) closestVanity(input string, options ParseOptions) (resolvedVanity, bool, error) {
	candidates := p.VanityCandidates(input)
	if len(candidates) == 0 {
		return resolvedVanity{}, false, nil
	}

	bestDistance := candidates[0].Distance
	best := []VanityCandidate{candidates[0]}
	for _, candidate := range candidates[1:] {
		if candidate.Distance != bestDistance {
			break
		}
		best = append(best, candidate)
	}

	if len(best) > 1 {
		return resolvedVanity{}, true, AmbiguousVanityError{
			Input:      input,
			Candidates: best,
			AsJSON:     options.JSONAmbiguity,
		}
	}
	return resolvedVanity{parseInput: best[0].To, vanity: input}, true, nil
}

// VanityCandidates returns fuzzy vanity alias candidates sorted by distance.
func (p Policy) VanityCandidates(input string) []VanityCandidate {
	candidates := make([]VanityCandidate, 0)
	for _, candidate := range p.VanityAliases {
		distance := levenshtein(input, candidate.From)
		if !withinFuzzyThreshold(input, candidate.From, distance) {
			continue
		}
		candidates = append(candidates, VanityCandidate{
			From:     candidate.From,
			To:       candidate.To,
			Distance: distance,
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Distance != candidates[j].Distance {
			return candidates[i].Distance < candidates[j].Distance
		}
		return candidates[i].From < candidates[j].From
	})
	return candidates
}

func withinFuzzyThreshold(input, candidate string, distance int) bool {
	longest := len(input)
	if len(candidate) > longest {
		longest = len(candidate)
	}
	threshold := longest / 5
	if threshold < 2 {
		threshold = 2
	}
	if threshold > 8 {
		threshold = 8
	}
	return distance <= threshold
}

func levenshtein(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}

	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i, ac := range ar {
		curr[0] = i + 1
		for j, bc := range br {
			cost := 0
			if ac != bc {
				cost = 1
			}
			curr[j+1] = min3(curr[j]+1, prev[j+1]+1, prev[j]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(br)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Canonical returns the canonical string representation.
func (u URI) Canonical() string {
	out := fmt.Sprintf("%s://%s/%s", u.Scheme, u.Namespace, u.ID)
	if u.Query != "" {
		out += "?" + u.Query
	}
	if u.Fragment != "" {
		out += "#" + u.Fragment
	}
	return out
}

// Vanity returns the URI using the alias scheme when the input used one.
func (u URI) Vanity() string {
	if u.Original == "" {
		return u.Canonical()
	}
	return u.Original
}

// String returns the canonical string representation.
func (u URI) String() string {
	return u.Canonical()
}
