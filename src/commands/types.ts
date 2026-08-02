import type { ReactNode } from 'react';
import type { ResumeData } from '@/types/resume';
import type { ThemeName } from '@/hooks/useTheme';

/** Side-effecting hooks the terminal exposes to commands. */
export interface TerminalActions {
  clear: () => void;
  setTheme: (t: ThemeName) => void;
  cycleTheme: () => void;
  toggleHighContrast: () => void;
  toggleCrt: () => void;
  /** Launch the full-screen matrix overlay (exit with ESC / tap). */
  startMatrix: () => void;
  downloadResume: () => void;
  printResume: () => void;
  /** Programmatically run another command line (used by aliases/among others). */
  runCommand: (input: string) => void;
  focusInput: () => void;
  clearHistory: () => void;
}

export interface CommandContext {
  /** Positional arguments after the command name. */
  args: string[];
  /** The full raw line the user typed. */
  rawInput: string;
  resume: ResumeData;
  registry: CommandRegistry;
  /** Past command lines (most-recent last). */
  history: string[];
  theme: ThemeName;
  actions: TerminalActions;
}

export type CommandCategory = 'info' | 'sections' | 'social' | 'system' | 'fun';

export interface Command {
  name: string;
  aliases?: string[];
  description: string;
  usage?: string;
  category: CommandCategory;
  /** Hidden from `help` (easter eggs) but still resolvable + autocompletable. */
  hidden?: boolean;
  run: (ctx: CommandContext) => ReactNode | void;
}

export interface CommandRegistry {
  all: Command[];
  /** Lookup map including aliases, all lower-cased. */
  byName: Map<string, Command>;
  get(name: string): Command | undefined;
  /** Primary command names for autocomplete/help. */
  names(includeHidden?: boolean): string[];
}
