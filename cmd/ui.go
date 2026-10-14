package cmd

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"strconv"
	"strings"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

type uiOptionsContextKey struct{}

func withUIOptions(ctx context.Context, opts UIOptions) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, uiOptionsContextKey{}, opts)
}

func optsForCommand(cmd *cobra.Command) UIOptions {
	if cmd == nil || cmd.Context() == nil {
		return UIOptions{}
	}
	if opts, ok := cmd.Context().Value(uiOptionsContextKey{}).(UIOptions); ok {
		return opts
	}
	return UIOptions{}
}

func uiFromCmd(cmd *cobra.Command) *uiRenderer {
	return uiForCommand(cmd, optsForCommand(cmd))
}

const (
	defaultTermWidth    = 80
	narrowTermThreshold = 68
)

type fdProvider interface {
	Fd() uintptr
}

// UIOptions configures terminal capability detection and layout for CLI output.
type UIOptions struct {
	Getenv     func(string) string
	IsTerminal func(io.Writer) bool
	TermWidth  func(io.Writer) int
}

func (o UIOptions) withDefaults() UIOptions {
	if o.Getenv == nil {
		o.Getenv = os.Getenv
	}
	if o.IsTerminal == nil {
		o.IsTerminal = defaultIsTerminal
	}
	if o.TermWidth == nil {
		o.TermWidth = func(w io.Writer) int {
			return defaultDetectWidth(w, o.Getenv)
		}
	}
	return o
}

func defaultIsTerminal(w io.Writer) bool {
	if f, ok := w.(fdProvider); ok {
		return term.IsTerminal(f.Fd())
	}
	return false
}

func defaultDetectWidth(w io.Writer, getenv func(string) string) int {
	if getenv != nil {
		if cols, err := strconv.Atoi(strings.TrimSpace(getenv("COLUMNS"))); err == nil && cols > 0 {
			return cols
		}
	}
	if f, ok := w.(fdProvider); ok && term.IsTerminal(f.Fd()) {
		if width, _, err := term.GetSize(f.Fd()); err == nil && width > 0 {
			return width
		}
	}
	return defaultTermWidth
}

type streamCapabilities struct {
	Styled bool
	Width  int
}

type uiTheme struct {
	accent  lipgloss.Style
	version lipgloss.Style
	muted   lipgloss.Style
	ok      lipgloss.Style
	warn    lipgloss.Style
	err     lipgloss.Style
	info    lipgloss.Style
}

func newUITheme(isDark bool) uiTheme {
	pick := lipgloss.LightDark(isDark)
	return uiTheme{
		accent:  lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#007F7B"), lipgloss.Color("#2EC4B6"))).Bold(true),
		version: lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#005F73"), lipgloss.Color("#5BC0BE"))).Bold(true),
		muted:   lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#5C677D"), lipgloss.Color("#8D99AE"))),
		ok:      lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#0F766E"), lipgloss.Color("#2EC4B6"))).Bold(true),
		warn:    lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#B45309"), lipgloss.Color("#F59E0B"))).Bold(true),
		err:     lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#B91C1C"), lipgloss.Color("#EF4444"))).Bold(true),
		info:    lipgloss.NewStyle().Foreground(pick(lipgloss.Color("#0369A1"), lipgloss.Color("#38BDF8"))).Bold(true),
	}
}

func detectDarkTheme(getenv func(string) string) bool {
	if getenv == nil {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(getenv("FGM_THEME"))) {
	case "light":
		return false
	case "dark":
		return true
	}
	if fgbg := strings.TrimSpace(getenv("COLORFGBG")); fgbg != "" {
		parts := strings.Split(fgbg, ";")
		if bg, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
			// Standard xterm/rxvt COLORFGBG background indices: 7 and 15 are light.
			return bg < 7 || (bg > 7 && bg < 15)
		}
	}
	return true
}

type uiRenderer struct {
	Out   streamCapabilities
	Err   streamCapabilities
	theme uiTheme
}

func newUI(out io.Writer, err io.Writer, opts UIOptions) *uiRenderer {
	resolved := opts.withDefaults()
	envAllowsStyle := resolved.Getenv("NO_COLOR") == "" &&
		resolved.Getenv("CI") == "" &&
		!strings.EqualFold(strings.TrimSpace(resolved.Getenv("TERM")), "dumb")

	outWidth := resolved.TermWidth(out)
	if outWidth <= 0 {
		outWidth = defaultTermWidth
	}
	errWidth := resolved.TermWidth(err)
	if errWidth <= 0 {
		errWidth = defaultTermWidth
	}

	return &uiRenderer{
		Out: streamCapabilities{
			Styled: envAllowsStyle && resolved.IsTerminal(out),
			Width:  outWidth,
		},
		Err: streamCapabilities{
			Styled: envAllowsStyle && resolved.IsTerminal(err),
			Width:  errWidth,
		},
		theme: newUITheme(detectDarkTheme(resolved.Getenv)),
	}
}

