import { useEffect, useMemo, useRef, useState } from 'react';
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion';
import { Search } from 'lucide-react';
import { registry } from '@/commands/registry';

interface CommandPaletteProps {
  open: boolean;
  onClose: () => void;
  onRun: (cmd: string) => void;
}

/** Ctrl/⌘+K command palette: fuzzy-ish filter, arrow-key nav, Enter to run. */
export function CommandPalette({ open, onClose, onRun }: CommandPaletteProps) {
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const reduced = useReducedMotion();

  const results = useMemo(() => {
    const q = query.trim().toLowerCase();
    const list = registry.all.filter((c) => !c.hidden);
    if (!q) return list;
    return list.filter(
      (c) =>
        c.name.includes(q) ||
        c.description.toLowerCase().includes(q) ||
        c.aliases?.some((a) => a.includes(q)),
    );
  }, [query]);

  useEffect(() => {
    if (open) {
      setQuery('');
      setSelected(0);
      // Focus after mount/paint.
      const id = requestAnimationFrame(() => inputRef.current?.focus());
      return () => cancelAnimationFrame(id);
    }
  }, [open]);

  useEffect(() => setSelected(0), [query]);

  if (!open) return null;

  const run = (name: string) => {
    onRun(name);
    onClose();
  };

  return (
    <AnimatePresence>
      <div
        className="fixed inset-0 z-40 flex items-start justify-center bg-black/60 p-4 pt-[12vh]"
        onClick={onClose}
        role="dialog"
        aria-modal="true"
        aria-label="Command palette"
      >
        <motion.div
          initial={reduced ? false : { opacity: 0, y: -8, scale: 0.98 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.12 }}
          onClick={(e) => e.stopPropagation()}
          className="w-full max-w-lg overflow-hidden rounded-lg border border-term-dim/40 bg-term-bg shadow-2xl"
        >
          <div className="flex items-center gap-2 border-b border-term-dim/30 px-3 py-2">
            <Search size={16} className="text-term-dim" />
            <input
              ref={inputRef}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Escape') onClose();
                else if (e.key === 'ArrowDown') {
                  e.preventDefault();
                  setSelected((s) => Math.min(results.length - 1, s + 1));
                } else if (e.key === 'ArrowUp') {
                  e.preventDefault();
                  setSelected((s) => Math.max(0, s - 1));
                } else if (e.key === 'Enter' && results[selected]) {
                  e.preventDefault();
                  run(results[selected].name);
                }
              }}
              placeholder="Type a command…"
              className="flex-1 bg-transparent text-term-fg outline-none placeholder:text-term-dim"
              aria-label="Search commands"
            />
            <kbd className="rounded border border-term-dim/40 px-1 text-xs text-term-dim">esc</kbd>
          </div>
          <ul className="max-h-72 overflow-y-auto py-1" role="listbox">
            {results.length === 0 && (
              <li className="px-3 py-2 text-sm text-term-dim">No matching commands.</li>
            )}
            {results.map((c, i) => (
              <li key={c.name} role="option" aria-selected={i === selected}>
                <button
                  type="button"
                  onMouseEnter={() => setSelected(i)}
                  onClick={() => run(c.name)}
                  className={`flex w-full items-center justify-between gap-3 px-3 py-1.5 text-left ${
                    i === selected ? 'bg-term-accent/15' : ''
                  }`}
                >
                  <span className="text-term-accent">{c.name}</span>
                  <span className="flex-1 truncate text-xs text-term-dim">{c.description}</span>
                </button>
              </li>
            ))}
          </ul>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}
