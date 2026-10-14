import { SpotlightCard } from "@/components/SpotlightCard";
import { SectionShell } from "@/components/ui/SectionShell";
import { BENEFIT_CARDS } from "@/data/landing-content";

export function WhyFgmSection() {
  return (
    <SectionShell
      id="why-fgm"
      kicker="Why use FGM"
      title="4 things FGM does that Go, brew, and generic managers don't"
      description={
        <>
          <code>brew upgrade go</code> breaks older repos.{" "}
          <code>GOTOOLCHAIN</code> and <code>mise</code> don&apos;t know which{" "}
          <code>golangci-lint</code> version supports your Go release. FGM
          bridges both.
        </>
      }
    >
      <div className="bento-grid">
        {BENEFIT_CARDS.map((card) => (
          <SpotlightCard key={card.id} tag={card.badge} title={card.title}>
            <span className="benefit-before">{card.beforeText}</span>
            <code className="benefit-cmd">{card.afterCommand}</code>
            <span className="benefit-summary">{card.summary}</span>
          </SpotlightCard>
        ))}
      </div>
    </SectionShell>
  );
}
