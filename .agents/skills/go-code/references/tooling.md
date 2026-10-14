# Verification and Toolchain Compatibility

Discover the target's commands in applicable project instructions, scripts, CI, `go.mod`/`go.work`, build tags, and lint configuration. Run commands from the intended module/workspace root. Do not assume one CI filename, PostgreSQL, Docker, a `.env` file, or a linter configuration exists.

## Verification choices

While iterating, run focused behavior tests for the changed path. Format changed Go files using the project's formatter. Then run applicable broader checks, typically:

```bash
go build ./...
go vet ./...
go test ./...
```

Use the target's tags and environment. Multi-module repositories may need checks in each affected module. For affected concurrent code, use `go test -race` on relevant packages where the platform and toolchain support it; follow CI for wider runs. `-count=1` forces fresh execution when useful; ordinary cached results are not inherently stale or invalid.

`gofmt -l` and `goimports -l` list files needing changes; success exit status alone does not prove formatting passed. Check their output, or use the repository's enforcement script. `goimports` is optional tooling, not a reason to install an unpinned latest release. Avoid formatting generated/vendor files or unrelated code accidentally.

Use configured lint/static checks, such as `golangci-lint run ./...`, when available. For dependency changes or security work, run the project's pinned vulnerability check, such as `govulncheck ./...`. Distinguish a vulnerability scanner's result from proof that the code is secure.

If a required tool or service is unavailable, run independent checks that remain useful and report the exact gap. Follow the project's authorized fixture setup; do not source arbitrary environment files, print secrets, or mutate a live database just to make a generic checklist pass.

## Version failures

- The module's `go` directive establishes its minimum baseline; the `toolchain` directive and local binary may select a newer compiler. See [modern-go.md](modern-go.md) for language/API gates.
- Inspect CI pins and local tool versions before attributing a failure. golangci-lint needs both a suitable build toolchain and analyzer support for the target Go release.
- If formatters disagree, reproduce on the changed file and compare versions/configuration before updating tools. Do not infer that an older binary is the only possible cause.
- For an authorized baseline update, check build, tests, lint, formatter, and vulnerability-tool compatibility together. Update only pins that need changing; do not fabricate CI policy.
- `GOTOOLCHAIN=go1.X.Y go test ./...` can select a released toolchain, but may download it. Use actual supported version numbers and existing cached toolchains when possible; respect network and installation constraints.

## Handling findings

Fix the underlying issue and verify the reported path. Review resource leaks, error handling, and race findings as behavior problems; structural/style metrics are signals unless the project makes them gates.

Do not enable a new linter set, replace config, or weaken a rule as incidental cleanup. If a false positive needs suppression, follow the project's narrow suppression syntax and document the reason. Check callers before deleting exports; absence of local callers does not prove a public API is unused by downstream consumers.

Report checks actually run and their results, including skipped prerequisites and supported-version coverage. Do not claim the full suite passed from a focused test or from `go vet` alone.

Sources: [Go command documentation](https://pkg.go.dev/cmd/go), [Go toolchains](https://go.dev/doc/toolchain), [golangci-lint FAQ](https://golangci-lint.run/docs/welcome/faq/), [govulncheck](https://go.dev/doc/security/vuln/).
