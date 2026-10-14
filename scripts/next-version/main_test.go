package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestNextVersion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		currentTag  string
		commits     []string
		wantTag     string
		wantRelease bool
		wantErr     bool
	}{
		{
			name:        "should bump minor version when feat commit is present on v0 release line",
			currentTag:  "v0.3.0",
			commits:     []string{"docs: update README", "feat(ui): add Lip Gloss output", "fix: handle nil writer"},
			wantTag:     "v0.4.0",
			wantRelease: true,
		},
		{
			name:        "should bump minor version on v0 when breaking change bang is present",
			currentTag:  "v0.3.0",
			commits:     []string{"fix!: drop legacy flag"},
			wantTag:     "v0.4.0",
			wantRelease: true,
		},
		{
			name:        "should bump major version on v1 when breaking change footer is present",
			currentTag:  "v1.2.3",
			commits:     []string{"feat: redesign config\n\nBREAKING CHANGE: config schema changed"},
			wantTag:     "v2.0.0",
			wantRelease: true,
		},
		{
			name:        "should bump minor version on v1 when feat commit is present without breaking changes",
			currentTag:  "v1.2.3",
			commits:     []string{"feat: add fish completion"},
			wantTag:     "v1.3.0",
			wantRelease: true,
		},
		{
			name:        "should bump patch version when fix or refactor commit is present",
			currentTag:  "v0.3.0",
			commits:     []string{"chore: tidy deps", "fix(shim): resolve path offline"},
			wantTag:     "v0.3.1",
			wantRelease: true,
		},
		{
			name:        "should bump patch version when plain non-conventional commit is present",
			currentTag:  "v0.3.0",
			commits:     []string{"Add auto-fix option to doctor command"},
			wantTag:     "v0.3.1",
			wantRelease: true,
		},
		{
			name:        "should skip release when only docs chore test ci and merge commits are present",
			currentTag:  "v0.3.0",
			commits:     []string{"docs: update installation guide", "chore: refresh compatibility", "test: add unit test", "ci: update workflow", "Merge branch 'main' of github.com:kostikovk/fgm"},
			wantTag:     "",
			wantRelease: false,
		},
		{
			name:        "should skip release when commit list is empty",
			currentTag:  "v0.3.0",
			commits:     nil,
			wantTag:     "",
			wantRelease: false,
		},
		{
			name:        "should default to v0.1.0 when no previous tag exists and feat commit is present",
			currentTag:  "",
			commits:     []string{"feat: initial release"},
			wantTag:     "v0.1.0",
			wantRelease: true,
		},
		{
			name:       "should return an error when current tag is malformed",
			currentTag: "invalid-tag",
			commits:    []string{"feat: something"},
			wantErr:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			currentTag := tc.currentTag
			commits := tc.commits

			// Act
			gotTag, gotRelease, err := nextVersion(currentTag, commits)

			// Assert
			if tc.wantErr {
				if err == nil {
					t.Fatalf("nextVersion(%q, %v) returned nil error; want error", currentTag, commits)
				}
				return
			}
			if err != nil {
				t.Fatalf("nextVersion(%q, %v) unexpected error: %v", currentTag, commits, err)
			}
			if gotRelease != tc.wantRelease {
				t.Errorf("nextVersion(%q, %v) release = %v; want %v", currentTag, commits, gotRelease, tc.wantRelease)
			}
			if gotTag != tc.wantTag {
				t.Errorf("nextVersion(%q, %v) tag = %q; want %q", currentTag, commits, gotTag, tc.wantTag)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("should print next tag to stdout when releasable commits exist since latest tag", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		gitStub := func(_ context.Context, args ...string) (string, error) {
			joined := strings.Join(args, " ")
			switch joined {
			case "describe --tags --match v[0-9]* --abbrev=0":
				return "v0.3.0\n", nil
			case "log --format=%B%x00 v0.3.0..HEAD":
				return "docs: update readme\x00feat: add new command\x00", nil
			default:
				return "", errors.New("unexpected git args: " + joined)
			}
		}

		// Act
		err := run(context.Background(), nil, &stdout, &stderr, gitStub)

		// Assert
		if err != nil {
			t.Fatalf("run returned unexpected error: %v", err)
		}
		if got := stdout.String(); got != "v0.4.0\n" {
			t.Errorf("stdout = %q; want %q", got, "v0.4.0\n")
		}
		if stderr.Len() != 0 {
			t.Errorf("stderr = %q; want empty", stderr.String())
		}
	})

	t.Run("should print nothing when only non-releasable commits exist since latest tag", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		gitStub := func(_ context.Context, args ...string) (string, error) {
			joined := strings.Join(args, " ")
			switch joined {
			case "describe --tags --match v[0-9]* --abbrev=0":
				return "v0.3.0\n", nil
			case "log --format=%B%x00 v0.3.0..HEAD":
				return "docs: clarify readme\x00chore: update catalog\x00", nil
			default:
				return "", errors.New("unexpected git args: " + joined)
			}
		}

		// Act
		err := run(context.Background(), nil, &stdout, &stderr, gitStub)

		// Assert
		if err != nil {
			t.Fatalf("run returned unexpected error: %v", err)
		}
		if got := stdout.String(); got != "" {
			t.Errorf("stdout = %q; want empty output", got)
		}
	})

	t.Run("should inspect all HEAD commits when repository has no tags yet", func(t *testing.T) {
		t.Parallel()

		// Arrange
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		gitStub := func(_ context.Context, args ...string) (string, error) {
			joined := strings.Join(args, " ")
			switch joined {
			case "describe --tags --match v[0-9]* --abbrev=0":
				return "", errors.New("fatal: No names found, cannot describe anything")
			case "log --format=%B%x00 HEAD":
				return "fix: initial patch\x00", nil
			default:
				return "", errors.New("unexpected git args: " + joined)
			}
		}

		// Act
		err := run(context.Background(), nil, &stdout, &stderr, gitStub)

		// Assert
		if err != nil {
			t.Fatalf("run returned unexpected error: %v", err)
		}
		if got := stdout.String(); got != "v0.0.1\n" {
			t.Errorf("stdout = %q; want %q", got, "v0.0.1\n")
		}
	})

	t.Run("should return error when git log fails", func(t *testing.T) {
		t.Parallel()

		// Arrange
		gitStub := func(_ context.Context, args ...string) (string, error) {
			if args[0] == "describe" {
				return "v0.3.0\n", nil
			}
			return "", errors.New("git log failed")
		}

		// Act
		err := run(context.Background(), nil, &bytes.Buffer{}, &bytes.Buffer{}, gitStub)

		// Assert
		if err == nil || !strings.Contains(err.Error(), "git log failed") {
			t.Fatalf("err = %v; want git log error", err)
		}
	})
}

