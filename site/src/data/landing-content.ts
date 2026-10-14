export interface NavItem {
  readonly id: string;
  readonly href: string;
  readonly label: string;
  readonly prominent?: boolean;
}

export type TerminalLine =
  | { readonly id: string; readonly kind: "command"; readonly text: string }
  | {
      readonly id: string;
      readonly kind: "toolchain";
      readonly tool: string;
      readonly version: string;
      readonly source: string;
    }
  | {
      readonly id: string;
      readonly kind: "outcome";
      readonly action: string;
      readonly tool: string;
      readonly version: string;
      readonly suffix?: string;
    }
  | {
      readonly id: string;
      readonly kind: "doctor-ok";
      readonly message: string;
    }
  | { readonly id: string; readonly kind: "output"; readonly text: string }
  | { readonly id: string; readonly kind: "spacer" };

export interface TerminalScenario {
  readonly id: string;
  readonly tabId: string;
  readonly panelId: string;
  readonly label: string;
  readonly lines: readonly TerminalLine[];
}

export interface BenefitCard {
  readonly id: string;
  readonly badge: string;
  readonly title: string;
  readonly beforeText: string;
  readonly afterCommand: string;
  readonly summary: string;
}

export interface AudiencePill {
  readonly id: string;
  readonly title: string;
  readonly detail: string;
  readonly link?: {
    readonly href: string;
    readonly label: string;
  };
}

export interface ComparisonRowItem {
  readonly id: string;
  readonly approach: string;
  readonly existingStrength: string;
  readonly whenToLookAtFgm: string;
}

export interface FaqItem {
  readonly id: string;
  readonly question: string;
  readonly answer: string;
  readonly docLink?: {
    readonly href: string;
    readonly label: string;
  };
}

export const HERO_INSTALL_CMD = "$ brew install kostikovk/fgm/fgm";

export const AI_SETUP_PROMPT = `Set up Fast Go Manager (fgm) for this Go repository following https://github.com/kostikovk/fgm/blob/main/docs/ai-agents.md:
1. Install fgm if missing (brew install kostikovk/fgm/fgm or install.sh) and install the repo skill from skills/fgm/SKILL.md.
2. Run fgm import auto, fgm --chdir . current, and fgm --chdir . install for this repository's Go and compatible golangci-lint.
3. Audit lint config with fgm --chdir . lint doctor and verify with fgm --chdir . exec -- go version and fgm --chdir . exec -- go test ./...`;

export const MODULE_QUICKSTART_CMDS = `$ fgm current
$ fgm install
$ fgm exec -- go version
$ fgm exec -- go test ./...`;

export const LINT_QUICKSTART_CMDS = `# Import existing Go/linter installs & enable auto-switching on cd
$ fgm import auto
$ fgm doctor --fix

# Upgrade project Go + paired golangci-lint in one command
$ fgm upgrade go --project --with-lint
$ fgm lint doctor`;

export const NAV_ITEMS: readonly NavItem[] = [
  { id: "nav-why", href: "#why-fgm", label: "Benefits" },
  { id: "nav-how", href: "#how-it-works", label: "How it works" },
  { id: "nav-start", href: "#get-started", label: "Quick start" },
  { id: "nav-who", href: "#who-it-is-for", label: "Compare" },
  { id: "nav-faq", href: "#faq", label: "FAQ" },
  {
    id: "nav-docs",
    href: "https://github.com/kostikovk/fgm/blob/main/docs/README.md",
    label: "Docs",
  },
  {
    id: "nav-github",
    href: "https://github.com/kostikovk/fgm",
    label: "GitHub",
    prominent: true,
  },
];

