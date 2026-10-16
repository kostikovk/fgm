package cmd

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/kostikovk/fgm/internal/app"
)

func TestUIStreamCapabilityDetection(t *testing.T) {
	t.Parallel()

	ttyOut := new(bytes.Buffer)
	ttyErr := new(bytes.Buffer)
	pipeOut := new(bytes.Buffer)

	cases := []struct {
		name          string
		out           io.Writer
		err           io.Writer
		env           map[string]string
		wantOutStyled bool
		wantErrStyled bool
	}{
		{
			name:          "should enable styled mode on both streams when both stdout and stderr are TTY with standard environment",
			out:           ttyOut,
			err:           ttyErr,
			env:           map[string]string{"TERM": "xterm-256color"},
			wantOutStyled: true,
			wantErrStyled: true,
		},
		{
			name:          "should disable styled mode on stdout and keep stderr styled when stdout is piped and stderr is a TTY",
			out:           pipeOut,
			err:           ttyErr,
			env:           map[string]string{"TERM": "xterm-256color"},
			wantOutStyled: false,
			wantErrStyled: true,
		},
		{
			name:          "should disable styled mode on stderr and keep stdout styled when stderr is redirected and stdout is a TTY",
			out:           ttyOut,
			err:           pipeOut,
			env:           map[string]string{"TERM": "xterm-256color"},
			wantOutStyled: true,
			wantErrStyled: false,
		},
		{
			name:          "should disable styled mode on both streams when NO_COLOR is non-empty",
			out:           ttyOut,
			err:           ttyErr,
			env:           map[string]string{"TERM": "xterm-256color", "NO_COLOR": "1"},
			wantOutStyled: false,
			wantErrStyled: false,
		},
		{
			name:          "should disable styled mode on both streams when NO_COLOR has arbitrary non-empty value",
			out:           ttyOut,
			err:           ttyErr,
			env:           map[string]string{"TERM": "xterm-256color", "NO_COLOR": "yes"},
			wantOutStyled: false,
			wantErrStyled: false,
		},
		{
			name:          "should disable styled mode on both streams when CI is non-empty",
			out:           ttyOut,
			err:           ttyErr,
			env:           map[string]string{"TERM": "xterm-256color", "CI": "true"},
			wantOutStyled: false,
			wantErrStyled: false,
		},
		{
			name:          "should disable styled mode on both streams when TERM is dumb",
			out:           ttyOut,
			err:           ttyErr,
			env:           map[string]string{"TERM": "dumb"},
			wantOutStyled: false,
			wantErrStyled: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			opts := UIOptions{
				Getenv: func(key string) string {
					return tc.env[key]
				},
				IsTerminal: func(w io.Writer) bool {
					return w == ttyOut || w == ttyErr
				},
				TermWidth: func(w io.Writer) int {
					return 100
				},
			}

			// Act
			ui := newUI(tc.out, tc.err, opts)

			// Assert
			if ui.Out.Styled != tc.wantOutStyled {
				t.Fatalf("ui.Out.Styled = %v, want %v", ui.Out.Styled, tc.wantOutStyled)
			}
			if ui.Err.Styled != tc.wantErrStyled {
				t.Fatalf("ui.Err.Styled = %v, want %v", ui.Err.Styled, tc.wantErrStyled)
			}
		})
	}
}

