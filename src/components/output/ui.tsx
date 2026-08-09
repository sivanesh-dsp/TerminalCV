import { type ReactNode, useState } from 'react';
import { Check, Copy } from 'lucide-react';

/* ---------- coloured text spans ---------- */

export const Accent = ({ children }: { children: ReactNode }) => (
  <span className="text-term-accent">{children}</span>
);
export const Accent2 = ({ children }: { children: ReactNode }) => (
  <span className="text-term-accent2">{children}</span>
);
export const Muted = ({ children }: { children: ReactNode }) => (
  <span className="text-term-dim">{children}</span>
);
export const Ok = ({ children }: { children: ReactNode }) => (
  <span className="text-term-success">{children}</span>
);
export const Warn = ({ children }: { children: ReactNode }) => (
  <span className="text-term-warn">{children}</span>
);
export const Err = ({ children }: { children: ReactNode }) => (
  <span className="text-term-error">{children}</span>
);
export const Bold = ({ children }: { children: ReactNode }) => (
  <span className="font-bold text-term-fg">{children}</span>
);

/**
 * Renders inline `**highlight**` markers (from resume.json) as an accented,
 * semi-bold span. Everything else is passed through unchanged.
 */
export function Markup({ children }: { children: string }) {
  const parts = children.split(/\*\*(.+?)\*\*/g);
  return (
    <>
      {parts.map((part, i) =>
        i % 2 === 1 ? (
          <span key={i} className="font-semibold text-term-accent">
            {part}
          </span>
        ) : (
          <span key={i}>{part}</span>
        ),
      )}
    </>
  );
}

/* ---------- structural helpers ---------- */

/** A row of box-drawing dashes, length derived from the label it underlines. */
export const Dashes = ({ len }: { len: number }) => (
  <span className="text-term-dim" aria-hidden="true">
    {'─'.repeat(Math.max(3, len))}
  </span>
);

/** Section heading in the style of `SKILLS` with an accent underline. */
export function Heading({ title }: { title: string }) {
  return (
    <div className="mb-2 mt-1">
      <span className="font-bold uppercase tracking-widest text-term-accent">{title}</span>
      <div className="mt-1">
        <Dashes len={title.length + 4} />
      </div>
    </div>
  );
}

/** A labelled sub-group: label, dashed underline, then its body. */
export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="mb-3">
      <div className="text-term-accent2">{label}</div>
      <div className="leading-none">
        <Dashes len={label.length} />
      </div>
      <div className="mt-1">{children}</div>
    </div>
  );
}

/** Bullet list item with an accent marker. */
export function Bullet({ children }: { children: ReactNode }) {
  return (
    <div className="flex gap-2">
      <span className="select-none text-term-accent" aria-hidden="true">
        ▸
      </span>
      <span className="flex-1">{children}</span>
    </div>
  );
}

/** Aligned key/value row. `label` is padded to `width` chars in monospace. */
export function KV({
  label,
  children,
  width = 12,
}: {
  label: string;
  children: ReactNode;
  width?: number;
}) {
  return (
    <div className="flex flex-wrap gap-x-2">
      <span className="text-term-dim">{label.padEnd(width, ' ')}</span>
      <span className="text-term-fg">{children}</span>
    </div>
  );
}

/** Responsive two-column layout that collapses to one column on small screens. */
export const Columns = ({ children }: { children: ReactNode }) => (
  <div className="grid grid-cols-1 gap-x-10 gap-y-1 sm:grid-cols-2">{children}</div>
);

/** Inline tech "chip". */
export const Chip = ({ children }: { children: ReactNode }) => (
  <span className="inline-block rounded border border-term-dim/40 px-1.5 py-0.5 text-xs text-term-fg/90">
    {children}
  </span>
);

/** External link that is keyboard-focusable and screen-reader friendly. */
export function Ext({ href, children }: { href: string; children: ReactNode }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="text-term-link underline-offset-2 hover:underline"
    >
      {children}
    </a>
  );
}

/** Copy-to-clipboard button used by `copy` output affordances. */
export function CopyButton({ text, label = 'copy' }: { text: string; label?: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      type="button"
      onClick={async () => {
        try {
          await navigator.clipboard.writeText(text);
          setCopied(true);
          setTimeout(() => setCopied(false), 1500);
        } catch {
          setCopied(false);
        }
      }}
      className="no-print inline-flex items-center gap-1 rounded border border-term-dim/40 px-1.5 py-0.5 text-xs text-term-dim transition-colors hover:border-term-accent/60 hover:text-term-accent"
      aria-label={copied ? 'Copied to clipboard' : `Copy ${label}`}
    >
      {copied ? <Check size={12} /> : <Copy size={12} />}
      {copied ? 'copied' : label}
    </button>
  );
}

/** Standard error line for unknown/invalid commands. */
export function ErrorLine({ children }: { children: ReactNode }) {
  return <div className="text-term-error">{children}</div>;
}