export const TERMINAL_SCENARIOS: readonly TerminalScenario[] = [
  {
    id: "resolve",
    tabId: "tab-resolve",
    panelId: "panel-resolve",
    label: "1. Switch repos",
    lines: [
      {
        id: "res-1",
        kind: "command",
        text: "$ fgm doctor --fix",
      },
      {
        id: "res-2",
        kind: "doctor-ok",
        message: "Added shell integration to ~/.zshrc",
      },
      { id: "res-3", kind: "spacer" },
      {
        id: "res-4",
        kind: "command",
        text: "$ cd ~/work/legacy-api && go version",
      },
      {
        id: "res-5",
        kind: "output",
        text: "go version go1.23.6 darwin/arm64",
      },
      {
        id: "res-6",
        kind: "command",
        text: "$ cd ~/work/billing-svc && fgm current",
      },
      {
        id: "res-7",
        kind: "toolchain",
        tool: "go           ",
        version: "1.27.2",
        source: "(go.mod)",
      },
      {
        id: "res-8",
        kind: "toolchain",
        tool: "golangci-lint",
        version: "v2.14.0",
        source: "(local)",
      },
    ],
  },
  {
    id: "exec",
    tabId: "tab-exec",
    panelId: "panel-exec",
    label: "2. Run checks",
    lines: [
      {
        id: "ex-1",
        kind: "command",
        text: "$ fgm upgrade go --project --with-lint",
      },
      {
        id: "ex-2",
        kind: "outcome",
        action: "Upgraded project to",
        tool: "Go",
        version: "1.27.2",
        suffix: "in ./go.mod",
      },
      {
        id: "ex-3",
        kind: "outcome",
        action: "Installed",
        tool: "golangci-lint",
        version: "v2.14.0",
      },
      { id: "ex-4", kind: "spacer" },
      {
        id: "ex-5",
        kind: "command",
        text: "$ fgm exec -- go version",
      },
      {
        id: "ex-6",
        kind: "output",
        text: "go version go1.27.2 darwin/arm64",
      },
      {
        id: "ex-7",
        kind: "command",
        text: "$ fgm exec -- go test ./...",
      },
      {
        id: "ex-8",
        kind: "output",
        text: "ok  \texample.com/billing\t0.247s",
      },
    ],
  },
  {
    id: "doctor",
    tabId: "tab-doctor",
    panelId: "panel-doctor",
    label: "3. Doctor & audit",
    lines: [
      {
        id: "doc-1",
        kind: "command",
        text: "$ fgm import auto",
      },
      {
        id: "doc-2",
        kind: "outcome",
        action: "Imported",
        tool: "Go",
        version: "1.27.2",
        suffix: "from /opt/homebrew/opt/go/libexec",
      },
      { id: "doc-3", kind: "spacer" },
      {
        id: "doc-4",
        kind: "command",
        text: "$ fgm doctor",
      },
      {
        id: "doc-5",
        kind: "doctor-ok",
        message: "selected Go version is installed: 1.27.2",
      },
      {
        id: "doc-6",
        kind: "command",
        text: "$ fgm lint doctor",
      },
      {
        id: "doc-7",
        kind: "doctor-ok",
        message: "golangci-lint config looks good (.golangci.yml)",
      },
    ],
  },
];

export const BENEFIT_CARDS: readonly BenefitCard[] = [
  {
    id: "ben-upgrade",
    badge: "01 · 1-Command Upgrade",
    title: "Upgrade Go, go.mod, and golangci-lint together",
    beforeText:
      "Before: download Go from go.dev/dl → edit go.mod → golangci-lint breaks with 'unsupported Go version' → hunt for a compatible linter release.",
    afterCommand: "$ fgm upgrade go --project --with-lint",
    summary:
      "Downloads the target Go binary, updates your project's go.mod / go.work, and installs a compatible golangci-lint version in one step.",
  },
  {
    id: "ben-shims",
    badge: "02 · Zero-Config Switching",
    title: "cd into any repo — go & golangci-lint just work",
    beforeText:
      "Before: manually exporting PATH / GOROOT or running version-switch commands every time you jump between a Go 1.23 service and a Go 1.27 repo.",
    afterCommand: "$ fgm doctor --fix   # shims both go & golangci-lint",
    summary:
      "Routes plain go and golangci-lint commands to your repo's resolved versions automatically — no extra .tool-versions or .go-version file required.",
  },
  {
    id: "ben-lint-doctor",
    badge: "03 · Built-In Linter Brain",
    title: "Never guess which golangci-lint works with your Go",
    beforeText:
      "Before: generic version managers (mise, asdf, gvm) don't know which golangci-lint version supports your Go version or why .golangci.yml fails.",
    afterCommand: "$ fgm lint doctor && fgm pin golangci-lint v2.14.0",
    summary:
      "Includes an offline Go ↔ golangci-lint compatibility catalog, repo pinning in .fgm.toml, and YAML config diagnostics.",
  },
  {
    id: "ben-import-exec",
    badge: "04 · Instant Import & Agent-Ready",
    title: "Reuse existing installs in <1s & run deterministically",
    beforeText:
      "Before: re-downloading gigabytes of Go toolchains you already have on disk, or AI coding agents running the wrong global Go binary.",
    afterCommand: "$ fgm import auto && fgm exec -- go test ./...",
    summary:
      "Symlinks existing Homebrew/system Go & linter binaries without re-downloading, and provides fgm exec for CI and non-interactive AI agents.",
  },
];