func TestUIStatusAndNarrowTerminalFormatting(t *testing.T) {
	t.Parallel()

	t.Run("should include both symbol and severity label in styled status badges without emojis when stdout is a TTY", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ttyOut := new(bytes.Buffer)
		ui := newUI(ttyOut, ttyOut, UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(io.Writer) bool { return true },
			TermWidth:  func(io.Writer) int { return 100 },
		})

		// Act
		okBadge := ui.FormatFinding("OK", "Go 1.27.2 is installed")
		warnBadge := ui.FormatFinding("WARN", "shims directory is not on PATH")
		errBadge := ui.FormatFinding("ERROR", "missing version: \"2\"")
		infoBadge := ui.FormatFinding("INFO", "consider enabling \"errcheck\"")

		// Assert
		for _, rendered := range []struct {
			raw    string
			symbol string
			label  string
		}{
			{okBadge, "✓", "OK"},
			{warnBadge, "!", "WARN"},
			{errBadge, "✗", "ERROR"},
			{infoBadge, "i", "INFO"},
		} {
			plain := ansi.Strip(rendered.raw)
			if !strings.Contains(plain, rendered.symbol) || !strings.Contains(plain, rendered.label) {
				t.Fatalf("plain = %q, want both symbol %q and label %q", plain, rendered.symbol, rendered.label)
			}
		}
	})

	t.Run("should stack details vertically without truncating paths or recovery commands when terminal width is narrow", func(t *testing.T) {
		t.Parallel()

		// Arrange
		ttyOut := new(bytes.Buffer)
		ui := newUI(ttyOut, ttyOut, UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(io.Writer) bool { return true },
			TermWidth:  func(io.Writer) int { return 48 },
		})
		longPath := "/Users/developer/very/deeply/nested/workspace/project/.local/share/fgm/go/1.27.2"
		recovery := `fgm install go 1.27.2`

		// Act
		actionLine := ui.FormatAction("Installed", "Go", "1.27.2", "to "+longPath)
		findingLine := ui.FormatFinding("WARN", "go 1.27.2 is missing (run "+recovery+") in "+longPath)

		// Assert
		plainAction := ansi.Strip(actionLine)
		if !strings.Contains(plainAction, "\n") {
			t.Fatalf("expected narrow terminal action output to wrap onto a second line, got %q", plainAction)
		}
		if !strings.Contains(plainAction, "1.27.2") || !strings.Contains(plainAction, longPath) {
			t.Fatalf("expected full version and path to be preserved without truncation, got %q", plainAction)
		}

		plainFinding := ansi.Strip(findingLine)
		if !strings.Contains(plainFinding, longPath) || !strings.Contains(plainFinding, recovery) {
			t.Fatalf("expected full path and recovery command to be preserved, got %q", plainFinding)
		}
	})
}

func TestRootHelpAndErrorsWithFang(t *testing.T) {
	t.Parallel()

	t.Run("should render Fang help without completion or man commands and without --version flag when help is requested", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout, stderr bytes.Buffer
		root := NewRootCmdWithUI(&app.App{}, UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(io.Writer) bool { return false },
		})
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"--help"})

		// Act
		err := Execute(context.Background(), root, UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(io.Writer) bool { return false },
		})

		// Assert
		if err != nil {
			t.Fatalf("Execute(--help): %v", err)
		}
		helpText := stdout.String()
		if strings.Contains(helpText, "\x1b[") {
			t.Fatalf("expected plain help without ANSI sequences, got %q", helpText)
		}
		for _, want := range []string{"USAGE", "COMMANDS", "FLAGS", "current", "doctor", "--chdir"} {
			if !strings.Contains(helpText, want) {
				t.Fatalf("helpText = %q, want it to contain %q", helpText, want)
			}
		}
		for _, forbidden := range []string{"completion", "Generates manpages", "--version"} {
			if strings.Contains(helpText, forbidden) {
				t.Fatalf("helpText = %q, want it NOT to contain %q", helpText, forbidden)
			}
		}
	})

	t.Run("should print error once to stderr without ANSI in plain mode and mark error as printed when command fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout, stderr bytes.Buffer
		root := NewRootCmdWithUI(&app.App{}, UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(io.Writer) bool { return false },
		})
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"use", "go", "1.27.2"})

		// Act
		err := Execute(context.Background(), root, UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(io.Writer) bool { return false },
		})

		// Assert
		if err == nil {
			t.Fatal("expected error from use go without --global")
		}
		if !IsPrintedError(err) {
			t.Fatalf("expected IsPrintedError(err) = true, got false")
		}
		if got := stderr.String(); got != "only --global is supported for now\n" {
			t.Fatalf("stderr = %q, want exact single error line", got)
		}
	})

	t.Run("should render styled error with symbol and ERROR label once when stderr is a TTY", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout, stderr bytes.Buffer
		opts := UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(w io.Writer) bool { return w == &stderr },
			TermWidth:  func(io.Writer) int { return 100 },
		}
		root := NewRootCmdWithUI(&app.App{}, opts)
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"use", "go", "1.27.2"})

		// Act
		err := Execute(context.Background(), root, opts)

		// Assert
		if err == nil {
			t.Fatal("expected error from use go without --global")
		}
		plain := ansi.Strip(stderr.String())
		if !strings.Contains(plain, "✗ ERROR") || !strings.Contains(plain, "only --global is supported for now") {
			t.Fatalf("plain stderr = %q, want '✗ ERROR' and error message", plain)
		}
		if strings.Count(plain, "only --global is supported for now") != 1 {
			t.Fatalf("expected error message to appear once, got %q", plain)
		}
	})
}

