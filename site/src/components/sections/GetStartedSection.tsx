import { CodeSnippetCard } from "@/components/ui/CodeSnippetCard";
import { SectionShell } from "@/components/ui/SectionShell";
import {
  AI_SETUP_PROMPT,
  MODULE_QUICKSTART_CMDS,
} from "@/data/landing-content";

export function GetStartedSection() {
  return (
    <SectionShell
      id="get-started"
      kicker="Quick start"
      title="Let your AI agent set it up — or run 4 commands"
      description="Paste the prompt into Claude Code, Cursor, or Codex to install fgm, import existing Go binaries, and verify your repository automatically."
    >
      <div className="quickstart-grid">
        <CodeSnippetCard
          title="1. Set up with AI Agent (1-click prompt)"
          codeId="qs-ai-prompt"
          rawCode={AI_SETUP_PROMPT}
          copyAriaLabel="Copy AI agent setup prompt for Quick Start"
        >
          Installs <code>fgm</code>, sets up the{" "}
          <a href="https://github.com/kostikovk/fgm/blob/main/docs/ai-agents.md">
            agent skill (<code>skills/fgm/SKILL.md</code>)
          </a>
          , symlinks existing local Go/linter installs via{" "}
          <code>fgm import auto</code>, and verifies your repo.
        </CodeSnippetCard>

        <CodeSnippetCard
          title="2. Or run manually in a Go module"
          codeId="qs-module-cmds"
          rawCode={MODULE_QUICKSTART_CMDS}
          copyAriaLabel="Copy module quick-start commands"
        >
          For a multi-module workspace, use your repository&apos;s workspace
          check command or run these checks in each module directory. See{" "}
          <a href="https://github.com/kostikovk/fgm/blob/main/docs/getting-started.md">
            Getting started
          </a>
          .
        </CodeSnippetCard>
      </div>
    </SectionShell>
  );
}
