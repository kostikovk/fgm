"use client";

import { useState, useRef, useEffect } from "react";
import { sanitizeSnippet } from "@/lib/snippet";

interface CopyButtonProps {
  rawText: string;
  ariaLabel: string;
}

export function CopyButton({ rawText, ariaLabel }: CopyButtonProps) {
  const [copyState, setCopyState] = useState<"idle" | "copied" | "error">("idle");
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const requestIdRef = useRef(0);

  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  async function handleCopy() {
    const requestId = ++requestIdRef.current;
    if (timerRef.current) {
      clearTimeout(timerRef.current);
    }

    const cleanText = sanitizeSnippet(rawText);
    const liveStatus =
      typeof document !== "undefined"
        ? document.getElementById("copy-live-status")
        : null;

    if (
      typeof navigator === "undefined" ||
      !navigator.clipboard ||
      typeof navigator.clipboard.writeText !== "function"
    ) {
      if (requestId !== requestIdRef.current) {
        return;
      }
      setCopyState("error");
      if (liveStatus) {
        liveStatus.textContent =
          "Clipboard unavailable — select and copy manually.";
      }
      timerRef.current = setTimeout(() => {
        if (requestId === requestIdRef.current) {
          setCopyState("idle");
        }
      }, 2600);
      return;
    }

    try {
      await navigator.clipboard.writeText(cleanText);
      if (requestId !== requestIdRef.current) {
        return;
      }
      setCopyState("copied");
      if (liveStatus) {
        liveStatus.textContent = "Copied command to clipboard.";
      }
    } catch {
      if (requestId !== requestIdRef.current) {
        return;
      }
      setCopyState("error");
      if (liveStatus) {
        liveStatus.textContent =
          "Clipboard permission denied — select and copy manually.";
      }
    }

    timerRef.current = setTimeout(() => {
      if (requestId === requestIdRef.current) {
        setCopyState("idle");
      }
    }, 2600);
  }

  const label =
    copyState === "copied"
      ? "✓ Copied"
      : copyState === "error"
        ? "Select text"
        : "Copy";

  return (
    <button
      type="button"
      className="copy-btn"
      data-copy-state={copyState}
      aria-label={ariaLabel}
      onClick={handleCopy}
    >
      {label}
    </button>
  );
}
