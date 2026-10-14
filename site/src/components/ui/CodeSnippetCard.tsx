import type { ReactNode } from "react";
import { CopyButton } from "@/components/CopyButton";

interface CodeSnippetCardProps {
  title: string;
  codeId: string;
  rawCode: string;
  copyAriaLabel: string;
  children?: ReactNode;
}

export function CodeSnippetCard({
  title,
  codeId,
  rawCode,
  copyAriaLabel,
  children,
}: CodeSnippetCardProps) {
  return (
    <div className="snippet-card">
      <div className="snippet-header">
        <h3>{title}</h3>
        <CopyButton rawText={rawCode} ariaLabel={copyAriaLabel} />
      </div>
      <pre className="pre-block">
        <code id={codeId}>{rawCode}</code>
      </pre>
      {children ? <div className="caveat-note">{children}</div> : null}
    </div>
  );
}
