package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type bumpLevel int

const (
	bumpNone bumpLevel = iota
	bumpPatch
	bumpMinor
	bumpMajor
)

type gitRunnerFunc func(ctx context.Context, args ...string) (string, error)

var (
	stdoutWriter     io.Writer     = os.Stdout
	stderrWriter     io.Writer     = os.Stderr
	defaultGitRunner gitRunnerFunc = func(ctx context.Context, args ...string) (string, error) {
		cmd := exec.CommandContext(ctx, "git", args...)
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}
	exitMain = os.Exit
)

func main() {
	if err := run(context.Background(), os.Args[1:], stdoutWriter, stderrWriter, defaultGitRunner); err != nil {
		_, _ = fmt.Fprintln(stderrWriter, err)
		exitMain(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	git gitRunnerFunc,
) error {
	fs := flag.NewFlagSet("next-version", flag.ContinueOnError)
	fs.SetOutput(stderr)

	currentTagFlag := fs.String("current-tag", "", "override current SemVer tag (defaults to latest reachable v* tag)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	currentTag := strings.TrimSpace(*currentTagFlag)
	revRange := "HEAD"

	if currentTag == "" {
		described, err := git(ctx, "describe", "--tags", "--match", "v[0-9]*", "--abbrev=0")
		if err == nil {
			currentTag = strings.TrimSpace(described)
		}
	}

	if currentTag != "" {
		if _, _, _, err := parseSemVer(currentTag); err != nil {
			return err
		}
		revRange = currentTag + "..HEAD"
	}

	rawLog, err := git(ctx, "log", "--format=%B%x00", revRange)
	if err != nil {
		return fmt.Errorf("read git log for %s: %w", revRange, err)
	}

	commits := splitCommitMessages(rawLog)
	tag, shouldRelease, err := nextVersion(currentTag, commits)
	if err != nil {
		return err
	}
	if !shouldRelease {
		return nil
	}

	_, err = fmt.Fprintln(stdout, tag)
	return err
}

func splitCommitMessages(raw string) []string {
	parts := strings.Split(raw, "\x00")
	messages := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			messages = append(messages, trimmed)
		}
	}
	return messages
}

func nextVersion(currentTag string, commitMessages []string) (string, bool, error) {
	major, minor, patch, err := parseSemVer(currentTag)
	if err != nil {
		return "", false, err
	}

	highest := bumpNone
	for _, msg := range commitMessages {
		if level := classifyCommit(msg); level > highest {
			highest = level
		}
	}

	if highest == bumpNone {
		return "", false, nil
	}

	switch highest {
	case bumpMajor:
		if major == 0 {
			minor++
			patch = 0
		} else {
			major++
			minor = 0
			patch = 0
		}
	case bumpMinor:
		minor++
		patch = 0
	case bumpPatch:
		patch++
	}

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch), true, nil
}

func parseSemVer(tag string) (int, int, int, error) {
	trimmed := strings.TrimSpace(tag)
	if trimmed == "" {
		return 0, 0, 0, nil
	}

	normalized := strings.TrimPrefix(trimmed, "v")
	parts := strings.Split(normalized, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semantic version tag %q: expected vMAJOR.MINOR.PATCH", tag)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return 0, 0, 0, fmt.Errorf("invalid major version in %q", tag)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil || minor < 0 {
		return 0, 0, 0, fmt.Errorf("invalid minor version in %q", tag)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil || patch < 0 {
		return 0, 0, 0, fmt.Errorf("invalid patch version in %q", tag)
	}

	return major, minor, patch, nil
}

func classifyCommit(message string) bumpLevel {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" || strings.HasPrefix(trimmed, "Merge ") {
		return bumpNone
	}

	lines := strings.Split(trimmed, "\n")
	header := strings.TrimSpace(lines[0])

	if strings.Contains(trimmed, "BREAKING CHANGE:") || strings.Contains(trimmed, "BREAKING-CHANGE:") {
		return bumpMajor
	}

	colonIdx := strings.IndexByte(header, ':')
	if colonIdx <= 0 {
		return bumpPatch
	}

	prefix := header[:colonIdx]
	isBreaking := strings.HasSuffix(prefix, "!")
	if isBreaking {
		prefix = strings.TrimSuffix(prefix, "!")
	}

	if scopeStart := strings.IndexByte(prefix, '('); scopeStart >= 0 {
		if !strings.HasSuffix(prefix, ")") {
			return bumpPatch
		}
		prefix = prefix[:scopeStart]
	}

	if prefix == "" || strings.ContainsAny(prefix, " \t") {
		return bumpPatch
	}

	if isBreaking {
		return bumpMajor
	}

	switch strings.ToLower(prefix) {
	case "feat":
		return bumpMinor
	case "fix", "perf", "refactor", "revert":
		return bumpPatch
	case "docs", "chore", "test", "ci", "style", "build":
		return bumpNone
	default:
		return bumpPatch
	}
}