export const AUDIENCE_PILLS: readonly AudiencePill[] = [
  {
    id: "pill-local-go",
    title: "No more go.dev/dl downloads",
    detail:
      "Install, switch, and upgrade local Go binaries directly from your terminal.",
  },
  {
    id: "pill-switching",
    title: "Multi-repo Go devs & CI",
    detail:
      "Keep Go and golangci-lint versions aligned across every service and CI pipeline.",
  },
  {
    id: "pill-agents",
    title: "AI coding agents",
    detail: "Run deterministic checks via fgm --chdir <dir> exec -- with the",
    link: {
      href: "https://github.com/kostikovk/fgm/blob/main/docs/ai-agents.md",
      label: "agent skill",
    },
  },
];

export const COMPARISON_ROWS: readonly ComparisonRowItem[] = [
  {
    id: "cmp-manual-go",
    approach: "go.dev/dl tarballs / brew install go",
    existingStrength: "Installs one machine-wide Go & linter binary.",
    whenToLookAtFgm:
      "Upgrading Go breaks older repos or golangci-lint; FGM keeps multiple versions side by side per repo.",
  },
  {
    id: "cmp-native-go",
    approach: "Native Go (GOTOOLCHAIN)",
    existingStrength: "Switches Go compiler from go.mod / go.work.",
    whenToLookAtFgm:
      "Doesn't manage golangci-lint compatibility, update go.mod on upgrade, or audit .golangci.yml.",
  },
  {
    id: "cmp-mise-asdf",
    approach: "mise / asdf / gvm",
    existingStrength: "General or shell-based version managers.",
    whenToLookAtFgm:
      "Require extra config files or plugins and don't know which golangci-lint version matches your Go.",
  },
  {
    id: "cmp-fgm",
    approach: "FGM (Single Go Binary)",
    existingStrength:
      "Native go.work / go.mod resolution + Go ↔ golangci-lint catalog + 1-command project upgrades + lint doctor.",
    whenToLookAtFgm:
      "You want one fast, Go-native CLI that handles both Go versions and golangci-lint without friction.",
  },
];

export const FAQ_ITEMS: readonly FaqItem[] = [
  {
    id: "faq-gotoolchain",
    question: "Doesn't Go already manage toolchains?",
    answer:
      "Native Go toolchain selection handles Go when a base Go binary is already installed. FGM lets you install, switch, and upgrade local Go binaries from scratch, updates go.mod / go.work with fgm upgrade go --project, pairs a compatible golangci-lint per repo, and audits .golangci.yml. Note that native GOTOOLCHAIN settings can still influence the compiler Go invokes; check with fgm exec -- go version.",
  },
  {
    id: "faq-version-file",
    question: "Does FGM require another Go version file?",
    answer:
      "No. It reads go.work and go.mod directly. .fgm.toml is optional and only needed to pin a repo-specific golangci-lint policy.",
  },
  {
    id: "faq-auto-pin",
    question: "Does auto mean the same linter version everywhere?",
    answer:
      "No. In auto mode, FGM prefers an already-installed compatible linter before falling back to the catalog recommendation. For identical versions across machines and CI, run fgm pin golangci-lint <version>.",
  },
  {
    id: "faq-exec-install",
    question: "Does fgm exec install missing tools automatically?",
    answer:
      "No. Run fgm install (or fgm import auto) first. If the selected linter isn't installed in FGM, fgm exec still runs other commands and may fall back to PATH.",
  },
  {
    id: "faq-windows",
    question: "Can I use FGM on Windows?",
    answer:
      "Yes — release archives and fgm exec work on macOS, Linux, and Windows. Automatic shell shims (shims/go and shims/golangci-lint) are macOS/Linux only.",
  },
  {
    id: "faq-agents",
    question: "Can an AI coding agent use FGM?",
    answer:
      "Yes. Use fgm --chdir <dir> exec -- in non-interactive shells, or load the standalone skill (skills/fgm/SKILL.md).",
    docLink: {
      href: "https://github.com/kostikovk/fgm/blob/main/docs/ai-agents.md",
      label: "docs/ai-agents.md",
    },
  },
  {
    id: "faq-existing-config",
    question: "Can I keep my existing golangci-lint configuration?",
    answer:
      "Yes. fgm lint doctor audits your existing .golangci.yml / .golangci.yaml without modifying it.",
  },
  {
    id: "faq-upgrade-fgm",
    question: "How do I upgrade FGM itself to the latest version?",
    answer:
      "Run brew update && brew upgrade --cask fgm (for Homebrew) or re-run the install.sh script. FGM also checks for newer CLI releases during fgm doctor / install / upgrade and notifies you in interactive terminals (set FGM_NO_UPDATE_NOTIFIER=1 to disable).",
    docLink: {
      href: "https://github.com/kostikovk/fgm/blob/main/docs/installation.md#upgrade-fgm",
      label: "docs/installation.md#upgrade-fgm",
    },
  },
];
