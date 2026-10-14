# Repository Agent Instructions

## Start here — always apply

Read [CONTEXT.md](CONTEXT.md) for the project's purpose, domain glossary, layered architecture, and invariants. Inspect `go.mod` (`go 1.26.1`), `Makefile`, relevant callers, and `git status --short --branch` before changing files.

Complete clear, authorized local work and verify it; ask only when missing information materially changes the outcome or an action requires additional authority. Preserve unrelated edits and staged content. A request to edit does not authorize a commit, push, or remote publication.

## Navigation — choose the task

Read the selected skill entrypoint in `.agents/skills/` and its mandatory boundaries. Load only the references needed for the current task, following prerequisites. Do not load every skill or reference by default.

| Task | Read / apply | Required context |
|---|---|---|
| Define or clarify product behavior and acceptance criteria | [requirements-definition](.agents/skills/requirements-definition/SKILL.md) | Ground outcomes, examples, and open product questions in repository facts |
| Prepare or revise a work plan | [plan](.agents/skills/plan/SKILL.md), then [plan-review](.agents/skills/plan-review/SKILL.md) | Investigate before asking questions; close material decisions; automatically review the completed plan before readiness |
| Review an existing plan | [plan-review](.agents/skills/plan-review/SKILL.md) | Establish baseline, sources, and review scope; review does not authorize fixes or execution |
| Split broad scope into deliverable slices and tasks | [task-decomposition](.agents/skills/task-decomposition/SKILL.md) | Map requirements to slices, dependencies, and scenario acceptance |
| Resolve an open architectural or technical design choice | [design-decision](.agents/skills/design-decision/SKILL.md) | Compare options against `CONTEXT.md` and existing package boundaries; design-only work does not authorize implementation |
| Start or continue repository edits; prepare commits or PRs | [git-delivery](.agents/skills/git-delivery/SKILL.md), relevant route | Use a verified task branch/worktree before the first edit; local commits and remote publication require separate requests |
| Write, fix, or refactor production Go code | [go-code](.agents/skills/go-code/SKILL.md), [code-change](.agents/skills/code-change/SKILL.md), [tdd-workflow](.agents/skills/tdd-workflow/SKILL.md), and [aaa-testing](.agents/skills/aaa-testing/SKILL.md) | Respect `go 1.26.1` in `go.mod`, thin `cmd/` facades, and `internal/app/app.go` interfaces; load TDD and AAA before executable edits |
| Implement or review UI, web, or landing-page code | [frontend-code](.agents/skills/frontend-code/SKILL.md), [code-change](.agents/skills/code-change/SKILL.md), [tdd-workflow](.agents/skills/tdd-workflow/SKILL.md), and [aaa-testing](.agents/skills/aaa-testing/SKILL.md) | Ground state, accessibility, responsive layout, and native web patterns in the selected frontend stack |
| Write, edit, or review automated tests | [aaa-testing](.agents/skills/aaa-testing/SKILL.md) ([Go reference](.agents/skills/aaa-testing/references/go-tests.md)), plus [tdd-workflow](.agents/skills/tdd-workflow/SKILL.md) for executable changes | Use Arrange–Act–Assert and `should <expected outcome> when <condition>` scenario titles for new or edited subtests/table cases |
| Investigate an unexplained failure or regression | [debugging](.agents/skills/debugging/SKILL.md) | Reproduce with a test or minimal probe and establish root cause; diagnosis alone does not authorize a fix |
| Review a diff, PR, or working changes | [code-review](.agents/skills/code-review/SKILL.md), plus relevant technical owners | Inspect actual diff layers and callers; rank findings with concrete evidence; review alone does not authorize fixes |
| Capture or retrieve lessons from mistakes and corrections | [learn-from-mistakes](.agents/skills/learn-from-mistakes/SKILL.md) | Record evidence-backed lessons and check prior pitfalls before similar work |
| Present technical explanations, reports, reviews, or plans | [developer-readable-output](.agents/skills/developer-readable-output/SKILL.md) | Lead with the answer or result, preserve required engineering evidence, and match the user's language and format |

## Verified build, test, and maintenance commands

Use the repository's `Makefile` targets and standard Go toolchain commands:

```bash
make build                # Build binary to tmp/fgm (injects buildVersion/buildCommit/buildDate via ldflags)
make test                 # Run all tests: go test ./...
make cover                # Run tests with coverage profile at tmp/coverage.out
make cover-html           # Generate HTML coverage report at tmp/coverage.html
make fmt                  # Format Go code: go fmt ./...
make fix                  # Run go fix ./... and go fmt ./... (repo-wide; may rewrite untouched files)
make tidy                 # Clean up go.mod and go.sum
make pre-commit           # Run fix, test, tidy, build (also executed by .hooky/hooks/pre-commit)
make update-lint-compat   # Regenerate internal/golangcilint/compatibility.json
```

Focused development and single-test execution:

```bash
go test ./internal/resolve/ -run TestResolverCurrent
go test -race ./...
go run . current --chdir .
make cmd CMD='current --chdir .'
```

## Project conventions and boundaries

- **Keep `cmd/` thin:** Put CLI flag parsing and output wiring in `cmd/`; place business logic in `internal/` packages and expose cross-package service contracts through interfaces in `internal/app/app.go` (note existing exceptions in `cmd/pin.go`, which calls `fgmconfig.SaveNearest` directly, and `cmd/shim.go`, which defines hidden `__shim` and invokes `shim.New` + `os/exec` directly).
- **Configuration and catalog ownership:** Repository pinning `.fgm.toml` is owned by `internal/fgmconfig/`; generated `.golangci.yml` and lint diagnostics are owned by `internal/lintconfig/`. Do not hand-edit `internal/golangcilint/compatibility.json` — update `scripts/update-lint-compatibility` if needed and run `make update-lint-compat`.
- **Release artifact contract:** `release_files_test.go` verifies `.goreleaser.yaml`, `.github/workflows/release.yml`, `install.sh`, and `Makefile` ldflags. Preserve those contracts when touching build or release configuration.
- **Protected paths:** Never overwrite or commit secrets into `.env*`, and never directly manipulate `.git/` internals.

## Validation and reporting

- **Executable Go or frontend changes:** Follow [tdd-workflow](.agents/skills/tdd-workflow/SKILL.md) (observed RED for new/changed behavior, minimal GREEN, then REFACTOR) and [aaa-testing](.agents/skills/aaa-testing/SKILL.md). Run targeted tests first, then `make test`, `go vet ./...`, and `make build` before claiming completion. Because `make fix` and `make pre-commit` run `go fix ./...` across the whole repository (which currently rewrites `cmd/shim.go`), check `git status` / `git diff` after running `make pre-commit` and revert any out-of-scope `go fix` edits unless modernizing that file is part of the task.
- **Documentation and configuration changes:** Inspect `git diff`, verify local Markdown links and heading anchors, run `git diff --check`, and run `make test` when build/release files are touched.
- Report changed files, checks actually executed, and any remaining verification gap. Do not claim a command passed unless it was run in the current session.