func TestCurrentAndVersionsUI(t *testing.T) {
	t.Parallel()

	t.Run("should render aligned styled current toolchain with highlighted versions and muted sources when stdout is a TTY", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout, stderr bytes.Buffer
		opts := UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(w io.Writer) bool { return w == &stdout },
			TermWidth:  func(io.Writer) int { return 100 },
		}
		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.27.2",
						GoSource:    "/repo/go.mod",
						LintVersion: "v2.13.0",
						LintSource:  "local",
					}, nil
				},
			},
		}
		root := NewRootCmdWithUI(application, opts)
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"current"})

		// Act
		err := Execute(context.Background(), root, opts)

		// Assert
		if err != nil {
			t.Fatalf("Execute(current): %v", err)
		}
		raw := stdout.String()
		if !strings.Contains(raw, "\x1b[") {
			t.Fatalf("expected ANSI styled output on TTY, got %q", raw)
		}
		plain := ansi.Strip(raw)
		for _, want := range []string{"go", "1.27.2", "(go.mod)", "golangci-lint", "v2.13.0", "(local)"} {
			if !strings.Contains(plain, want) {
				t.Fatalf("plain = %q, want it to contain %q", plain, want)
			}
		}
	})

	t.Run("should preserve exact plain current output without ANSI when stdout is not a TTY or NO_COLOR is set", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout, stderr bytes.Buffer
		opts := UIOptions{
			Getenv:     func(key string) string { return map[string]string{"NO_COLOR": "1"}[key] },
			IsTerminal: func(io.Writer) bool { return true },
			TermWidth:  func(io.Writer) int { return 100 },
		}
		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{
						GoVersion:   "1.27.2",
						GoSource:    "/repo/go.mod",
						LintVersion: "v2.13.0",
						LintSource:  "local",
					}, nil
				},
			},
		}
		root := NewRootCmdWithUI(application, opts)
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"current"})

		// Act
		err := Execute(context.Background(), root, opts)

		// Assert
		if err != nil {
			t.Fatalf("Execute(current): %v", err)
		}
		want := "go 1.27.2 (go.mod)\ngolangci-lint v2.13.0 (local)\n"
		if stdout.String() != want {
			t.Fatalf("stdout = %q, want exact %q", stdout.String(), want)
		}
	})

	t.Run("should render styled versions with explicit current and recommended labels and info symbol on empty list when stdout is a TTY", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var goOut, lintOut, emptyOut bytes.Buffer
		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.27.2"}, nil
				},
			},
			GoStore: stubGoStore{
				listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
					return []string{"1.27.2", "1.26.9"}, nil
				},
			},
			LintRemoteProvider: stubLintRemoteProvider{
				listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
					return []app.LintVersion{
						{Version: "v2.13.0", Recommended: true},
						{Version: "v2.12.2", Recommended: false},
					}, nil
				},
			},
			LintStore: stubLintStore{
				listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
					return nil, nil
				},
			},
		}

		runTTY := func(buf *bytes.Buffer, args ...string) string {
			opts := UIOptions{
				Getenv:     func(string) string { return "" },
				IsTerminal: func(w io.Writer) bool { return w == buf },
				TermWidth:  func(io.Writer) int { return 100 },
			}
			root := NewRootCmdWithUI(application, opts)
			root.SetOut(buf)
			root.SetErr(new(bytes.Buffer))
			root.SetArgs(args)
			if err := Execute(context.Background(), root, opts); err != nil {
				t.Fatalf("Execute(%v): %v", args, err)
			}
			return buf.String()
		}

		// Act
		rawGo := runTTY(&goOut, "versions", "go", "--local")
		rawLint := runTTY(&lintOut, "versions", "golangci-lint", "--remote", "--go", "1.27.2")
		rawEmpty := runTTY(&emptyOut, "versions", "golangci-lint", "--local")

		// Assert
		if !strings.Contains(rawGo, "\x1b[") || !strings.Contains(ansi.Strip(rawGo), "* 1.27.2") || !strings.Contains(ansi.Strip(rawGo), "(current)") {
			t.Fatalf("rawGo = %q (plain %q), want ANSI + '* 1.27.2' + '(current)'", rawGo, ansi.Strip(rawGo))
		}
		if !strings.Contains(rawLint, "\x1b[") || !strings.Contains(ansi.Strip(rawLint), "* v2.13.0") || !strings.Contains(ansi.Strip(rawLint), "(recommended)") {
			t.Fatalf("rawLint = %q (plain %q), want ANSI + '* v2.13.0' + '(recommended)'", rawLint, ansi.Strip(rawLint))
		}
		if !strings.Contains(rawEmpty, "\x1b[") || !strings.Contains(ansi.Strip(rawEmpty), "i No local golangci-lint versions installed.") {
			t.Fatalf("rawEmpty = %q (plain %q), want ANSI + 'i No local golangci-lint versions installed.'", rawEmpty, ansi.Strip(rawEmpty))
		}
	})
}