func uiForCommand(cmd *cobra.Command, opts UIOptions) *uiRenderer {
	return newUI(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
}

func (u *uiRenderer) isNarrowOut() bool {
	return u.Out.Width > 0 && u.Out.Width < narrowTermThreshold
}

// FormatFinding renders a diagnostic finding in plain or styled TTY mode.
func (u *uiRenderer) FormatFinding(severity string, message string) string {
	if !u.Out.Styled {
		return severity + " " + message
	}

	var badge string
	switch severity {
	case "OK":
		badge = u.theme.ok.Render("✓ OK")
	case "WARN":
		badge = u.theme.warn.Render("! WARN")
	case "ERROR":
		badge = u.theme.err.Render("✗ ERROR")
	case "INFO":
		badge = u.theme.info.Render("i INFO")
	default:
		badge = u.theme.accent.Render("• " + severity)
	}

	singleLine := badge + "  " + message
	if u.isNarrowOut() && ansi.StringWidth(singleLine) > u.Out.Width {
		return badge + "\n  " + message
	}
	return singleLine
}

// FormatAction renders a styled command outcome when stdout is a styled TTY.
func (u *uiRenderer) FormatAction(verb string, tool string, version string, detail string) string {
	symbol := u.theme.ok.Render("✓")
	if strings.HasPrefix(verb, "Would ") {
		symbol = u.theme.info.Render("~")
	}

	prefixVerb, hasToSuffix := strings.CutSuffix(verb, " to")
	head := symbol + " " + prefixVerb
	if tool != "" {
		head += " " + u.theme.accent.Render(tool)
	}
	if hasToSuffix {
		head += " to"
	}
	if version != "" {
		head += " " + u.theme.version.Render(version)
	}
	if detail == "" {
		return head
	}

	styledDetail := u.theme.muted.Render(detail)
	singleLine := head + " " + styledDetail
	if u.isNarrowOut() || (u.Out.Width > 0 && ansi.StringWidth(singleLine) > u.Out.Width) {
		return head + "\n  " + styledDetail
	}
	return singleLine
}

// FormatOutcome returns plain in plain mode and FormatAction in styled TTY mode.
func (u *uiRenderer) FormatOutcome(plain string, verb string, tool string, version string, detail string) string {
	if !u.Out.Styled {
		return plain
	}
	return u.FormatAction(verb, tool, version, detail)
}

// FormatUseGoGlobal renders the output of `fgm use go <version> --global`.
func (u *uiRenderer) FormatUseGoGlobal(version string, shimDir string) string {
	if !u.Out.Styled {
		return fmt.Sprintf(
			"Selected Go %s as the global default.\nAdd %s to PATH ahead of other Go installations.",
			version,
			shimDir,
		)
	}
	firstLine := u.FormatAction("Selected", "Go", version, "as the global default.")
	secondLine := "  " + u.theme.muted.Render(fmt.Sprintf("Add %s to PATH ahead of other Go installations.", shimDir))
	return firstLine + "\n" + secondLine
}

// FormatCurrentLine renders a single resolved tool row for `fgm current`.
func (u *uiRenderer) FormatCurrentLine(tool string, version string, label string, toolColWidth int) string {
	if !u.Out.Styled {
		line := fmt.Sprintf("%s %s", tool, version)
		if label != "" {
			line += " (" + label + ")"
		}
		return line
	}

	paddedTool := tool
	if !u.isNarrowOut() && toolColWidth > len(tool) {
		paddedTool = tool + strings.Repeat(" ", toolColWidth-len(tool))
	}
	head := u.theme.accent.Render(paddedTool) + "  " + u.theme.version.Render(version)
	if label == "" {
		return head
	}
	sourceText := u.theme.muted.Render("(" + label + ")")
	singleLine := head + "  " + sourceText
	if u.isNarrowOut() && ansi.StringWidth(singleLine) > u.Out.Width {
		return u.theme.accent.Render(tool) + " " + u.theme.version.Render(version) + "\n  " + sourceText
	}
	return singleLine
}

// FormatVersionLine renders a single row for `fgm versions`.
func (u *uiRenderer) FormatVersionLine(version string, marked bool, markerTag string) string {
	if !u.Out.Styled {
		if marked {
			return "* " + version
		}
		return version
	}
	if marked {
		line := u.theme.ok.Render("*") + " " + u.theme.version.Render(version)
		if markerTag != "" {
			line += "  " + u.theme.muted.Render("("+markerTag+")")
		}
		return line
	}
	return "  " + u.theme.accent.Render(version)
}

// FormatEmptyNotice renders an informational empty-state message.
func (u *uiRenderer) FormatEmptyNotice(message string) string {
	if !u.Out.Styled {
		return message
	}
	return u.theme.info.Render("i") + " " + u.theme.muted.Render(message)
}

// FormatError renders an error message for stderr in plain or styled TTY mode.
func (u *uiRenderer) FormatError(err error) string {
	if err == nil {
		return ""
	}
	if !u.Err.Styled {
		return err.Error()
	}
	badge := u.theme.err.Render("✗ ERROR")
	singleLine := badge + "  " + err.Error()
	if u.Err.Width > 0 && u.Err.Width < narrowTermThreshold && ansi.StringWidth(singleLine) > u.Err.Width {
		return badge + "\n  " + err.Error()
	}
	return singleLine
}

func fgmFangColorScheme(c lipgloss.LightDarkFunc) fang.ColorScheme {
	base := c(lipgloss.Color("#1F2937"), lipgloss.Color("#E5E7EB"))
	accent := c(lipgloss.Color("#007F7B"), lipgloss.Color("#2EC4B6"))
	version := c(lipgloss.Color("#005F73"), lipgloss.Color("#5BC0BE"))
	muted := c(lipgloss.Color("#5C677D"), lipgloss.Color("#8D99AE"))
	errColor := c(lipgloss.Color("#B91C1C"), lipgloss.Color("#EF4444"))

	return fang.ColorScheme{
		Base:           base,
		Title:          accent,
		Description:    base,
		Codeblock:      c(lipgloss.Color("#F3F4F6"), lipgloss.Color("#1E242B")),
		Program:        version,
		DimmedArgument: muted,
		Comment:        muted,
		Flag:           c(lipgloss.Color("#0F766E"), lipgloss.Color("#38BDF8")),
		FlagDefault:    muted,
		Command:        accent,
		QuotedString:   version,
		Argument:       base,
		Help:           base,
		Dash:           muted,
		ErrorHeader:    [2]color.Color{lipgloss.Color("#FFFFFF"), errColor},
		ErrorDetails:   errColor,
	}
}

type plainStripWriter struct {
	w io.Writer
}

func (p plainStripWriter) Write(b []byte) (int, error) {
	stripped := ansi.Strip(string(b))
	if _, err := io.WriteString(p.w, stripped); err != nil {
		return 0, err
	}
	return len(b), nil
}

type printedError struct {
	err error
}

func (e printedError) Error() string {
	return e.err.Error()
}

func (e printedError) Unwrap() error {
	return e.err
}

// IsPrintedError reports whether an error was already rendered to stderr.
func IsPrintedError(err error) bool {
	_, ok := errors.AsType[printedError](err)
	return ok
}

func buildFangHelpFunc() func(*cobra.Command, []string) {
	probe := &cobra.Command{
		Use: "fgm",
		Run: func(*cobra.Command, []string) {},
	}
	probe.SetOut(io.Discard)
	probe.SetErr(io.Discard)
	probe.SetArgs([]string{})
	_ = fang.Execute(
		context.Background(),
		probe,
		fang.WithoutCompletions(),
		fang.WithoutManpage(),
		fang.WithoutVersion(),
		fang.WithColorSchemeFunc(fgmFangColorScheme),
	)
	return probe.HelpFunc()
}

func configureFangHelp(root *cobra.Command, opts UIOptions) {
	root.SetHelpFunc(func(c *cobra.Command, args []string) {
		ui := uiForCommand(c, opts)
		origOut := c.OutOrStdout()
		if !ui.Out.Styled {
			c.SetOut(plainStripWriter{w: origOut})
			defer c.SetOut(origOut)
		}
		buildFangHelpFunc()(c, args)
	})
}

// Execute runs the root command with Fang help/error handling and single-owner error printing.
func Execute(ctx context.Context, root *cobra.Command, opts UIOptions) error {
	root.SilenceUsage = true
	root.SilenceErrors = true
	root.CompletionOptions.DisableDefaultCmd = true

	if err := root.ExecuteContext(withUIOptions(ctx, opts)); err != nil {
		ui := uiForCommand(root, opts)
		_, _ = fmt.Fprintln(root.ErrOrStderr(), ui.FormatError(err))
		return printedError{err: err}
	}
	return nil
}

// RootRunner adapts a root Cobra command to ExecuteContext using Fang.
type RootRunner struct {
	cmd  *cobra.Command
	opts UIOptions
}

// NewRootRunner constructs a RootRunner for the CLI entrypoint.
func NewRootRunner(root *cobra.Command, opts UIOptions) *RootRunner {
	return &RootRunner{cmd: root, opts: opts}
}

// ExecuteContext executes the root command with Fang and single-owner error output.
func (r *RootRunner) ExecuteContext(ctx context.Context) error {
	return Execute(ctx, r.cmd, r.opts)
}
