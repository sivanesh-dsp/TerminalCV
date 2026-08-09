import { useEffect, useRef } from 'react';
import { useReducedMotion } from 'framer-motion';

const GLYPHS = 'アカサタナハマヤラワ0123456789ABCDEFｸﾂﾅﾋﾐ<>*/{}[]$#@'.split('');

/** Full-screen "digital rain" overlay. Exits on ESC, click or tap. */
export function MatrixRain({ onExit }: { onExit: () => void }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const exitArmed = useRef<(() => boolean) | null>(null);
  const reduced = useReducedMotion();

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let width = (canvas.width = window.innerWidth);
    let height = (canvas.height = window.innerHeight);
    const fontSize = 16;
    let columns = Math.floor(width / fontSize);
    let drops = new Array(columns).fill(1).map(() => Math.random() * -50);

    const resize = () => {
      width = canvas.width = window.innerWidth;
      height = canvas.height = window.innerHeight;
      columns = Math.floor(width / fontSize);
      drops = new Array(columns).fill(1).map(() => Math.random() * -50);
    };

    const draw = () => {
      ctx.fillStyle = 'rgba(0, 0, 0, 0.06)';
      ctx.fillRect(0, 0, width, height);
      ctx.fillStyle = '#5ef2b0';
      ctx.font = `${fontSize}px monospace`;
      for (let i = 0; i < drops.length; i++) {
        const text = GLYPHS[Math.floor(Math.random() * GLYPHS.length)];
        ctx.fillText(text, i * fontSize, drops[i] * fontSize);
        if (drops[i] * fontSize > height && Math.random() > 0.975) drops[i] = 0;
        drops[i]++;
      }
    };

    let timer: ReturnType<typeof setInterval> | undefined;
    if (reduced) {
      // Reduced motion: paint a single static frame instead of animating.
      ctx.fillStyle = '#000';
      ctx.fillRect(0, 0, width, height);
      for (let i = 0; i < 40; i++) draw();
    } else {
      timer = setInterval(draw, 45);
    }

    window.addEventListener('resize', resize);
    return () => {
      if (timer) clearInterval(timer);
      window.removeEventListener('resize', resize);
    };
  }, [reduced]);

  useEffect(() => {
    // Arm exit handlers on the next tick. Without this, the very Enter keypress
    // that launches matrix bubbles to window *after* React has synchronously
    // mounted this overlay, so the fresh keydown listener would catch it and
    // exit immediately.
    let armed = false;
    const armTimer = setTimeout(() => {
      armed = true;
    }, 100);
    const onKey = (e: KeyboardEvent) => {
      if (!armed) return;
      if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') onExit();
    };
    window.addEventListener('keydown', onKey);
    exitArmed.current = () => armed;
    return () => {
      clearTimeout(armTimer);
      window.removeEventListener('keydown', onKey);
    };
  }, [onExit]);

  return (
    <div
      className="fixed inset-0 z-50 bg-black"
      onClick={() => {
        if (exitArmed.current?.()) onExit();
      }}
      role="dialog"
      aria-label="Matrix animation. Press Escape or tap to exit."
    >
      <canvas ref={canvasRef} className="block h-full w-full" />
      <div className="pointer-events-none absolute bottom-4 left-1/2 -translate-x-1/2 rounded bg-black/60 px-3 py-1 text-xs text-term-accent">
        press ESC or tap to exit
      </div>
    </div>
  );
}
