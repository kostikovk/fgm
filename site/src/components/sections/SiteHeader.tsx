"use client";

import type { MouseEvent } from "react";
import { NAV_ITEMS } from "@/data/landing-content";

export function SiteHeader() {
  function handleAnchorClick(
    event: MouseEvent<HTMLAnchorElement>,
    href: string
  ) {
    if (!href.startsWith("#")) {
      return;
    }
    const targetId = href.slice(1);
    const target = document.getElementById(targetId);
    if (!target) {
      return;
    }
    event.preventDefault();

    const header = document.querySelector(".site-header");
    const headerHeight = header ? header.getBoundingClientRect().height : 0;
    const targetTop =
      target.getBoundingClientRect().top + window.scrollY - headerHeight - 12;

    const prefersReducedMotion =
      typeof window.matchMedia === "function" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    window.history.pushState(null, "", href);
    window.scrollTo({
      top: Math.max(0, targetTop),
      behavior: prefersReducedMotion ? "auto" : "smooth",
    });
  }

  return (
    <header className="site-header">
      <div className="container header-inner">
        <a
          className="brand"
          href="#top"
          aria-label="FGM Home"
          onClick={(e) => handleAnchorClick(e, "#top")}
        >
          <span className="brand-mark">&gt;_</span>
          <span>fgm</span>
          <span className="brand-tag">CLI</span>
        </a>
        <nav className="site-nav" aria-label="Primary">
          {NAV_ITEMS.map((item) => (
            <a
              key={item.id}
              href={item.href}
              className={item.prominent ? "nav-github" : undefined}
              onClick={(e) => handleAnchorClick(e, item.href)}
            >
              {item.label}
            </a>
          ))}
        </nav>
      </div>
    </header>
  );
}