type stubLintDoctor struct {
	diagnoseFn func(ctx context.Context, workDir string) ([]app.LintFinding, error)
}

func (s stubLintDoctor) Diagnose(ctx context.Context, workDir string) ([]app.LintFinding, error) {
	return s.diagnoseFn(ctx, workDir)
}

func TestDoctorAndLintDoctorUI(t *testing.T) {
	t.Parallel()

	t.Run("should render styled doctor and lint doctor findings with severity symbols and labels when stdout is a TTY", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var docOut, lintDocOut bytes.Buffer
		application := &app.App{
			Doctor: stubDoctor{
				diagnoseFn: func(ctx context.Context, workDir string) ([]app.DoctorFinding, error) {
					return []app.DoctorFinding{
						{Severity: "OK", Message: "Go 1.27.2 is installed (go.mod)"},
						{Severity: "WARN", Message: `FGM shims directory is not on PATH (in non-interactive sessions, run "fgm exec -- <command>")`, FixKind: "shell_profile"},
					}, nil
				},
			},
			LintDoctor: stubLintDoctor{
				diagnoseFn: func(ctx context.Context, workDir string) ([]app.LintFinding, error) {
					return []app.LintFinding{
						{Severity: "ERROR", Message: `config is missing 'version: "2"'`},
						{Severity: "INFO", Message: `consider enabling "errcheck"`},
					}, nil
				},
			},
		}

		docOpts := UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(w io.Writer) bool { return w == &docOut },
			TermWidth:  func(io.Writer) int { return 100 },
		}
		docRoot := NewRootCmdWithUI(application, docOpts)
		docRoot.SetOut(&docOut)
		docRoot.SetErr(new(bytes.Buffer))
		docRoot.SetArgs([]string{"doctor"})

		lintDocOpts := UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(w io.Writer) bool { return w == &lintDocOut },
			TermWidth:  func(io.Writer) int { return 100 },
		}
		lintDocRoot := NewRootCmdWithUI(application, lintDocOpts)
		lintDocRoot.SetOut(&lintDocOut)
		lintDocRoot.SetErr(new(bytes.Buffer))
		lintDocRoot.SetArgs([]string{"lint", "doctor"})

		// Act
		if err := Execute(context.Background(), docRoot, docOpts); err != nil {
			t.Fatalf("Execute(doctor): %v", err)
		}
		if err := Execute(context.Background(), lintDocRoot, lintDocOpts); err != nil {
			t.Fatalf("Execute(lint doctor): %v", err)
		}

		// Assert
		if !strings.Contains(docOut.String(), "\x1b[") {
			t.Fatalf("expected ANSI styled output for doctor on TTY, got %q", docOut.String())
		}
		plainDoc := ansi.Strip(docOut.String())
		for _, want := range []string{"✓ OK", "! WARN", `fgm exec -- <command>`, `eval "$(fgm env)"`} {
			if !strings.Contains(plainDoc, want) {
				t.Fatalf("plainDoc = %q, want it to contain %q", plainDoc, want)
			}
		}

		if !strings.Contains(lintDocOut.String(), "\x1b[") {
			t.Fatalf("expected ANSI styled output for lint doctor on TTY, got %q", lintDocOut.String())
		}
		plainLintDoc := ansi.Strip(lintDocOut.String())
		for _, want := range []string{"✗ ERROR", "i INFO", `version: "2"`, `"errcheck"`} {
			if !strings.Contains(plainLintDoc, want) {
				t.Fatalf("plainLintDoc = %q, want it to contain %q", plainLintDoc, want)
			}
		}
	})
}

