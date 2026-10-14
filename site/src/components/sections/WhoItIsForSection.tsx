import { SectionShell } from "@/components/ui/SectionShell";
import { AUDIENCE_PILLS, COMPARISON_ROWS } from "@/data/landing-content";

export function WhoItIsForSection() {
  return (
    <SectionShell
      id="who-it-is-for"
      kicker="When to use FGM"
      title="Honest comparison with existing tools"
      description={
        <>
          If you work in a single Go repo where native <code>GOTOOLCHAIN</code>{" "}
          already covers you, you may not need FGM. For polyglot setups across
          Node/Python/Rust, compare{" "}
          <a href="https://mise.jdx.dev/lang/go.html">mise</a> or{" "}
          <a href="https://asdf-vm.com/manage/configuration.html">asdf</a>.
        </>
      }
    >
      <div className="audience-strip">
        {AUDIENCE_PILLS.map((pill) => (
          <div key={pill.id} className="audience-pill">
            <strong>{pill.title}</strong>
            <span>
              {pill.detail}
              {pill.link ? (
                <>
                  {" "}
                  <a href={pill.link.href}>{pill.link.label}</a>.
                </>
              ) : null}
            </span>
          </div>
        ))}
      </div>

      <div
        className="table-wrap"
        role="region"
        aria-label="Comparison with alternatives"
        tabIndex={0}
      >
        <table>
          <thead>
            <tr>
              <th scope="col">Approach</th>
              <th scope="col">What it already does</th>
              <th scope="col">When to look at FGM</th>
            </tr>
          </thead>
          <tbody>
            {COMPARISON_ROWS.map((row) => (
              <tr key={row.id}>
                <td>
                  <strong>{row.approach}</strong>
                </td>
                <td>{row.existingStrength}</td>
                <td>{row.whenToLookAtFgm}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </SectionShell>
  );
}
