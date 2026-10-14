"use client";

import { useEffect, useRef } from "react";

const GRID_SIZE = 44;

interface GridPulse {
  horizontal: boolean;
  lineIndex: number;
  pos: number;
  speed: number;
  length: number;
}

export function GridBackground() {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) {
      return;
    }
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      return;
    }

    let width = window.innerWidth;
    let height = window.innerHeight;
    let rafId = 0;

    const mediaQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    let reducedMotion = mediaQuery.matches;

    const pulses: GridPulse[] = [
      { horizontal: true, lineIndex: 3, pos: 0, speed: 0.9, length: 110 },
      { horizontal: true, lineIndex: 7, pos: 320, speed: 0.65, length: 90 },
      { horizontal: true, lineIndex: 12, pos: 640, speed: 0.75, length: 120 },
      { horizontal: false, lineIndex: 4, pos: 120, speed: 0.7, length: 95 },
      { horizontal: false, lineIndex: 11, pos: 420, speed: 0.85, length: 110 },
      { horizontal: false, lineIndex: 18, pos: 260, speed: 0.6, length: 85 },
    ];

    function syncSize() {
      if (!canvas || !ctx) {
        return;
      }
      width = window.innerWidth;
      height = window.innerHeight;
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      canvas.width = Math.max(1, Math.floor(width * dpr));
      canvas.height = Math.max(1, Math.floor(height * dpr));
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    }

    function drawFrame(timestamp: number) {
      if (!ctx) {
        return;
      }
      ctx.clearRect(0, 0, width, height);

      // Top ambient radial glow
      const glow = ctx.createRadialGradient(
        width * 0.5,
        -height * 0.1,
        20,
        width * 0.5,
        -height * 0.1,
        Math.max(width, height) * 0.72
      );
      glow.addColorStop(0, "rgba(46, 196, 182, 0.13)");
      glow.addColorStop(0.55, "rgba(46, 196, 182, 0.03)");
      glow.addColorStop(1, "rgba(8, 12, 20, 0)");
      ctx.fillStyle = glow;
      ctx.fillRect(0, 0, width, height);

      // Continuous 44x44 architectural grid
      ctx.strokeStyle = "rgba(148, 163, 184, 0.075)";
      ctx.lineWidth = 1;
      ctx.beginPath();

      for (let x = 0.5; x < width; x += GRID_SIZE) {
        ctx.moveTo(x, 0);
        ctx.lineTo(x, height);
      }
      for (let y = 0.5; y < height; y += GRID_SIZE) {
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
      }
      ctx.stroke();

      // Subtle breathing intersection nodes
      const cols = Math.ceil(width / GRID_SIZE);
      const rows = Math.ceil(height / GRID_SIZE);
      const timeSec = timestamp * 0.001;

      for (let col = 2; col < cols; col += 4) {
        for (let row = 2; row < rows; row += 4) {
          const x = col * GRID_SIZE + 0.5;
          const y = row * GRID_SIZE + 0.5;
          const wave = reducedMotion
            ? 0.22
            : 0.12 + 0.18 * (0.5 + 0.5 * Math.sin(timeSec * 1.4 + col * 0.7 + row * 0.5));
          ctx.fillStyle = `rgba(46, 196, 182, ${wave.toFixed(3)})`;
          ctx.fillRect(x - 1.5, y - 1.5, 3, 3);
        }
      }

      // Animated turquoise data pulses traveling along grid lines
      if (!reducedMotion) {
        for (const pulse of pulses) {
          if (pulse.horizontal) {
            const y = ((pulse.lineIndex * GRID_SIZE) % Math.max(GRID_SIZE, height)) + 0.5;
            pulse.pos += pulse.speed;
            if (pulse.pos - pulse.length > width) {
              pulse.pos = -40;
            }
            const grad = ctx.createLinearGradient(
              pulse.pos - pulse.length,
              y,
              pulse.pos,
              y
            );
            grad.addColorStop(0, "rgba(46, 196, 182, 0)");
            grad.addColorStop(1, "rgba(46, 196, 182, 0.38)");
            ctx.strokeStyle = grad;
            ctx.lineWidth = 1.5;
            ctx.beginPath();
            ctx.moveTo(pulse.pos - pulse.length, y);
            ctx.lineTo(pulse.pos, y);
            ctx.stroke();
          } else {
            const x = ((pulse.lineIndex * GRID_SIZE) % Math.max(GRID_SIZE, width)) + 0.5;
            pulse.pos += pulse.speed;
            if (pulse.pos - pulse.length > height) {
              pulse.pos = -40;
            }
            const grad = ctx.createLinearGradient(
              x,
              pulse.pos - pulse.length,
              x,
              pulse.pos
            );
            grad.addColorStop(0, "rgba(91, 192, 190, 0)");
            grad.addColorStop(1, "rgba(91, 192, 190, 0.34)");
            ctx.strokeStyle = grad;
            ctx.lineWidth = 1.5;
            ctx.beginPath();
            ctx.moveTo(x, pulse.pos - pulse.length);
            ctx.lineTo(x, pulse.pos);
            ctx.stroke();
          }
        }
      }
    }

    function loop(timestamp: number) {
      drawFrame(timestamp);
      if (!reducedMotion && !document.hidden) {
        rafId = window.requestAnimationFrame(loop);
      }
    }

    function startOrRenderStatic() {
      window.cancelAnimationFrame(rafId);
      if (reducedMotion || document.hidden) {
        drawFrame(0);
      } else {
        rafId = window.requestAnimationFrame(loop);
      }
    }

    function handleResize() {
      syncSize();
      drawFrame(performance.now());
    }

    function handleMotionChange(event: MediaQueryListEvent) {
      reducedMotion = event.matches;
      startOrRenderStatic();
    }

    function handleVisibilityChange() {
      startOrRenderStatic();
    }

    syncSize();
    startOrRenderStatic();

    window.addEventListener("resize", handleResize);
    document.addEventListener("visibilitychange", handleVisibilityChange);
    mediaQuery.addEventListener("change", handleMotionChange);

    return () => {
      window.cancelAnimationFrame(rafId);
      window.removeEventListener("resize", handleResize);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      mediaQuery.removeEventListener("change", handleMotionChange);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      id="grid-background-canvas"
      className="grid-canvas"
      aria-hidden="true"
    />
  );
}
