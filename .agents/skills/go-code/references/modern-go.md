# Version-Aware Go

Use modern APIs when they improve the current task and fit the project's supported versions. Existing valid code does not need a mechanical rewrite merely because another spelling is newer.

## Establish availability

Read the `go` directive, `toolchain` directive, CI matrix, and relevant build tags. The `go` directive sets the minimum Go version and language baseline; `toolchain` selects a preferred toolchain. A newer compiler does not make later APIs portable to older supported toolchains. Per-file build constraints can affect language version, so inspect them when code intentionally supports multiple releases.

For an unfamiliar API, check the official versioned release notes and signature, then compile a minimal example under the intended version. `go doc` and `GOROOT/api/go1.N.txt` show what the installed toolchain contains; they do not prove older-version compatibility. Do not repeat inherited claims that examples were compiled unless you actually ran them.

## Useful feature gates

The following introduction versions were checked against official release notes on 2026-10-09. This is a task-oriented selection, not an exhaustive API inventory.

| Minimum Go | Features | Decision notes |
|---|---|---|
| [1.18](https://go.dev/doc/go1.18) | `any`, generics, `strings.Cut` | Keep constraints meaningful; generic code is optional. |
| [1.20](https://go.dev/doc/go1.20) | `errors.Join`, `ReverseProxy.Rewrite` | Rewriting proxy configuration changes header handling; test intended forwarding. |
| [1.21](https://go.dev/doc/go1.21) | `slices`, `maps`, `cmp`, `slog`, `min/max/clear`, `sync.OnceValue`, `context.WithoutCancel` | `WithoutCancel` removes deadline and cancellation; detached work needs its own lifetime. |
| [1.22](https://go.dev/doc/go1.22) | Integer range, per-iteration declared loop variables, ServeMux method/path patterns, `cmp.Or`, `math/rand/v2`, `sql.Null[T]` | Loop variables assigned with `=` remain shared; `cmp.Or` treats every zero as absent. |
| [1.23](https://go.dev/doc/go1.23) | Range over functions, `iter`, `maps.Keys` iterators, `slices.Sorted`, `go vet`'s `stdversion` | With `iter.Pull`, call `stop` when abandoning iteration. |
| [1.24](https://go.dev/doc/go1.24) | Generic aliases, `strings/bytes.SplitSeq`, `T.Context`, `B.Loop`, JSON `omitzero`, `os.Root`, `tool` directive | `os.Root` constrains traversal but is not a complete filesystem sandbox. |
| [1.25](https://go.dev/doc/go1.25) | `WaitGroup.Go`, stable `testing/synctest.Test/Wait`, `reflect.TypeAssert` | A `WaitGroup.Go` callback must not panic; see [data-and-io.md](data-and-io.md). |
| [1.26](https://go.dev/doc/go1.26) | `new(expr)`, `errors.AsType`, `slog.NewMultiHandler`, modernized `go fix` | `errors.AsType[T](err)` returns `(T, bool)`; preview modernization before applying it. |
| [1.27](https://go.dev/doc/go1.27) | Generic methods, stable `encoding/json/v2`, stdlib `uuid`, `strings/bytes.CutLast`, `synctest.Sleep` | Introducing these requires a compatible baseline; JSON/UUID migrations are separate contract changes. |

## Preserve semantics when modernizing

- `slices.SortFunc` uses a three-way comparator; retain stable sorting when equal elements must keep their order. Map key collection does not imply sorted order.
- `slices.Clone` and `maps.Clone` are shallow. Slices mutation helpers may alter the backing array and clear vacated elements; check aliases and callers before replacing manual code.
- Range-over-function and sequence APIs can avoid intermediate collections, but short-circuit behavior and ownership still matter. Prefer the clearest suitable form rather than replacing every loop.
- `math/rand/v2` is for non-security randomness. Use `crypto/rand` for secrets; do not change deterministic test sequences without updating their intended contract.
- `os.Root` prevents traversal outside its root through supported operations, including symlink escapes. It does not restrict access through mount points or device files, and is not a process sandbox.
- Modern features in tests follow the same version gates. `synctest` does not virtualize external databases or arbitrary network I/O; inspect its restrictions before adopting it.

## JSON is a wire-contract decision

`encoding/json` retains v1 semantics. Stable `encoding/json/v2` is available from Go 1.27; earlier experimental availability is not a production baseline.

By default, v2 rejects duplicate object names and invalid UTF-8, matches field names case-sensitively, emits empty arrays/objects for nil slices/maps, and requires an explicit representation for `time.Duration`. Map output is not deterministic by default; use the documented option where required. Check the current API for compatibility options and custom encoders; do not copy experimental-era tags into stable v2.

Keep existing parsing and serialization unless a migration is requested. Test the affected field names, omission/null behavior, duplicate keys, malformed input, durations, and ordering before switching packages. The v1 `omitzero` option introduced in Go 1.24 is itself a contract change, not an automatic correction for every `omitempty` tag.

Sources: [JSON v2 documentation](https://pkg.go.dev/encoding/json/v2), [Go 1.27 release notes](https://go.dev/doc/go1.27).

## Language, library, and compatibility behavior differ

The Go 1.22 per-iteration rule applies to variables declared by the loop in code with the corresponding language version. Library changes such as slice tail clearing are not all gated by that rule.

Go 1.23 introduced collectable timers and synchronous timer channels with defaults affected by the main module's `go` line and `GODEBUG` on releases supporting `asynctimerchan`. Go 1.27 removed that setting and timer channels are always synchronous. Do not depend on `len(t.C)` or transfer old Stop/Reset recipes between releases without checking semantics.

Use `go vet` to catch applicable version mistakes, and build/test on the supported toolchains for stronger evidence. On Go 1.26+, `go fix -diff ./...` previews modernization; review its output before an authorized application. Updating the baseline requires checking CI and tool compatibility rather than assuming every pin must change.

Sources: [Go toolchains](https://go.dev/doc/toolchain), [language specification](https://go.dev/ref/spec), [GODEBUG compatibility](https://go.dev/doc/godebug), [Go 1.23 timers](https://go.dev/wiki/Go123Timer), [os.Root](https://pkg.go.dev/os#Root).
