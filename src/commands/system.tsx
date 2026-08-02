import { type ReactNode } from 'react';
import { resume } from '@/data/resume';
import { searchResume, splitHighlights } from '@/utils/search';
import { AsciiLogo } from '@/components/AsciiLogo';
import { THEMES, type ThemeName } from '@/hooks/useTheme';
import type { Command, CommandCategory, CommandContext } from '@/commands/types';
import { sectionViews } from '@/commands/info';
import { Accent, Bold, Dashes, Muted, Ok, Warn, ErrorLine } from '@/components/output/ui';

const CATEGORY_LABEL: Record<CommandCategory, string> = {
  info: 'INFO',
  sections: 'RÉSUMÉ',
  social: 'CONNECT',
  system: 'SYSTEM',
  fun: 'FUN',
};
const CATEGORY_ORDER: CommandCategory[] = ['info', 'sections', 'social', 'system', 'fun'];

/* -------------------------------- help -------------------------------- */

function HelpView({ ctx }: { ctx: CommandContext }) {
  const groups = CATEGORY_ORDER.map((cat) => ({
    cat,
    items: ctx.registry.all
      .filter((c) => c.category === cat && !c.hidden)
      .sort((a, b) => a.name.localeCompare(b.name)),
  })).filter((g) => g.items.length > 0);

  return (
    <div>
      <div className="text-term-fg">
        Available commands — <Accent>Tab</Accent> completes, <Accent>↑/↓</Accent> recall history.
      </div>
      <div className="mt-3 space-y-3">
        {groups.map((g) => (
          <div key={g.cat}>
            <div className="text-term-accent2">{CATEGORY_LABEL[g.cat]}</div>
            <div className="grid grid-cols-1 gap-x-8 md:grid-cols-2">
              {g.items.map((c) => (
                <div key={c.name} className="flex gap-2">
                  <span className="w-32 shrink-0 text-term-accent">{c.name}</span>
                  <span className="flex-1 text-term-dim">{c.description}</span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
      <div className="mt-3 text-term-dim">
        Hidden goodies exist. Try <Accent>sudo hire-me</Accent> or <Accent>matrix</Accent>. 🥚
      </div>
    </div>
  );
}

/* --------------------------- virtual filesystem --------------------------- */

const FILES: Record<string, ReactNode> = {
  'about.txt': sectionViews.about,
  'experience.md': sectionViews.experience,
  'projects.md': sectionViews.projects,
  'certifications.txt': sectionViews.certifications,
  'education.txt': sectionViews.education,
  'achievements.txt': sectionViews.achievements,
  'timeline.txt': sectionViews.timeline,
  'stats.txt': sectionViews.stats,
  'contact.vcf': sectionViews.contact,
};

export const LS_ENTRIES = [
  ...Object.keys(FILES),
  'skills.json',
  'resume.pdf',
];

function SkillsJson() {
  const obj = Object.fromEntries(resume.skills.map((c) => [c.name, c.skills]));
  return (
    <pre className="overflow-x-auto text-term-fg">
      {JSON.stringify(obj, null, 2)}
    </pre>
  );
}

/* -------------------------------- search -------------------------------- */

function SearchView({ query }: { query: string }) {
  const hits = searchResume(query);
  if (!query.trim()) {
    return (
      <div>
        <Warn>usage:</Warn> <Accent>search &lt;term&gt;</Accent>{' '}
        <Muted>— e.g. search kubernetes</Muted>
      </div>
    );
  }
  if (hits.length === 0) {
    return (
      <div className="text-term-warn">
        No matches for “{query}”. Try <Accent>search kubernetes</Accent> or{' '}
        <Accent>search terraform</Accent>.
      </div>
    );
  }
  return (
    <div>
      <div className="text-term-dim">
        {hits.length} match{hits.length > 1 ? 'es' : ''} for “
        <span className="text-term-fg">{query}</span>”
      </div>
      <div className="mt-2 space-y-2">
        {hits.map((h, i) => (
          <div key={i}>
            <div className="text-term-accent2">{h.section}</div>
            <div className="text-term-fg">
              {splitHighlights(h.text, query).map((seg, j) =>
                seg.match ? (
                  <mark key={j} className="bg-term-accent/30 text-term-fg">
                    {seg.text}
                  </mark>
                ) : (
                  <span key={j}>{seg.text}</span>
                ),
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

/* -------------------------------- theme -------------------------------- */

function themeCommand(ctx: CommandContext): ReactNode {
  const arg = (ctx.args[0] ?? '').toLowerCase();
  const actions = ctx.actions;
  if (!arg || arg === 'list') {
    return (
      <div>
        <div>
          Current theme: <Accent>{ctx.theme}</Accent>
        </div>
        <div className="mt-1 text-term-dim">
          usage: <Accent>theme</Accent> [{THEMES.join(' | ')} | hc | crt | next]
        </div>
      </div>
    );
  }
  if ((THEMES as readonly string[]).includes(arg)) {
    actions.setTheme(arg as ThemeName);
    return (
      <div>
        Theme set to <Accent>{arg}</Accent>. <Ok>✓</Ok>
      </div>
    );
  }
  if (arg === 'next') {
    actions.cycleTheme();
    return <div>Cycled to the next theme.</div>;
  }
  if (arg === 'hc' || arg === 'contrast') {
    actions.toggleHighContrast();
    return <div>Toggled high-contrast mode.</div>;
  }
  if (arg === 'crt') {
    actions.toggleCrt();
    return <div>Toggled CRT scanline effect.</div>;
  }
  return <ErrorLine>Unknown theme “{arg}”. Try: {THEMES.join(', ')}, hc, crt, next.</ErrorLine>;
}

/* ------------------------------- commands ------------------------------- */

export const systemCommands: Command[] = [
  {
    name: 'help',
    aliases: ['?', 'commands'],
    description: 'List everything you can do',
    category: 'system',
    run: (ctx) => <HelpView ctx={ctx} />,
  },
  {
    name: 'clear',
    aliases: ['cls'],
    description: 'Clear the screen',
    category: 'system',
    run: (ctx) => {
      ctx.actions.clear();
    },
  },
  {
    name: 'history',
    description: 'Show command history (history -c to clear)',
    category: 'system',
    run: (ctx) => {
      if ((ctx.args[0] ?? '') === '-c') {
        ctx.actions.clearHistory();
        return <Muted>History cleared.</Muted>;
      }
      if (ctx.history.length === 0) return <Muted>No history yet.</Muted>;
      return (
        <div className="font-mono">
          {ctx.history.map((h, i) => (
            <div key={i} className="flex gap-3">
              <span className="w-8 shrink-0 text-right text-term-dim">{i + 1}</span>
              <span className="text-term-fg">{h}</span>
            </div>
          ))}
        </div>
      );
    },
  },
  {
    name: 'search',
    aliases: ['grep', 'find'],
    description: 'Search the résumé, e.g. search kubernetes',
    usage: 'search <term>',
    category: 'system',
    run: (ctx) => <SearchView query={ctx.args.join(' ')} />,
  },
  {
    name: 'theme',
    description: 'Switch theme: dark | green | amber | hc | crt',
    usage: 'theme [name]',
    category: 'system',
    run: themeCommand,
  },
  {
    name: 'download',
    aliases: ['save'],
    description: 'Download the original PDF résumé',
    category: 'system',
    run: (ctx) => {
      ctx.actions.downloadResume();
      return (
        <div>
          <Ok>↓</Ok> Downloading <Accent>{resume.resumeFile}</Accent>…
        </div>
      );
    },
  },
  {
    name: 'print',
    description: 'Open the print dialog for the résumé',
    category: 'system',
    run: (ctx) => {
      ctx.actions.printResume();
      return <Muted>Opening print dialog…</Muted>;
    },
  },
  {
    name: 'ls',
    description: 'List résumé "files"',
    category: 'system',
    run: () => (
      <div className="grid grid-cols-2 gap-x-6 sm:grid-cols-3 md:grid-cols-4">
        {LS_ENTRIES.map((f) => (
          <span
            key={f}
            className={f.endsWith('.pdf') ? 'text-term-error' : 'text-term-accent'}
          >
            {f}
          </span>
        ))}
      </div>
    ),
  },
  {
    name: 'cat',
    description: 'Print a file, e.g. cat about.txt',
    usage: 'cat <file>',
    category: 'system',
    run: (ctx) => {
      const file = (ctx.args[0] ?? '').toLowerCase();
      if (!file) return <Warn>usage: cat &lt;file&gt; — try `ls` to see files.</Warn>;
      if (file === 'skills.json') return <SkillsJson />;
      if (file === 'resume.pdf')
        return (
          <div className="text-term-warn">
            resume.pdf: binary file. Run <Accent>download</Accent> to save it.
          </div>
        );
      const node = FILES[file];
      if (node) return node;
      return <ErrorLine>cat: {file}: No such file. Try `ls`.</ErrorLine>;
    },
  },
  {
    name: 'pwd',
    description: 'Print working directory',
    category: 'system',
    run: () => <span className="text-term-fg">/home/{resume.username}</span>,
  },
  {
    name: 'echo',
    description: 'Echo text back',
    category: 'system',
    run: (ctx) => <span className="text-term-fg">{ctx.args.join(' ')}</span>,
  },
  {
    name: 'date',
    description: 'Show the current date & time',
    category: 'system',
    run: () => <span className="text-term-fg">{new Date().toString()}</span>,
  },
  {
    name: 'banner',
    aliases: ['logo'],
    description: 'Reprint the ASCII banner',
    category: 'system',
    run: () => <AsciiLogo />,
  },
  {
    name: 'man',
    description: 'Show help for a command, e.g. man skills',
    usage: 'man <command>',
    category: 'system',
    run: (ctx) => {
      const name = (ctx.args[0] ?? '').toLowerCase();
      if (!name) return <Warn>What manual page do you want? e.g. `man stats`.</Warn>;
      const cmd = ctx.registry.get(name);
      if (!cmd) return <ErrorLine>No manual entry for “{name}”.</ErrorLine>;
      return (
        <div>
          <div>
            <Bold>{cmd.name.toUpperCase()}</Bold>
            <Dashes len={cmd.name.length + 4} />
          </div>
          <div className="mt-1">
            <span className="text-term-accent2">NAME</span>
            <div className="ml-4">
              {cmd.name}
              {cmd.aliases?.length ? ` (aliases: ${cmd.aliases.join(', ')})` : ''} — {cmd.description}
            </div>
          </div>
          <div className="mt-1">
            <span className="text-term-accent2">USAGE</span>
            <div className="ml-4">
              <Accent>{cmd.usage ?? cmd.name}</Accent>
            </div>
          </div>
        </div>
      );
    },
  },
  {
    name: 'exit',
    aliases: ['quit', 'logout'],
    description: 'End the session',
    category: 'system',
    run: () => (
      <div>
        <Muted>Connection to </Muted>
        <Accent>
          {resume.username}@{resume.host}
        </Accent>
        <Muted> closed.</Muted>
        <div className="mt-1 text-term-dim">
          (It’s a browser — nothing really closes. Type <Accent>help</Accent> to keep exploring.)
        </div>
      </div>
    ),
  },
];
