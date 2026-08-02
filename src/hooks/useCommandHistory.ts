import { useCallback, useRef, useState } from 'react';

const HISTORY_KEY = 'tr:history';
const MAX_HISTORY = 200;

function load(): string[] {
  try {
    const raw = localStorage.getItem(HISTORY_KEY);
    const parsed = raw ? (JSON.parse(raw) as unknown) : [];
    return Array.isArray(parsed) ? (parsed as string[]).slice(-MAX_HISTORY) : [];
  } catch {
    return [];
  }
}

function persist(entries: string[]) {
  try {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(entries));
  } catch {
    /* storage may be unavailable (private mode) — history stays in-memory */
  }
}

/**
 * Persistent visitor command history with bash-style ↑/↓ recall.
 *
 * `cursor` points at the entry the user is currently viewing;
 * `cursor === history.length` means "editing a fresh line", in which case
 * `draft` holds whatever they had half-typed before pressing ↑.
 */
export function useCommandHistory() {
  const [history, setHistory] = useState<string[]>(load);
  const historyRef = useRef<string[]>(history);
  const cursor = useRef<number>(history.length);
  const draft = useRef<string>('');

  historyRef.current = history;

  const push = useCallback((entry: string) => {
    const trimmed = entry.trim();
    setHistory((prev) => {
      // Skip consecutive duplicates, mirroring `HISTCONTROL=ignoredups`.
      const next =
        prev[prev.length - 1] === trimmed ? prev : [...prev, trimmed].slice(-MAX_HISTORY);
      cursor.current = next.length;
      draft.current = '';
      persist(next);
      return next;
    });
  }, []);

  /** Recall the previous command (↑). `current` is preserved as a draft the first time. */
  const prev = useCallback((current: string): string | null => {
    const list = historyRef.current;
    if (list.length === 0) return null;
    if (cursor.current === list.length) draft.current = current;
    cursor.current = Math.max(0, cursor.current - 1);
    return list[cursor.current] ?? null;
  }, []);

  /** Recall the next command (↓). Returns the saved draft once past the newest entry. */
  const next = useCallback((): string | null => {
    const list = historyRef.current;
    if (cursor.current >= list.length) {
      cursor.current = list.length;
      return draft.current;
    }
    cursor.current += 1;
    if (cursor.current >= list.length) {
      cursor.current = list.length;
      return draft.current;
    }
    return list[cursor.current] ?? null;
  }, []);

  /** Return to the bottom of the stack (called after a command is executed). */
  const reset = useCallback(() => {
    cursor.current = historyRef.current.length;
    draft.current = '';
  }, []);

  const clear = useCallback(() => {
    setHistory([]);
    cursor.current = 0;
    draft.current = '';
    try {
      localStorage.removeItem(HISTORY_KEY);
    } catch {
      /* ignore */
    }
  }, []);

  return { history, push, prev, next, reset, clear };
}
