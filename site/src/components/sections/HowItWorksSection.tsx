import { SectionShell } from "@/components/ui/SectionShell";

export function HowItWorksSection() {
  return (
    <SectionShell
      id="how-it-works"
      kicker="How it works"
      title="Under the hood: 3-step resolution pipeline"
      description={
        <>
          FGM resolves your repository&apos;s Go version, pairs a compatible{" "}
          <code>golangci-lint</code>, and executes checks without mutating your
          global environment.
        </>
      }
    >
      <div className="pipeline-box">
        <svg
          className="pulse-line"
          viewBox="0 0 800 18"
          aria-hidden="true"
          preserveAspectRatio="none"
        >
          <line
            x1="0"
            y1="9"
            x2="800"
            y2="9"
            stroke="rgba(148,163,184,0.2)"
            strokeWidth="2"
          />
          <line className="pulse-stroke" x1="0" y1="9" x2="800" y2="9" />
        </svg>

        <ol className="pipeline-steps">
          <li className="pipeline-step">
            <span className="pipeline-step-num">01 / READ</span>
            <strong>go.work &rarr; go.mod &rarr; global</strong>
            <span>
              Detects the target Go version (<code>toolchain</code> directive
              first) without custom Go version files.
            </span>
          </li>
          <li className="pipeline-step">
            <span className="pipeline-step-num">02 / PAIR</span>
            <strong>Catalog or .fgm.toml pin</strong>
            <span>
              Selects a compatible <code>golangci-lint</code> from the built-in
              catalog or your repo&apos;s explicit pin.
            </span>
          </li>
          <li className="pipeline-step">
            <span className="pipeline-step-num">03 / RUN</span>
            <strong>Shims or fgm exec</strong>
            <span>
              Routes <code>go</code> and <code>golangci-lint</code>{" "}
              automatically in interactive shells or runs deterministically via{" "}
              <code>fgm exec --</code>.
            </span>
          </li>
        </ol>
      </div>
    </SectionShell>
  );
}