type stubLintConfigGenerator struct {
	generateFn func(ctx context.Context, opts app.LintConfigOptions) ([]byte, error)
}

func (s stubLintConfigGenerator) Generate(ctx context.Context, opts app.LintConfigOptions) ([]byte, error) {
	return s.generateFn(ctx, opts)
}

type stubExecutor struct {
	execFn func(ctx context.Context, workDir string, args []string, options app.ExecOptions, stdin io.Reader, stdout io.Writer, stderr io.Writer) error
}

func (s stubExecutor) Exec(ctx context.Context, workDir string, args []string, options app.ExecOptions, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	return s.execFn(ctx, workDir, args, options, stdin, stdout, stderr)
}

func TestCommandSuccessMessagesAndRegressionsUI(t *testing.T) {
	t.Parallel()

	t.Run("should render styled success messages with status symbol and highlighted version when stdout is a TTY", func(t *testing.T) {
		t.Parallel()

		// Arrange
		pinDir := t.TempDir()
		application := &app.App{
			GoInstaller: stubGoInstaller{
				installGoVersionFn: func(ctx context.Context, version string) (string, error) {
					return "/store/go/" + version, nil
				},
			},
			GoImporter: stubGoImporter{
				importAutoFn: func(ctx context.Context) ([]app.ImportedGo, error) {
					return []app.ImportedGo{{Version: "1.27.2", Path: "/usr/local/go"}}, nil
				},
			},
			GoStore: stubGoStore{
				hasGoVersionFn: func(ctx context.Context, version string) (bool, error) {
					return true, nil
				},
				setGlobalGoVersionFn: func(ctx context.Context, version string) error {
					return nil
				},
				ensureShimsFn: func() error {
					return nil
				},
				shimDirFn: func() string {
					return "/store/shims"
				},
				deleteGoVersionFn: func(ctx context.Context, version string) (string, error) {
					return "/store/go/" + version, nil
				},
			},
			GoUpgrader: stubGoUpgrader{
				upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
					return app.GoUpgradeResult{Version: "1.27.2", DryRun: options.DryRun}, nil
				},
			},
			LintConfigGenerator: stubLintConfigGenerator{
				generateFn: func(ctx context.Context, opts app.LintConfigOptions) ([]byte, error) {
					return []byte("version: \"2\"\n"), nil
				},
			},
		}

		cases := []struct {
			name string
			args []string
			want []string
		}{
			{
				name: "should style install go output when stdout is a TTY",
				args: []string{"install", "go", "1.27.2"},
				want: []string{"✓", "Installed", "Go", "1.27.2", "/store/go/1.27.2"},
			},
			{
				name: "should style import auto output when stdout is a TTY",
				args: []string{"import", "auto"},
				want: []string{"✓", "Imported", "Go", "1.27.2", "/usr/local/go"},
			},
			{
				name: "should style use go --global output when stdout is a TTY",
				args: []string{"use", "go", "1.27.2", "--global"},
				want: []string{"✓", "Selected", "Go", "1.27.2", "/store/shims"},
			},
			{
				name: "should style pin golangci-lint output when stdout is a TTY",
				args: []string{"pin", "golangci-lint", "v2.13.0", "--chdir", pinDir},
				want: []string{"✓", "Pinned", "golangci-lint", "v2.13.0", ".fgm.toml"},
			},
			{
				name: "should style remove go output when stdout is a TTY",
				args: []string{"remove", "go", "1.27.2"},
				want: []string{"✓", "Removed", "Go", "1.27.2", "/store/go/1.27.2"},
			},
			{
				name: "should style upgrade go --global output when stdout is a TTY",
				args: []string{"upgrade", "go", "--global"},
				want: []string{"✓", "Upgraded global Go to 1.27.2"},
			},
			{
				name: "should style lint init output when stdout is a TTY",
				args: []string{"lint", "init", "--preset", "standard"},
				want: []string{"✓", "Generated", ".golangci.yml", "(preset: standard)"},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// Arrange
				var stdout bytes.Buffer
				opts := UIOptions{
					Getenv:     func(string) string { return "" },
					IsTerminal: func(w io.Writer) bool { return w == &stdout },
					TermWidth:  func(io.Writer) int { return 100 },
				}
				root := NewRootCmdWithUI(application, opts)
				root.SetOut(&stdout)
				root.SetErr(new(bytes.Buffer))
				root.SetArgs(tc.args)

				// Act
				err := Execute(context.Background(), root, opts)

				// Assert
				if err != nil {
					t.Fatalf("Execute(%v): %v", tc.args, err)
				}
				raw := stdout.String()
				if !strings.Contains(raw, "\x1b[") {
					t.Fatalf("expected ANSI styled output for %v on TTY, got %q", tc.args, raw)
				}
				plain := ansi.Strip(raw)
				for _, expected := range tc.want {
					if !strings.Contains(plain, expected) {
						t.Fatalf("plain(%v) = %q, want it to contain %q", tc.args, plain, expected)
					}
				}
			})
		}
	})

	t.Run("should keep env, version, and context cancellation completely unstyled even when stdout is a TTY", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var envOut, verOut bytes.Buffer
		application := &app.App{
			BuildInfo: app.BuildInfo{
				Version: "v1.2.3",
				Commit:  "abc1234",
				Date:    "2026-10-14T12:00:00Z",
			},
			EnvRenderer: stubEnvRenderer{
				renderFn: func(ctx context.Context, shell string) ([]string, error) {
					return []string{`export PATH="/store/shims:$PATH"`}, nil
				},
			},
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{}, ctx.Err()
				},
			},
		}

		envOpts := UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(io.Writer) bool { return true },
			TermWidth:  func(io.Writer) int { return 100 },
		}
		envRoot := NewRootCmdWithUI(application, envOpts)
		envRoot.SetOut(&envOut)
		envRoot.SetErr(new(bytes.Buffer))
		envRoot.SetArgs([]string{"env"})

		verRoot := NewRootCmdWithUI(application, envOpts)
		verRoot.SetOut(&verOut)
		verRoot.SetErr(new(bytes.Buffer))
		verRoot.SetArgs([]string{"version"})

		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		cancelRoot := NewRootCmdWithUI(application, envOpts)
		cancelRoot.SetOut(new(bytes.Buffer))
		cancelRoot.SetErr(new(bytes.Buffer))
		cancelRoot.SetArgs([]string{"current"})

		// Act
		if err := Execute(context.Background(), envRoot, envOpts); err != nil {
			t.Fatalf("Execute(env): %v", err)
		}
		if err := Execute(context.Background(), verRoot, envOpts); err != nil {
			t.Fatalf("Execute(version): %v", err)
		}
		cancelErr := Execute(cancelCtx, cancelRoot, envOpts)

		// Assert
		if envOut.String() != "export PATH=\"/store/shims:$PATH\"\n" {
			t.Fatalf("envOut = %q, want exact unstyled shell code", envOut.String())
		}
		if verOut.String() != "fgm v1.2.3\ncommit abc1234\ndate 2026-10-14T12:00:00Z\n" {
			t.Fatalf("verOut = %q, want exact unstyled version payload", verOut.String())
		}
		if cancelErr == nil || !strings.Contains(cancelErr.Error(), context.Canceled.Error()) {
			t.Fatalf("cancelErr = %v, want context.Canceled", cancelErr)
		}
	})

	t.Run("should pass stdout writer and child ANSI sequences transparently through exec even when fgm is in plain mode", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var execOut bytes.Buffer
		var receivedStdout io.Writer
		application := &app.App{
			Executor: stubExecutor{
				execFn: func(ctx context.Context, workDir string, args []string, options app.ExecOptions, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					receivedStdout = stdout
					_, err := io.WriteString(stdout, "\x1b[32mchild-ansi\x1b[0m\n")
					return err
				},
			},
		}
		plainOpts := UIOptions{
			Getenv:     func(key string) string { return map[string]string{"NO_COLOR": "1"}[key] },
			IsTerminal: func(io.Writer) bool { return false },
		}
		execRoot := NewRootCmdWithUI(application, plainOpts)
		execRoot.SetOut(&execOut)
		execRoot.SetErr(new(bytes.Buffer))
		execRoot.SetArgs([]string{"exec", "--", "go", "test"})

		// Act
		err := Execute(context.Background(), execRoot, plainOpts)

		// Assert
		if err != nil {
			t.Fatalf("Execute(exec): %v", err)
		}
		if receivedStdout != &execOut {
			t.Fatalf("receivedStdout = %T, want original *bytes.Buffer without wrapper", receivedStdout)
		}
		if execOut.String() != "\x1b[32mchild-ansi\x1b[0m\n" {
			t.Fatalf("execOut = %q, want child ANSI preserved verbatim", execOut.String())
		}
	})
}

