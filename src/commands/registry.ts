import type { Command, CommandRegistry } from '@/commands/types';
import { infoCommands, resumeCommand } from '@/commands/info';
import { systemCommands, LS_ENTRIES } from '@/commands/system';
import { funCommands } from '@/commands/fun';

/** Every command the terminal knows about. Order affects nothing except `all`. */
const ALL: Command[] = [
  resumeCommand,
  ...infoCommands,
  ...systemCommands,
  ...funCommands,
];

function build(): CommandRegistry {
  const byName = new Map<string, Command>();
  for (const c of ALL) {
    byName.set(c.name.toLowerCase(), c);
    c.aliases?.forEach((a) => byName.set(a.toLowerCase(), c));
  }
  return {
    all: ALL,
    byName,
    get: (name: string) => byName.get(name.trim().toLowerCase()),
    names: (includeHidden = true) =>
      ALL.filter((c) => includeHidden || !c.hidden)
        .map((c) => c.name)
        .sort((a, b) => a.localeCompare(b)),
  };
}

export const registry = build();

/** Second-token completions for commands that take a known set of arguments. */
export const ARG_OPTIONS: Record<string, string[]> = {
  theme: ['dark', 'green', 'amber', 'hc', 'crt', 'next'],
  cat: LS_ENTRIES,
  man: registry.all.map((c) => c.name).sort((a, b) => a.localeCompare(b)),
  sudo: ['hire-me'],
};
