import { SectionShell } from "@/components/ui/SectionShell";
import { FAQ_ITEMS } from "@/data/landing-content";

export function FaqSection() {
  return (
    <SectionShell
      id="faq"
      kicker="FAQ"
      title="Technical questions and boundaries"
    >
      <div className="faq-list">
        {FAQ_ITEMS.map((item, idx) => (
          <details key={item.id} className="faq-item" open={idx === 0}>
            <summary>{item.question}</summary>
            <p>
              {item.answer}
              {item.docLink ? (
                <>
                  {" "}
                  See <a href={item.docLink.href}>{item.docLink.label}</a>.
                </>
              ) : null}
            </p>
          </details>
        ))}
      </div>
    </SectionShell>
  );
}
