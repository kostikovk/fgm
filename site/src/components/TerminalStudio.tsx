"use client";

import { useState, useRef, useEffect, type KeyboardEvent } from "react";
import {
  TERMINAL_SCENARIOS,
  type TerminalLine,
} from "@/data/landing-content";

function TerminalLineRow({
  line,
  lineIndex,
  isRevealed,
}: {
  line: TerminalLine;
  lineIndex: number;
  isRevealed: boolean;
}) {
  const baseClass = `term-line${isRevealed ? " is-revealed" : ""}`;
  const style = { ["--line-index" as string]: lineIndex };

  switch (line.kind) {
    case "command":
      return (
        <div className={`${baseClass} t-cmd`} style={style}>
          {line.text}
        </div>
      );
    case "toolchain":
      return (
        <div className={baseClass} style={style}>
          <span className="t-tool">{line.tool}</span>
          {"  "}
          <span className="t-ver">{line.version}</span>
          {"   "}
          <span className="t-muted">{line.source}</span>
        </div>
      );
    case "outcome":
      return (
        <div className={baseClass} style={style}>
          <span className="t-ok">✓</span> {line.action}{" "}
          <span className="t-tool">{line.tool}</span>{" "}
          <span className="t-ver">{line.version}</span>
          {line.suffix ? (
            <>
              {" "}
              <span className="t-muted">{line.suffix}</span>
            </>
          ) : null}
        </div>
      );
    case "doctor-ok":
      return (
        <div className={baseClass} style={style}>
          <span className="t-ok">✓ OK</span>
          {"  "}
          {line.message}
        </div>
      );
    case "output":
      return (
        <div className={baseClass} style={style}>
          {line.text}
        </div>
      );
    case "spacer":
      return (
        <div className={baseClass} style={style}>
          &nbsp;
        </div>
      );
  }
}

export function TerminalStudio() {
  const [activeIndex, setActiveIndex] = useState(0);
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([]);

  useEffect(() => {
    document.documentElement.classList.add("js-ready");
  }, []);

  function selectTab(nextIndex: number, focusTab: boolean) {
    const count = TERMINAL_SCENARIOS.length;
    const wrapped = ((nextIndex % count) + count) % count;
    setActiveIndex(wrapped);
    if (focusTab) {
      tabRefs.current[wrapped]?.focus();
    }
  }

  function handleKeyDown(
    index: number,
    event: KeyboardEvent<HTMLButtonElement>
  ) {
    if (event.key === "ArrowRight") {
      event.preventDefault();
      selectTab(index + 1, true);
    } else if (event.key === "ArrowLeft") {
      event.preventDefault();
      selectTab(index - 1, true);
    } else if (event.key === "Home") {
      event.preventDefault();
      selectTab(0, true);
    } else if (event.key === "End") {
      event.preventDefault();
      selectTab(TERMINAL_SCENARIOS.length - 1, true);
    }
  }

  return (
    <figure
      className="terminal-studio"
      aria-label="Verified FGM terminal workflow examples"
    >
      <div className="terminal-topbar">
        <div className="window-dots" aria-hidden="true">
          <span />
          <span />
          <span />
        </div>
        <div
          className="terminal-tabs"
          role="tablist"
          aria-label="Terminal workflow scenarios"
        >
          {TERMINAL_SCENARIOS.map((scenario, index) => {
            const isSelected = index === activeIndex;
            return (
              <button
                key={scenario.id}
                ref={(el) => {
                  tabRefs.current[index] = el;
                }}
                id={scenario.tabId}
                type="button"
                role="tab"
                className={`term-tab${isSelected ? " is-active" : ""}`}
                aria-selected={isSelected ? "true" : "false"}
                aria-controls={scenario.panelId}
                tabIndex={isSelected ? 0 : -1}
                onClick={() => selectTab(index, false)}
                onKeyDown={(e) => handleKeyDown(index, e)}
              >
                {scenario.label}
              </button>
            );
          })}
        </div>
      </div>

      <div className="terminal-body">
        {TERMINAL_SCENARIOS.map((scenario, panelIndex) => {
          const isSelected = panelIndex === activeIndex;
          return (
            <div
              key={scenario.id}
              id={scenario.panelId}
              className={`term-panel${isSelected ? " is-active" : ""}`}
              role="tabpanel"
              aria-labelledby={scenario.tabId}
              hidden={!isSelected}
            >
              {scenario.lines.map((line, lineIndex) => (
                <TerminalLineRow
                  key={line.id}
                  line={line}
                  lineIndex={lineIndex}
                  isRevealed={isSelected}
                />
              ))}
            </div>
          );
        })}
      </div>

      <figcaption className="terminal-caption">
        <span>Single static Go binary · macOS, Linux &amp; Windows</span>
        <span>TTY styled · plain text in pipes &amp; CI</span>
      </figcaption>
    </figure>
  );
}
