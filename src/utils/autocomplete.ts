import { registry, ARG_OPTIONS } from '@/commands/registry';

export interface CompletionResult {
  /** The (possibly) extended input line after applying completion. */
  line: string;
  /** Candidate options to surface when the completion is ambiguous. */
  options: string[];
}

/** Longest common prefix shared by all strings. */
function commonPrefix(strings: string[]): string {
  if (strings.length === 0) return '';
  let prefix = strings[0];
  for (const s of strings) {
    while (!s.startsWith(prefix)) prefix = prefix.slice(0, -1);
  }
  return prefix;
}

/**
 * bash-style Tab completion.
 * - First token → complete against command names.
 * - Known second tokens (theme/cat/man/sudo) → complete against their options.
 * Extends to the longest unambiguous prefix and returns candidates otherwise.
 */
export function complete(line: string): CompletionResult {
  const trailingSpace = /\s$/.test(line);
  const tokens = line.split(/\s+/).filter(Boolean);

  // ---- command name completion ----
  if (tokens.length <= 1 && !trailingSpace) {
    const partial = (tokens[0] ?? '').toLowerCase();
    const names = registry.all.map((c) => c.name);
    const matches = names.filter((n) => n.startsWith(partial)).sort();
    if (matches.length === 0) return { line, options: [] };
    if (matches.length === 1) return { line: matches[0] + ' ', options: [] };
    const cp = commonPrefix(matches);
    return { line: cp.length > partial.length ? cp : line, options: matches };
  }

  // ---- argument completion ----
  const cmd = tokens[0].toLowerCase();
  const opts = ARG_OPTIONS[cmd];
  if (!opts) return { line, options: [] };

  const partial = trailingSpace ? '' : (tokens[tokens.length - 1] ?? '').toLowerCase();
  const matches = opts.filter((o) => o.toLowerCase().startsWith(partial)).sort();
  if (matches.length === 0) return { line, options: [] };

  const baseTokens = trailingSpace ? tokens : tokens.slice(0, -1);
  const base = baseTokens.join(' ');
  if (matches.length === 1) return { line: `${base} ${matches[0]} `, options: [] };

  const cp = commonPrefix(matches);
  const line2 = cp.length > partial.length ? `${base} ${cp}` : line;
  return { line: line2, options: matches };
}

/** Live command suggestions shown as the user types the first token. */
export function suggest(line: string, limit = 6): string[] {
  const trailingSpace = /\s$/.test(line);
  const tokens = line.split(/\s+/).filter(Boolean);
  if (tokens.length !== 1 || trailingSpace) return [];
  const partial = tokens[0].toLowerCase();
  if (!partial) return [];
  return registry.all
    .map((c) => c.name)
    .filter((n) => n.startsWith(partial) && n !== partial)
    .sort()
    .slice(0, limit);
}
