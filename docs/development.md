# Development

[Home](../README.md) · [Documentation](README.md)

Requires the Go version in [go.mod](../go.mod). Start with the repository's [architecture context](../CONTEXT.md) and [agent instructions](../AGENTS.md).

```bash
make build        # Build binary to tmp/fgm
make test         # Run all tests
make fmt          # Format Go files
make pre-commit   # Run fix, test, tidy, build
make cover-html   # Generate HTML coverage report
```

Run directly:

```bash
go run . --help
go run . current --chdir .
```

`make pre-commit` runs `go fix`, formatting, tests, module tidy, and the build across the repository. Review the resulting diff for unrelated changes. For checks that do not rewrite source or module files, use `make test`, `go vet ./...`, and `make build`.

The lint compatibility catalog is generated; update it with `make update-lint-compat` rather than editing the JSON directly.

Report bugs and request features through [GitHub Issues](https://github.com/kostikovk/fgm/issues). For changes, keep pull requests focused and include verification results.
