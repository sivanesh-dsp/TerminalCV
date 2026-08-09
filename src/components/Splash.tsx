import { useEffect } from 'react';
import { motion, useReducedMotion } from 'framer-motion';
import { resume } from '@/data/resume';

/**
 * Brief full-screen boot splash shown on first load. Auto-dismisses after a
 * short delay, or immediately on any key / tap. Honours reduced-motion.
 */
export function Splash({ onDone }: { onDone: () => void }) {
  const reduced = useReducedMotion();
  const holdMs = reduced ? 350 : 1700;

  useEffect(() => {
    // Prevent skip-keys from leaking into the (auto-focused) terminal input.
    (document.activeElement as HTMLElement | null)?.blur?.();
    const timer = setTimeout(onDone, holdMs);
    const skip = () => onDone();
    window.addEventListener('keydown', skip);
    window.addEventListener('pointerdown', skip);
    return () => {
      clearTimeout(timer);
      window.removeEventListener('keydown', skip);
      window.removeEventListener('pointerdown', skip);
    };
  }, [onDone, holdMs]);

  return (
    <motion.div
      className="fixed inset-0 z-[60] grid place-items-center bg-term-bg px-6 text-center"
      initial={false}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.4 }}
      role="dialog"
      aria-label={`${resume.name} — loading portfolio`}
    >
      <div>
        <motion.div
          initial={reduced ? false : { opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="font-mono text-3xl font-bold tracking-tight text-term-accent sm:text-5xl"
        >
          {resume.name.toLowerCase()}
          <span className="ml-1 inline-block h-[1em] w-[0.6ch] translate-y-[0.12em] animate-blink bg-term-accent align-baseline" />
        </motion.div>

        <motion.div
          initial={reduced ? false : { opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: reduced ? 0 : 0.35, duration: 0.4 }}
          className="mt-3 font-mono text-sm text-term-dim sm:text-base"
        >
          {resume.title}
        </motion.div>

        {/* Boot progress bar. */}
        <div className="mx-auto mt-6 h-[3px] w-48 overflow-hidden rounded-full bg-term-dim/25 sm:w-64">
          <motion.div
            className="h-full bg-term-accent"
            initial={{ width: reduced ? '100%' : '0%' }}
            animate={{ width: '100%' }}
            transition={{ duration: reduced ? 0 : holdMs / 1000, ease: 'easeInOut' }}
          />
        </div>

        <motion.div
          initial={reduced ? false : { opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: reduced ? 0 : 0.9, duration: 0.5 }}
          className="mt-4 font-mono text-xs text-term-dim"
        >
          press any key to continue
        </motion.div>
      </div>
    </motion.div>
  );
}