type stubUpdateChecker struct {
	checkLatestFn  func(ctx context.Context) (app.FGMUpdateStatus, bool, error)
	cachedNoticeFn func(ctx context.Context) (app.FGMUpdateStatus, bool)
}

func (s stubUpdateChecker) CheckLatest(ctx context.Context) (app.FGMUpdateStatus, bool, error) {
	if s.checkLatestFn == nil {
		return app.FGMUpdateStatus{}, false, nil
	}
	return s.checkLatestFn(ctx)
}

func (s stubUpdateChecker) CachedNotice(ctx context.Context) (app.FGMUpdateStatus, bool) {
	if s.cachedNoticeFn == nil {
		return app.FGMUpdateStatus{}, false
	}
	return s.cachedNoticeFn(ctx)
}

func TestUpdateNotifier_ShowsCachedNoticeOnTTYStderrAndSuppressesWhenDisabled(t *testing.T) {
	t.Parallel()

	t.Run("should print cached update notice to TTY stderr after current command and suppress when FGM_NO_UPDATE_NOTIFIER=1", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var outBuf, errBuf bytes.Buffer
		application := &app.App{
			Resolver: stubSelectionResolver{
				currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
					return app.Selection{GoVersion: "1.27.2", GoSource: "go.mod"}, nil
				},
			},
			FGMUpdateChecker: stubUpdateChecker{
				cachedNoticeFn: func(ctx context.Context) (app.FGMUpdateStatus, bool) {
					return app.FGMUpdateStatus{
						CurrentVersion:  "0.5.0",
						LatestVersion:   "0.6.0",
						UpdateAvailable: true,
						UpgradeCommand:  "brew update && brew upgrade --cask fgm",
					}, true
				},
			},
		}

		ttyOpts := UIOptions{
			Getenv:     func(string) string { return "" },
			IsTerminal: func(w io.Writer) bool { return w == &outBuf || w == &errBuf },
			TermWidth:  func(io.Writer) int { return 100 },
		}
		root := NewRootCmdWithUI(application, ttyOpts)
		root.SetOut(&outBuf)
		root.SetErr(&errBuf)
		root.SetArgs([]string{"current"})

		// Act
		err := Execute(context.Background(), root, ttyOpts)

		// Assert
		if err != nil {
			t.Fatalf("Execute(current): %v", err)
		}
		plainErr := ansi.Strip(errBuf.String())
		for _, want := range []string{"0.5.0 -> 0.6.0", "brew update && brew upgrade --cask fgm"} {
			if !strings.Contains(plainErr, want) {
				t.Fatalf("plainErr = %q, want it to contain %q", plainErr, want)
			}
		}

		// Arrange — disabled via FGM_NO_UPDATE_NOTIFIER=1
		var suppressedOut, suppressedErr bytes.Buffer
		disabledOpts := UIOptions{
			Getenv: func(key string) string {
				if key == "FGM_NO_UPDATE_NOTIFIER" {
					return "1"
				}
				return ""
			},
			IsTerminal: func(w io.Writer) bool { return w == &suppressedOut || w == &suppressedErr },
			TermWidth:  func(io.Writer) int { return 100 },
		}
		disabledRoot := NewRootCmdWithUI(application, disabledOpts)
		disabledRoot.SetOut(&suppressedOut)
		disabledRoot.SetErr(&suppressedErr)
		disabledRoot.SetArgs([]string{"current"})

		// Act
		if err := Execute(context.Background(), disabledRoot, disabledOpts); err != nil {
			t.Fatalf("Execute(current disabled): %v", err)
		}

		// Assert
		if suppressedErr.String() != "" {
			t.Fatalf("suppressedErr = %q, want empty stderr when FGM_NO_UPDATE_NOTIFIER=1", suppressedErr.String())
		}
	})
}

