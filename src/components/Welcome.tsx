import { motion, useReducedMotion } from 'framer-motion';
import { AsciiLogo } from '@/components/AsciiLogo';
import { Typewriter } from '@/components/Typewriter';
import { resume } from '@/data/resume';

const SUGGESTIONS = ['about', 'experience', 'skills', 'projects', 'contact', 'resume'];

/** Small clickable chip that runs a command when activated. */
function RunChip({ cmd, onRun }: { cmd: string; onRun?: (c: string) => void }) {
  return (
    <button
      type="button"
      onClick={() => onRun?.(cmd)}
      className="rounded border border-term-dim/40 px-2 py-0.5 text-term-fg/90 transition-colors hover:border-term-accent hover:text-term-accent"
    >
      {cmd}
    </button>
  );
}

export function Welcome({ onRun }: { onRun?: (c: string) => void }) {
  const reduced = useReducedMotion();
  return (
    <div className="pb-2">
      <AsciiLogo />

      <div className="mt-2 text-term-fg">
        <span className="font-bold">{resume.name}</span>{' '}
        <span className="text-term-dim">— {resume.title}</span>
      </div>
      <div className="text-term-dim">
        Certified Kubernetes Administrator (CKA) · {resume.contact.location}
      </div>

      <div className="mt-3">
        <Typewriter
          text="Welcome to my interactive terminal résumé — explore it by typing commands."
          speed={16}
          className="text-term-accent"
        />
      </div>

      <motion.div
        initial={reduced ? false : { opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: reduced ? 0 : 1.4, duration: 0.4 }}
      >
        <div className="mt-3 text-term-fg">
          Type <span className="text-term-accent">help</span> to see everything, or try:
        </div>
        <div className="mt-2 flex flex-wrap gap-2 text-sm no-print">
          {SUGGESTIONS.map((c) => (
            <RunChip key={c} cmd={c} onRun={onRun} />
          ))}
        </div>
        <div className="mt-3 text-xs text-term-dim">
          <span className="text-term-accent2">↑/↓</span> history ·{' '}
          <span className="text-term-accent2">Tab</span> autocomplete ·{' '}
          <span className="text-term-accent2">Ctrl+L</span> clear ·{' '}
          <span className="text-term-accent2">Ctrl+K</span> palette
        </div>
      </motion.div>
    </div>
  );
}
