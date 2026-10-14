import type { ReactNode } from "react";

interface SectionShellProps {
  id: string;
  kicker: string;
  title: string;
  description?: ReactNode;
  children: ReactNode;
}

export function SectionShell({
  id,
  kicker,
  title,
  description,
  children,
}: SectionShellProps) {
  return (
    <section id={id} className="section">
      <div className="container">
        <div className="section-header">
          <div className="section-kicker">{kicker}</div>
          <h2>{title}</h2>
          {description ? <p className="section-desc">{description}</p> : null}
        </div>
        {children}
      </div>
    </section>
  );
}
