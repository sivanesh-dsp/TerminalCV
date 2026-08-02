import { useEffect, useRef, useState } from 'react';
import { useReducedMotion } from 'framer-motion';

interface TypewriterProps {
  text: string;
  /** Milliseconds per character. */
  speed?: number;
  /** Delay before typing starts (ms). */
  startDelay?: number;
  className?: string;
  /** Show a blinking block cursor while (and optionally after) typing. */
  cursor?: boolean;
  keepCursor?: boolean;
  onDone?: () => void;
}

/**
 * Types out `text` one character at a time (newlines included).
 * Honours `prefers-reduced-motion` by rendering the full text instantly.
 */
export function Typewriter({
  text,
  speed = 18,
  startDelay = 0,
  className,
  cursor = true,
  keepCursor = false,
  onDone,
}: TypewriterProps) {
  const reduced = useReducedMotion();
  const [count, setCount] = useState(reduced ? text.length : 0);
  const [done, setDone] = useState(reduced);
  const onDoneRef = useRef(onDone);
  onDoneRef.current = onDone;

  useEffect(() => {
    if (reduced) {
      onDoneRef.current?.();
      return;
    }
    let i = 0;
    let interval: ReturnType<typeof setInterval>;
    const startTimer = setTimeout(() => {
      interval = setInterval(() => {
        i += 1;
        setCount(i);
        if (i >= text.length) {
          clearInterval(interval);
          setDone(true);
          onDoneRef.current?.();
        }
      }, speed);
    }, startDelay);
    return () => {
      clearTimeout(startTimer);
      clearInterval(interval);
    };
  }, [text, speed, startDelay, reduced]);

  return (
    <span className={className} style={{ whiteSpace: 'pre-wrap' }}>
      {text.slice(0, count)}
      {cursor && (!done || keepCursor) && (
        <span className="ml-0.5 inline-block h-[1em] w-[0.6ch] translate-y-[0.15em] animate-blink bg-term-accent align-baseline" />
      )}
    </span>
  );
}
