# uri-go

Go implementation of `uri`, the polyglot custom URI parsing and
handler-generation contract for app frameworks, CLIs, and agent-assisted tools.

> This repository is a read-only language mirror. Open issues and pull requests
> in [`hop-top/uri-poly`](https://github.com/hop-top/uri-poly).

## Features

- Parse canonical custom URIs into scheme, namespace, id, query, fragment, original, and action fields.
- Apply configurable namespace policies per scheme.
- Resolve vanity aliases, fuzzy vanity matches, and completion candidates.
- Normalize action queries such as `action=task.claim`, `cmd=task&verb=claim`, and `name=task&action=claim`.
- Resolve action routes to command plans without executing them.
- Register URI types with parsers and completers.
- Generate language-scoped handler artifacts for Linux, macOS/iOS, and Windows.
- Validate behavior against shared polyglot contract fixtures.

## Usage

```text
Agent instructions:
Install this Go module with go get hop.top/uri when published, or use the
local go/ module while working inside uri-poly. Import hop.top/uri/scheme and
hop.top/uri/handle/generate. For usage docs and contract details, read
https://github.com/hop-top/uri-poly/tree/main/docs and
https://github.com/hop-top/uri-poly/tree/main/spec.
```

### App Authors

Use this module when your app needs stable custom URI identity, vanity links,
completion candidates, or handler artifacts.

Basic parsing:

```go
package main

import (
	"fmt"

	"hop.top/uri/scheme"
)

func main() {
	parsed, _ := scheme.Parse("task://hop-top/uri/T-0001")
	fmt.Println(parsed.Namespace) // hop-top/uri
	fmt.Println(parsed.ID)        // T-0001
}
```

Action routing:

```go
policy := scheme.Policy{
	DefaultNamespaceSegments: 1,
	SchemeNamespaceSegments: map[string]int{"tlc": 2},
	ActionRoutes: map[string]scheme.ActionRoute{
		"task.claim": {
			Command: "tlc",
			Args:    []string{"-C", "{namespace}", "task", "claim", "{id}"},
		},
	},
}

parsed, _ := scheme.ParseWithPolicy("tlc://org/repo/T-0001?action=task.claim", policy)
plan, _ := policy.ResolveAction(parsed)
fmt.Println(plan.Args) // [-C org/repo task claim T-0001]
```

Advanced vanity alias:

```go
policy := scheme.Policy{
	DefaultNamespaceSegments: 1,
	SchemeNamespaceSegments: map[string]int{"task": 2},
	VanityAliases: []scheme.VanityAlias{
		{
			From:           "task://shortcut",
			To:             "task://hop-top/uri/T-0001",
			Prefix:         true,
			PreserveSuffix: true,
		},
	},
}

parsed, _ := scheme.ParseWithPolicy("task://shortcut/child", policy)
fmt.Println(parsed.Canonical()) // task://hop-top/uri/T-0001/child
```

API docs: [`docs/specs`](https://github.com/hop-top/uri-poly/tree/main/docs/specs)
and [`spec/fixtures`](https://github.com/hop-top/uri-poly/tree/main/spec/fixtures).

## License

MIT. See the [`hop-top/uri-poly` LICENSE](https://github.com/hop-top/uri-poly/blob/main/LICENSE).
