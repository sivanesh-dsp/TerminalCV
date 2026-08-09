import {
  forwardRef,
  type KeyboardEvent,
  type ReactNode,
  useCallback,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from 'react';
import { useReducedMotion } from 'framer-motion';
import { resume } from '@/data/resume';
import { registry } from '@/commands/registry';
import { complete, suggest } from '@/utils/autocomplete';
import { useCommandHistory } from '@/hooks/useCommandHistory';
import type { ThemeName } from '@/hooks/useTheme';
import type { TerminalActions } from '@/commands/types';
import { Welcome } from '@/components/Welcome';
import { InputLine } from '@/components/InputLine';
import { HistoryBlock, type Block } from '@/components/HistoryBlock';
import { Accent, ErrorLine } from '@/components/output/ui';

export interface TerminalHandle {
  runCommand: (cmd: string) => void;
  focusInput: () => void;
}

interface TerminalProps {
  theme: ThemeName;
  setTheme: (t: ThemeName) => void;
  cycleTheme: () => void;
  toggleHighContrast: () => void;
  toggleCrt: () => void;
  onStartMatrix: () => void;
  onDownload: () => void;
  onPrint: () => void;
}

export const Terminal = forwardRef<TerminalHandle, TerminalProps>(function Terminal(props, ref) {
  const [blocks, setBlocks] = useState<Block[]>([]);
  const [showWelcome, setShowWelcome] = useState(true);
  const [input, setInput] = useState('');
  const [caret, setCaret] = useState(0);
  const [focused, setFocused] = useState(true);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [revealId, setRevealId] = useState<string | null>(null);

  const reduced = useReducedMotion();
  const busy = revealId !== null;
  const busyRef = useRef(false);
  busyRef.current = busy;

  const inputRef = useRef<HTMLInputElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const stickToBottom = useRef(true);
  const blockId = useRef(0);
  const suppressAppend = useRef(false);

  const { history, push, prev, next, reset, clear: clearHistory } = useCommandHistory();

  /* ---------- scrolling: stay pinned to the bottom as content grows ---------- */
  const scrollToBottom = useCallback((smooth = false) => {
    const el = scrollRef.current;
    if (!el) return;
    el.scrollTo({ top: el.scrollHeight, behavior: smooth ? 'smooth' : 'auto' });
  }, []);

  useEffect(() => {
    const content = contentRef.current;
    if (!content) return;
    const ro = new ResizeObserver(() => {
      if (stickToBottom.current) scrollToBottom(false);
    });
    ro.observe(content);
    return () => ro.disconnect();
  }, [scrollToBottom]);

  const onScroll = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    stickToBottom.current = el.scrollHeight - el.scrollTop - el.clientHeight < 48;
  }, []);

  /* ------------------------- caret + input line helpers ------------------------- */
  const setLine = useCallback((value: string) => {
    setInput(value);
    setCaret(value.length);
    requestAnimationFrame(() => {
      const el = inputRef.current;
      if (el) {
        el.focus();
        el.setSelectionRange(value.length, value.length);
      }
    });
  }, []);

  const syncCaret = useCallback(() => {
    const el = inputRef.current;
    if (el) setCaret(el.selectionStart ?? el.value.length);
  }, []);

  const focusInput = useCallback(() => inputRef.current?.focus(), []);

  /* -------------------------------- execution -------------------------------- */
  const appendBlock = useCallback(
    (line: string | null, output: ReactNode, reveal = false) => {
      const id = `b${blockId.current++}`;
      setBlocks((prev) => [...prev, { id, input: line, output }]);
      // Hold the input until the output has finished revealing.
      if (reveal && output != null && !reduced) setRevealId(id);
    },
    [reduced],
  );

  const onRevealed = useCallback(() => {
    setRevealId(null);
    requestAnimationFrame(() => scrollToBottom(true));
  }, [scrollToBottom]);

  // Refocus the input whenever it returns after an output reveal.
  useEffect(() => {
    if (!busy) inputRef.current?.focus();
  }, [busy]);

  const execute = useCallback(
    (rawInput: string) => {
      if (busyRef.current) return; // ignore while output is still revealing
      const line = rawInput;
      const trimmed = line.trim();

      const actions: TerminalActions = {
        clear: () => {
          suppressAppend.current = true;
          setBlocks([]);
          setShowWelcome(false);
        },
        setTheme: props.setTheme,
        cycleTheme: props.cycleTheme,
        toggleHighContrast: props.toggleHighContrast,
        toggleCrt: props.toggleCrt,
        startMatrix: props.onStartMatrix,
        downloadResume: props.onDownload,
        printResume: props.onPrint,
        runCommand: (cmd) => execute(cmd),
        focusInput,
        clearHistory,
      };

      let output: ReactNode = null;
      if (trimmed) {
        const [name, ...args] = trimmed.split(/\s+/);
        const cmd = registry.get(name);
        if (cmd) {
          output =
            cmd.run({
              args,
              rawInput: trimmed,
              resume,
              registry,
              history: [...history, trimmed],
              theme: props.theme,
              actions,
            }) ?? null;
        } else {
          output = (
            <ErrorLine>
              command not found: {name} — type <Accent>help</Accent> for the list.
            </ErrorLine>
          );
        }
        push(trimmed);
      }

      if (suppressAppend.current) {
        suppressAppend.current = false;
      } else {
        appendBlock(line || '', output, true);
      }

      setInput('');
      setCaret(0);
      setSuggestions([]);
      reset();
      stickToBottom.current = true;
      requestAnimationFrame(() => scrollToBottom(true));
    },
    [appendBlock, clearHistory, focusInput, history, props, push, reset, scrollToBottom],
  );

  const executeRef = useRef(execute);
  executeRef.current = execute;

  useImperativeHandle(
    ref,
    () => ({
      runCommand: (cmd: string) => executeRef.current(cmd),
      focusInput: () => inputRef.current?.focus(),
    }),
    [],
  );

  /* -------------------------------- key handling -------------------------------- */
  const onKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement>) => {
      const ctrl = e.ctrlKey || e.metaKey;

      if (e.key === 'Enter') {
        e.preventDefault();
        execute(input);
        return;
      }
      if (e.key === 'Tab') {
        e.preventDefault();
        const res = complete(input);
        if (res.line !== input) setLine(res.line);
        setSuggestions(res.options);
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        const v = prev(input);
        if (v !== null) setLine(v);
        return;
      }
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        const v = next();
        setLine(v ?? '');
        return;
      }
      if (e.key === 'Escape') {
        setSuggestions([]);
        return;
      }
      if (ctrl && (e.key === 'l' || e.key === 'L')) {
        e.preventDefault();
        setBlocks([]);
        setShowWelcome(false);
        return;
      }
      if (ctrl && (e.key === 'c' || e.key === 'C')) {
        // Only intercept when there is no active text selection to copy.
        if (!window.getSelection()?.toString()) {
          e.preventDefault();
          appendBlock(input + ' ^C', null);
          setInput('');
          setCaret(0);
          setSuggestions([]);
          reset();
        }
        return;
      }
      if (ctrl && (e.key === 'u' || e.key === 'U')) {
        e.preventDefault();
        setLine('');
        setSuggestions([]);
        return;
      }
      if (ctrl && (e.key === 'a' || e.key === 'A')) {
        e.preventDefault();
        inputRef.current?.setSelectionRange(0, 0);
        setCaret(0);
        return;
      }
      if (ctrl && (e.key === 'e' || e.key === 'E')) {
        e.preventDefault();
        inputRef.current?.setSelectionRange(input.length, input.length);
        setCaret(input.length);
        return;
      }
    },
    [appendBlock, execute, input, next, prev, reset, setLine],
  );

  const onChange = useCallback((value: string, nextCaret: number) => {
    setInput(value);
    setCaret(nextCaret);
    setSuggestions(suggest(value));
  }, []);

  /* ------------------------- focus behaviour on click ------------------------- */
  const onContainerMouseUp = useCallback(() => {
    if (window.getSelection()?.toString()) return; // preserve user text selection
    focusInput();
  }, [focusInput]);

  useEffect(() => {
    focusInput();
  }, [focusInput]);

  return (
    <div
      ref={scrollRef}
      onScroll={onScroll}
      onMouseUp={onContainerMouseUp}
      className="h-full overflow-y-auto px-4 py-4 text-sm leading-relaxed sm:text-base"
    >
      <h1 className="sr-only">
        {resume.name} — {resume.title}. Interactive terminal résumé.
      </h1>

      <div ref={contentRef} className="mx-auto max-w-4xl space-y-3 pb-24">
        <div role="log" aria-live="polite" aria-label="Terminal output" className="space-y-3">
          {showWelcome && <Welcome onRun={(c) => execute(c)} />}
          {blocks.map((b) => (
            <HistoryBlock
              key={b.id}
              block={b}
              reveal={b.id === revealId}
              onRevealed={onRevealed}
            />
          ))}
        </div>

        {!busy && (
          <InputLine
            value={input}
            caret={caret}
            focused={focused}
            inputRef={inputRef}
            suggestions={suggestions}
            onChange={onChange}
            onKeyDown={onKeyDown}
            onCaretSync={syncCaret}
            onSuggestionClick={(s) => {
              setSuggestions([]);
              execute(s);
            }}
          />
        )}
      </div>

      {/* Focus tracking (kept off the input element to avoid re-render churn). */}
      <FocusProbe inputRef={inputRef} setFocused={setFocused} rebind={busy} />
    </div>
  );
});

/** Tracks focus/blur of the input to drive the block-cursor style. */
function FocusProbe({
  inputRef,
  setFocused,
  rebind,
}: {
  inputRef: React.RefObject<HTMLInputElement>;
  setFocused: (v: boolean) => void;
  rebind: boolean;
}) {
  useEffect(() => {
    const el = inputRef.current;
    if (!el) return;
    const onFocus = () => setFocused(true);
    const onBlur = () => setFocused(false);
    el.addEventListener('focus', onFocus);
    el.addEventListener('blur', onBlur);
    setFocused(document.activeElement === el);
    return () => {
      el.removeEventListener('focus', onFocus);
      el.removeEventListener('blur', onBlur);
    };
    // `rebind` toggles when the input unmounts/remounts (busy cycle), so we
    // re-attach the listeners to the fresh element.
  }, [inputRef, setFocused, rebind]);
  return null;
}
