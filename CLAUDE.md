# GoBlog — Claude Instructions

## Documentation

All API changes (new exported types, functions, options, flags, or fields) must include corresponding documentation updates: godoc comments on the new symbols, relevant sections in `README.md`, and any package-level `doc.go` entries that reference available options or features.

The README.md does not need to be flooded with documentation. Just the relevant information for a user to get started with the 3 methods to consume the library:
1. CLI tool via go install
2. Docker image via docker pull/run
3. The library itself

## Config Pattern

All configurable values are defined as named types in `pkg/config` (e.g. `type Port int`, `type WatcherDebounce struct{ Debounce time.Duration }`). These types are **embedded directly** into the structs that consume them (`Server`, `Generator`, `Watcher`), not stored as plain unexported fields. Option constructor functions (e.g. `WithPort`, `WithLogger`) live in `pkg/config` and return an option struct with exactly one non-nil function pointer. Options are applied in an `else if` chain passing a pointer to the embedded field: `if opt.WithXFunc != nil { opt.WithXFunc(&s.X) } else if opt.WithYFunc != nil { ... }`. New configurable behaviour always follows this pattern end-to-end: config type in `pkg/config`, option function in `pkg/config`, embedded field in the consumer struct.
