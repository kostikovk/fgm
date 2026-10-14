import { test, expect } from "@playwright/test";

test.describe("FGM Landing Page (/fgm/)", () => {
  test("should render hero, why-fgm, how-it-works, who-it-is-for, get-started, and faq sections under /fgm/ without console or page errors when page loads", async ({
    page,
  }) => {
    // Arrange
    const errors: string[] = [];
    page.on("pageerror", (err) => errors.push(err.message));
    page.on("console", (msg) => {
      if (msg.type() === "error") {
        errors.push(msg.text());
      }
    });

    // Act
    const response = await page.goto("/fgm/");

    // Assert
    expect(response?.status()).toBe(200);
    await expect(page.locator("h1")).toContainText(
      "Your project's Go and golangci-lint, managed together."
    );
    await expect(page.locator("#why-fgm")).toBeVisible();
    await expect(page.locator("#how-it-works")).toBeVisible();
    await expect(page.locator("#who-it-is-for")).toBeVisible();
    await expect(page.locator("#get-started")).toBeVisible();
    await expect(page.locator("#faq")).toBeVisible();
    await expect(page.locator("#get-started")).toContainText(
      "For a multi-module workspace, use your repository's workspace check command or run these checks in each module directory."
    );
    expect(errors).toEqual([]);
  });

  test("should switch terminal studio panels on click and ArrowRight/ArrowLeft keyboard navigation when interacting with scenario tabs", async ({
    page,
  }) => {
    // Arrange
    await page.goto("/fgm/");
    const tabResolve = page.getByRole("tab", { name: /1\. Switch repos/i });
    const tabExec = page.getByRole("tab", { name: /2\. Run checks/i });
    const tabDoctor = page.getByRole("tab", { name: /3\. Doctor & audit/i });
    const panelResolve = page.locator("#panel-resolve");
    const panelExec = page.locator("#panel-exec");
    const panelDoctor = page.locator("#panel-doctor");

    // Act & Assert — initial state
    await expect(tabResolve).toHaveAttribute("aria-selected", "true");
    await expect(panelResolve).toBeVisible();
    await expect(panelExec).toBeHidden();
    await expect(panelResolve.locator(".term-line").first()).toHaveClass(
      /is-revealed/
    );
    await expect(panelExec.locator(".term-line").first()).not.toHaveClass(
      /is-revealed/
    );

    // Act — click second tab
    await tabExec.click();
    await expect(tabExec).toHaveAttribute("aria-selected", "true");
    await expect(panelExec).toBeVisible();
    await expect(panelResolve).toBeHidden();
    await expect(panelExec.locator(".term-line").first()).toHaveClass(
      /is-revealed/
    );
    await expect(panelResolve.locator(".term-line").first()).not.toHaveClass(
      /is-revealed/
    );

    // Act — ArrowRight moves to third tab
    await tabExec.press("ArrowRight");
    await expect(tabDoctor).toHaveAttribute("aria-selected", "true");
    await expect(panelDoctor).toBeVisible();

    // Act — ArrowLeft moves back to second tab
    await tabDoctor.press("ArrowLeft");
    await expect(tabExec).toHaveAttribute("aria-selected", "true");
    await expect(panelExec).toBeVisible();
  });

  test("should copy clean command text without prompt symbols and display Copied state when clipboard writeText succeeds", async ({
    page,
  }) => {
    // Arrange
    await page.addInitScript(() => {
      (window as unknown as { __copiedText?: string }).__copiedText = "";
      Object.defineProperty(navigator, "clipboard", {
        configurable: true,
        value: {
          writeText: async (text: string) => {
            (window as unknown as { __copiedText?: string }).__copiedText = text;
          },
        },
      });
    });
    await page.goto("/fgm/");
    const copyBtn = page.getByRole("button", {
      name: /Copy module quick-start commands/i,
    });

    // Act
    await copyBtn.click();

    // Assert
    await expect(copyBtn).toHaveText("✓ Copied");
    await expect(copyBtn).toHaveAttribute("data-copy-state", "copied");
    const copied = await page.evaluate(
      () => (window as unknown as { __copiedText?: string }).__copiedText
    );
    expect(copied).toBe(
      "fgm current\nfgm install\nfgm exec -- go version\nfgm exec -- go test ./..."
    );
  });

  test("should show manual-copy fallback state and not show Copied when clipboard writeText rejects with NotAllowedError", async ({
    page,
  }) => {
    // Arrange
    await page.addInitScript(() => {
      Object.defineProperty(navigator, "clipboard", {
        configurable: true,
        value: {
          writeText: async () => {
            const err = new Error("Write permission denied.");
            err.name = "NotAllowedError";
            throw err;
          },
        },
      });
    });
    await page.goto("/fgm/");
    const copyBtn = page.getByRole("button", {
      name: /Copy Homebrew install command/i,
    });

    // Act
    await copyBtn.click();

    // Assert
    await expect(copyBtn).toHaveAttribute("data-copy-state", "error");
    await expect(copyBtn).toHaveText("Select text");
    await expect(page.locator("#copy-live-status")).toContainText(
      /select and copy manually/i
    );
  });

  test("should ignore stale clipboard rejection when a newer click resolves first", async ({
    page,
  }) => {
    // Arrange
    await page.addInitScript(() => {
      let callCount = 0;
      Object.defineProperty(navigator, "clipboard", {
        configurable: true,
        value: {
          writeText: () => {
            callCount += 1;
            if (callCount === 1) {
              return new Promise((_, reject) => {
                setTimeout(() => reject(new Error("Stale failure")), 120);
              });
            }
            return Promise.resolve();
          },
        },
      });
    });
    await page.goto("/fgm/");
    const copyBtn = page.getByRole("button", {
      name: /Copy Homebrew install command/i,
    });

    // Act
    await copyBtn.click();
    await copyBtn.click();
    await page.waitForTimeout(200);

    // Assert
    await expect(copyBtn).toHaveAttribute("data-copy-state", "copied");
    await expect(copyBtn).toHaveText("✓ Copied");
  });

  test("should render all terminal scenarios and command snippets readably when JavaScript is disabled", async ({
    browser,
  }) => {
    // Arrange
    const context = await browser.newContext({ javaScriptEnabled: false });
    const page = await context.newPage();

    // Act
    await page.goto("http://127.0.0.1:4173/fgm/");

    // Assert
    await expect(page.locator("h1")).toContainText(
      "Your project's Go and golangci-lint, managed together."
    );
    await expect(page.locator("#panel-resolve")).toBeVisible();
    await expect(page.locator("#panel-exec")).toBeVisible();
    await expect(page.locator("#panel-doctor")).toBeVisible();
    await expect(page.locator("#hero-install-cmd")).toContainText(
      "brew install kostikovk/fgm/fgm"
    );
    await context.close();
  });

  test("should fit within 360px mobile and 1280px desktop viewports without horizontal page overflow when resizing viewport", async ({
    page,
  }) => {
    for (const width of [360, 1280]) {
      // Arrange
      await page.setViewportSize({ width, height: 900 });

      // Act
      await page.goto("/fgm/");

      // Assert
      const overflow = await page.evaluate(() => {
        return (
          document.documentElement.scrollWidth >
          document.documentElement.clientWidth
        );
      });
      expect(overflow, `horizontal overflow at width=${width}`).toBe(false);
    }
  });

  test("should have valid internal anchor targets, 200 OK static assets under /fgm/, single h1, and no prohibited marketing claims when auditing DOM", async ({
    page,
    request,
  }) => {
    // Arrange & Act
    await page.goto("/fgm/");

    // Assert — single h1
    await expect(page.locator("h1")).toHaveCount(1);

    // Assert — all internal #anchors resolve to existing id attributes
    const brokenAnchors = await page.evaluate(() => {
      const links = Array.from(
        document.querySelectorAll<HTMLAnchorElement>('a[href^="#"]')
      );
      return links
        .map((a) => a.getAttribute("href") || "")
        .filter((href) => {
          const id = href.slice(1);
          return !id || !document.getElementById(id);
        });
    });
    expect(brokenAnchors).toEqual([]);

    // Assert — static assets under /fgm/assets/ return 200 OK
    for (const assetPath of [
      "/fgm/assets/favicon.svg",
      "/fgm/assets/social.png",
    ]) {
      const res = await request.get(assetPath);
      expect(res.status(), `status for ${assetPath}`).toBe(200);
    }

    // Assert — no prohibited marketing claims
    const bodyText = (await page.locator("body").innerText()).toLowerCase();
    for (const forbidden of [
      "fastest go manager",
      "zero configuration",
      "always compatible",
      "identical everywhere",
      "production-proven",
    ]) {
      expect(bodyText).not.toContain(forbidden);
    }

    // Assert — fixed animated canvas background is active and sections do not slice the grid with border-top
    const canvasInfo = await page.evaluate(() => {
      const canvas = document.getElementById(
        "grid-background-canvas"
      ) as HTMLCanvasElement | null;
      if (!canvas) {
        return null;
      }
      const style = window.getComputedStyle(canvas);
      const section = document.querySelector(".section");
      const sectionStyle = section ? window.getComputedStyle(section) : null;
      return {
        position: style.position,
        pointerEvents: style.pointerEvents,
        width: canvas.width,
        height: canvas.height,
        sectionBorderTopWidth: sectionStyle?.borderTopWidth ?? "none",
      };
    });
    expect(canvasInfo).not.toBeNull();
    expect(canvasInfo?.position).toBe("fixed");
    expect(canvasInfo?.pointerEvents).toBe("none");
    expect(canvasInfo?.width).toBeGreaterThan(0);
    expect(canvasInfo?.height).toBeGreaterThan(0);
    expect(canvasInfo?.sectionBorderTopWidth).toBe("0px");
  });

  test("should scroll each header section link cleanly below the sticky header when clicked", async ({
    page,
  }) => {
    // Arrange
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.goto("/fgm/");

    for (const hash of [
      "#why-fgm",
      "#how-it-works",
      "#get-started",
      "#who-it-is-for",
      "#faq",
    ]) {
      // Act
      await page.locator(`.site-nav a[href="${hash}"]`).click();
      await expect
        .poll(
          async () => {
            return page.evaluate((targetHash) => {
              const header = document.querySelector(".site-header");
              const target = document.querySelector(targetHash);
              if (!header || !target) {
                return false;
              }
              const gap =
                target.getBoundingClientRect().top -
                header.getBoundingClientRect().bottom;
              return gap >= 4 && gap <= 36;
            }, hash);
          },
          { timeout: 3000 }
        )
        .toBe(true);

      // Assert
      const metrics = await page.evaluate((targetHash) => {
        const header = document.querySelector(".site-header");
        const target = document.querySelector(targetHash);
        const heading = target?.querySelector("h2, h3");
        if (!header || !target || !heading) {
          return null;
        }
        const headerBottom = header.getBoundingClientRect().bottom;
        const targetTop = target.getBoundingClientRect().top;
        return {
          headerBottom: Math.round(headerBottom),
          targetTop: Math.round(targetTop),
        };
      }, hash);

      expect(metrics, `metrics for ${hash}`).not.toBeNull();
      expect(
        metrics!.targetTop,
        `${hash} top (${metrics?.targetTop}) should not be hidden under sticky header bottom (${metrics?.headerBottom})`
      ).toBeGreaterThanOrEqual(metrics!.headerBottom);
      expect(
        metrics!.targetTop,
        `${hash} top (${metrics?.targetTop}) should land near sticky header bottom (${metrics?.headerBottom})`
      ).toBeLessThanOrEqual(metrics!.headerBottom + 48);
    }
  });
});
