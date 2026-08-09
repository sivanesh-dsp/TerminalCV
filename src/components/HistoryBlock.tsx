import { memo, type ReactNode, useRef, useState } from 'react';
import { motion, useReducedMotion } from 'framer-motion';
import { Check, Copy } from 'lucide-react';
import { Prompt } from '@/components/Prompt';

export interface Block {
  id: string;
  /** The command line the user typed, or null for the welcome block. */
  input: string | null;
  output: ReactNode;
}

/** Copies the rendered plain-text of a block's output via the DOM. */
function CopyOutput({ targetRef }: { targetRef: React.RefObject<HTMLDivElement> }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      type="button"
      aria-label={copied ? 'Output copied' : 'Copy output'}
      title="Copy output"
      onClick={async () => {
        const text = targetRef.current?.innerText ?? '';
        try {
          await navigator.clipboard.writeText(text.trim());
          setCopied(true);
          setTimeout(() => setCopied(false), 1500);
        } catch {
          /* clipboard unavailable */
        }
      }}
      className="no-print absolute right-0 top-0 hidden items-center gap-1 rounded border border-term-dim/40 bg-term-bg/80 px-1.5 py-0.5 text-xs text-term-dim opacity-0 transition-opacity hover:border-term-accent/60 hover:text-term-accent focus:opacity-100 group-hover:flex group-hover:opacity-100"
    >
      {copied ? <Check size={12} /> : <Copy size={12} />}
      {copied ? 'copied' : 'copy'}
    </button>
  );
}

function HistoryBlockImpl({
  block,
  reveal = false,
  onRevealed,
}: {
  block: Block;
  reveal?: boolean;
  onRevealed?: () => void;
}) {
  const reduced = useReducedMotion();
  const outRef = useRef<HTMLDivElement>(null);
  const hasOutput = block.output != null;
  const doReveal = reveal && hasOutput && !reduced;
  return (
    <motion.div
      initial={reduced ? false : { opacity: 0, y: 2 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.15 }}
      className="group relative"
    >
      {block.input !== null && (
        <div className="flex flex-wrap">
          <Prompt />
          <span className="break-all text-term-fg">{block.input}</span>
        </div>
      )}
      {hasOutput &&
        (doReveal ? (
          <motion.div
            ref={outRef}
            className="mt-0.5"
            initial={{ clipPath: 'inset(0 0 100% 0)', opacity: 0.4 }}
            animate={{ clipPath: 'inset(0 0 0% 0)', opacity: 1 }}
            transition={{ duration: 0.5, ease: 'easeOut' }}
            onAnimationComplete={() => onRevealed?.()}
          >
            {block.output}
          </motion.div>
        ) : (
          <div ref={outRef} className="mt-0.5">
            {block.output}
          </div>
        ))}
      {hasOutput && block.input !== null && <CopyOutput targetRef={outRef} />}
    </motion.div>
  );
}

export const HistoryBlock = memo(HistoryBlockImpl);
