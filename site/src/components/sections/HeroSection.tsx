import { CopyButton } from "@/components/CopyButton";
import { TerminalStudio } from "@/components/TerminalStudio";
import { AI_SETUP_PROMPT, HERO_INSTALL_CMD } from "@/data/landing-content";

export function HeroSection() {
  return (
    <section id="top" className="hero">
      <div className="container hero-grid">
        <div>
          <div className="eyebrow">
            <span className="eyebrow-dot" aria-hidden="true" />
            <span>Fast Go Manager · Single Binary CLI</span>
          </div>
          <h1>
            Your project&apos;s Go and{" "}
            <span className="text-gradient">golangci-lint</span>, managed
            together.
          </h1>
          <p className="hero-lead">
            Stop downloading Go tarballs from <code>go.dev</code> or debugging{" "}
            <code>unsupported Go version</code> in <code>golangci-lint</code>.
            Install, auto-switch on <code>cd</code>, and upgrade Go + your
            repo&apos;s compatible linter in one command.
          </p>

          <ul className="hero-highlights" aria-label="Key FGM benefits">
            <li>
              <span className="hl-check">✓</span>
              <span>
                <strong>1-command upgrade:</strong>{" "}
                <code>fgm upgrade go --project --with-lint</code> updates Go,{" "}
                <code>go.mod</code>, and <code>golangci-lint</code> together
              </span>
            </li>
            <li>
              <span className="hl-check">✓</span>
              <span>
                <strong>Zero extra config files:</strong> reads existing{" "}
                <code>go.work</code> &amp; <code>go.mod</code> directly
              </span>
            </li>
            <li>
              <span className="hl-check">✓</span>
              <span>
                <strong>AI Agent ready:</strong> paste one prompt into Claude
                Code, Cursor, or Codex to install &amp; configure everything
              </span>
            </li>
          </ul>

          <div className="cta-row">
            <a className="btn btn-primary" href="#get-started">
              Set up with AI or CLI
            </a>
            <a
              className="btn btn-secondary"
              href="https://github.com/kostikovk/fgm"
            >
              View on GitHub
            </a>
          </div>

          <div className="install-card">
            <div className="install-meta">
              <span>
                <strong>AI Agent Setup</strong> (Claude Code, Cursor, Codex)
              </span>
              <a href="#get-started">Full prompt &rarr;</a>
            </div>
            <div className="code-bar">
              <code id="hero-ai-prompt">
                <span className="prompt-sign">&gt;</span>Set up Fast Go Manager
                (fgm) for this Go repository...
              </code>
              <CopyButton
                rawText={AI_SETUP_PROMPT}
                ariaLabel="Copy AI agent setup prompt"
              />
            </div>

            <div className="install-meta install-meta-spaced">
              <span>Or install via Homebrew (macOS &amp; Linux)</span>
              <a href="https://github.com/kostikovk/fgm/blob/main/docs/installation.md">
                Other options &rarr;
              </a>
            </div>
            <div className="code-bar">
              <code id="hero-install-cmd">
                <span className="prompt-sign">$</span>brew install
                kostikovk/fgm/fgm
              </code>
              <CopyButton
                rawText={HERO_INSTALL_CMD}
                ariaLabel="Copy Homebrew install command"
              />
            </div>
          </div>
        </div>

        <TerminalStudio />
      </div>
    </section>
  );
}