func TestInstallAndUpgradeSubcommands_RefreshUpdateCache(t *testing.T) {
	t.Parallel()

	t.Run("should refresh update cache when install go and upgrade go complete", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var checkCount int
		application := &app.App{
			GoInstaller: stubGoInstaller{
				installGoVersionFn: func(ctx context.Context, version string) (string, error) {
					return "/tmp/fgm/go/" + version, nil
				},
			},
			GoUpgrader: stubGoUpgrader{
				upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
					return app.GoUpgradeResult{Version: "1.27.2"}, nil
				},
			},
			FGMUpdateChecker: stubUpdateChecker{
				checkLatestFn: func(ctx context.Context) (app.FGMUpdateStatus, bool, error) {
					checkCount++
					return app.FGMUpdateStatus{}, false, nil
				},
			},
		}

		// Act
		root1 := NewRootCmd(application)
		root1.SetOut(new(bytes.Buffer))
		root1.SetErr(new(bytes.Buffer))
		root1.SetArgs([]string{"install", "go", "1.27.2"})
		if err := root1.Execute(); err != nil {
			t.Fatalf("install go: %v", err)
		}

		root2 := NewRootCmd(application)
		root2.SetOut(new(bytes.Buffer))
		root2.SetErr(new(bytes.Buffer))
		root2.SetArgs([]string{"upgrade", "go", "--global"})
		if err := root2.Execute(); err != nil {
			t.Fatalf("upgrade go --global: %v", err)
		}

		// Assert
		if checkCount != 2 {
			t.Fatalf("checkCount = %d, want 2", checkCount)
		}
	})
}
