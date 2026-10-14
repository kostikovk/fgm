"use client";

import type { ReactNode, PointerEvent } from "react";

interface SpotlightCardProps {
  tag?: string;
  title: string;
  children: ReactNode;
}

export function SpotlightCard({ tag, title, children }: SpotlightCardProps) {
  function handlePointerMove(event: PointerEvent<HTMLElement>) {
    const rect = event.currentTarget.getBoundingClientRect();
    event.currentTarget.style.setProperty(
      "--mouse-x",
      `${event.clientX - rect.left}px`
    );
    event.currentTarget.style.setProperty(
      "--mouse-y",
      `${event.clientY - rect.top}px`
    );
  }

  return (
    <article className="spotlight-card" onPointerMove={handlePointerMove}>
      {tag ? <span className="card-tag">{tag}</span> : null}
      <h3>{title}</h3>
      <p>{children}</p>
    </article>
  );
}
