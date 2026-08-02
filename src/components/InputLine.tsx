import { type KeyboardEvent, type RefObject } from 'react';
import { Prompt } from '@/components/Prompt';

interface InputLineProps {
  value: string;
  caret: number;
  focused: boolean;
  inputRef: RefObject<HTMLInputElement>;
  suggestions: string[];
  onChange: (value: string, caret: number) => void;
  onKeyDown: (e: KeyboardEvent<HTMLInputElement>) => void;
  onCaretSync: () => void;
  onSuggestionClick: (s: string) => void;
}

/**
 * Active input line: a transparent native <input> (for real keyboard, IME,
 * mobile and screen-reader support) overlaid by a rendered mirror that draws a
 * classic blinking block cursor at the caret position.
 */
export function InputLine({
  value,
  caret,
  focused,
  inputRef,
  suggestions,
  onChange,
  onKeyDown,
  onCaretSync,
  onSuggestionClick,
}: InputLineProps) {
  const before = value.slice(0, caret);
  const under = value.charAt(caret) || ' ';
  const after = value.slice(caret + 1);

  return (
    <div>
      <div className="flex flex-wrap items-start">
        <Prompt />
        <div className="relative min-w-[2ch] flex-1">
          <input
            ref={inputRef}
            value={value}
            onChange={(e) => onChange(e.target.value, e.target.selectionStart ?? e.target.value.length)}
            onKeyDown={onKeyDown}
            onKeyUp={onCaretSync}
            onClick={onCaretSync}
            onSelect={onCaretSync}
            className="absolute inset-0 h-full w-full border-0 bg-transparent p-0 font-mono text-base text-transparent caret-transparent outline-none"
            style={{ caretColor: 'transparent' }}
            type="text"
            autoComplete="off"
            autoCorrect="off"
            autoCapitalize="off"
            spellCheck={false}
            enterKeyHint="go"
            aria-label="Terminal command input. Type a command such as help, then press Enter."
          />
          <div aria-hidden="true" className="whitespace-pre-wrap break-all text-term-fg">
            {before}
            <span
              className={
                focused
                  ? 'animate-blink bg-term-accent text-term-bg'
                  : 'bg-term-dim/50 text-term-bg'
              }
            >
              {under}
            </span>
            {after}
          </div>
        </div>
      </div>

      {focused && suggestions.length > 0 && (
        <div className="mt-1 flex flex-wrap gap-2 pl-0 text-xs text-term-dim no-print">
          <span className="text-term-dim">↳</span>
          {suggestions.map((s) => (
            <button
              key={s}
              type="button"
              // Prevent the input from losing focus before we handle the click.
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => onSuggestionClick(s)}
              className="rounded px-1 text-term-accent/80 transition-colors hover:bg-term-accent/10 hover:text-term-accent"
            >
              {s}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