func TestMainWritesErrorAndExits(t *testing.T) {
	t.Run("should write error to stderr and exit with status 1 when run fails", func(t *testing.T) {
		// Arrange
		oldStdout := stdoutWriter
		oldStderr := stderrWriter
		oldGitRunner := defaultGitRunner
		oldExit := exitMain
		oldArgs := os.Args
		t.Cleanup(func() {
			stdoutWriter = oldStdout
			stderrWriter = oldStderr
			defaultGitRunner = oldGitRunner
			exitMain = oldExit
			os.Args = oldArgs
		})

		var stderr bytes.Buffer
		stdoutWriter = &bytes.Buffer{}
		stderrWriter = &stderr
		defaultGitRunner = func(_ context.Context, _ ...string) (string, error) {
			return "", errors.New("git failure")
		}
		exitCode := 0
		exitMain = func(code int) {
			exitCode = code
			panic("exit")
		}
		os.Args = []string{"next-version", "-current-tag", "invalid-tag"}

		// Act & Assert
		defer func() {
			if recovered := recover(); recovered != "exit" {
				t.Fatalf("recovered = %v; want exit panic", recovered)
			}
			if exitCode != 1 {
				t.Errorf("exitCode = %d; want 1", exitCode)
			}
			if !strings.Contains(stderr.String(), "invalid-tag") {
				t.Errorf("stderr = %q; want invalid-tag error", stderr.String())
			}
		}()

		main()
		t.Fatal("main should have exited")
	})
}
