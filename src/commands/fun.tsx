import { useEffect, useRef, useState } from 'react';
import { useReducedMotion } from 'framer-motion';
import { resume, allTechnologies } from '@/data/resume';
import { experienceLabel } from '@/utils/format';
import type { Command } from '@/commands/types';
import { Accent, Accent2, Bold, Muted, Ok, Warn } from '@/components/output/ui';

/* --------------------------- shared animation hooks --------------------------- */

/** Reveal `total` steps one-by-one; returns how many are currently visible. */
function useStepReveal(total: number, stepMs = 500, startDelay = 0): number {
  const reduced = useReducedMotion();
  const [visible, setVisible] = useState(reduced ? total : 0);
  useEffect(() => {
    if (reduced) return;
    let n = 0;
    let interval: ReturnType<typeof setInterval>;
    const t = setTimeout(() => {
      interval = setInterval(() => {
        n += 1;
        setVisible(n);
        if (n >= total) clearInterval(interval);
      }, stepMs);
    }, startDelay);
    return () => {
      clearTimeout(t);
      clearInterval(interval);
    };
  }, [total, stepMs, startDelay, reduced]);
  return visible;
}

/** Animated block progress bar. */
function ProgressBar({ width = 22, durationMs = 1600 }: { width?: number; durationMs?: number }) {
  const reduced = useReducedMotion();
  const [pct, setPct] = useState(reduced ? 100 : 0);
  useEffect(() => {
    if (reduced) return;
    const start = performance.now();
    let raf = 0;
    const tick = (now: number) => {
      const p = Math.min(100, Math.round(((now - start) / durationMs) * 100));
      setPct(p);
      if (p < 100) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [durationMs, reduced]);
  const filled = Math.round((pct / 100) * width);
  return (
    <span className="text-term-accent">
      {'█'.repeat(filled)}
      <span className="text-term-dim">{'░'.repeat(width - filled)}</span> {pct}%
    </span>
  );
}

/* --------------------------------- sudo hire-me --------------------------------- */

function HireMe() {
  const step = useStepReveal(6, 550);
  return (
    <div className="font-mono">
      <div>
        Password: <span className="text-term-accent">********</span>
      </div>
      {step >= 1 && (
        <div className="mt-1">
          <Ok>Access Granted.</Ok>
        </div>
      )}
      {step >= 2 && (
        <div className="mt-2">
          <Bold>Congratulations.</Bold>
        </div>
      )}
      {step >= 3 && (
        <div>
          You have successfully hired <Accent>{resume.name}</Accent>.
        </div>
      )}
      {step >= 4 && <div className="mt-2 text-term-dim">Starting onboarding…</div>}
      {step >= 5 && (
        <div className="mt-1">
          <ProgressBar />
        </div>
      )}
      {step >= 5 && (
        <div className="mt-2 text-term-accent2">
          → Run <Accent>contact</Accent> to reach me and make it official. 🚀
        </div>
      )}
    </div>
  );
}

/* ------------------------------------ coffee ------------------------------------ */

function Coffee() {
  const reduced = useReducedMotion();
  const [frame, setFrame] = useState(0);
  useEffect(() => {
    if (reduced) return;
    const id = setInterval(() => setFrame((f) => (f + 1) % 3), 320);
    return () => clearInterval(id);
  }, [reduced]);
  const steam = [
    ['  ) (  ', '  ( )  '],
    ['  ( )  ', '  ) (  '],
    ['  ) (  ', '  ( )  '],
  ][frame];
  return (
    <div className="font-mono leading-tight text-term-fg">
      <div className="text-term-dim">{steam[0]}</div>
      <div className="text-term-dim">{steam[1]}</div>
      <pre className="text-term-accent">{`   .-------.
  |  ~~~~~  |]
   \\       /
    \`-----'`}</pre>
      <div className="mt-2 text-term-fg">
        <Bold>☕ Brewing…</Bold> a DevOps engineer’s true runtime.
      </div>
      <div className="text-term-dim">Fun fact: uptime is powered by caffeine and YAML.</div>
    </div>
  );
}

/* ----------------------------------- fortune ----------------------------------- */

const FORTUNES = [
  'It works on my cluster. — every engineer, before the incident',
  'There are two hard problems in DevOps: cache invalidation, naming things, and off-by-one errors in YAML indentation.',
  'The best time to write the runbook was before the outage. The second best time is now.',
  'Automate the boring, observe the rest, and sleep through the pager.',
  'Infrastructure as Code: because clicking in the console does not scale.',
  'A rollback a day keeps the incident review away.',
  'To err is human; to recover automatically is SRE.',
  'Immutable infrastructure: treat servers like cattle, not pets.',
  '“Have you tried kubectl describe?” — the answer to 80% of questions.',
  'Ship small, ship often, and let the pipeline carry the fear.',
];

function Fortune() {
  const pick = useRef(FORTUNES[Math.floor(Math.random() * FORTUNES.length)]);
  return (
    <div>
      <span className="text-term-accent" aria-hidden="true">
        ✶{' '}
      </span>
      <span className="text-term-fg">{pick.current}</span>
    </div>
  );
}

/* ---------------------------------- neofetch ---------------------------------- */

const NEO_LOGO = ` ____
/ __ \\
\\____/   .-.
 |  |   ( o )
 |  |    \`-'
 '--'  sivaOS`;

function Neofetch() {
  const info: [string, string][] = [
    ['OS', 'SivaneshOS — Cloud-Native Edition'],
    ['Host', resume.title],
    ['Kernel', 'kubernetes-1.x-cka'],
    ['Uptime', `${experienceLabel(resume.careerStartISO)} in DevOps`],
    ['Shell', 'zsh (bash-compatible)'],
    ['DE', 'ArgoCD + GitOps'],
    ['WM', 'Jenkins on Kubernetes'],
    ['Packages', `${allTechnologies.length} technologies`],
    ['CPU', 'Terraform × Ansible (IaC)'],
    ['Certs', 'CKA · Terraform Associate'],
    ['Location', resume.contact.location],
  ];
  const palette = [
    'bg-term-fg',
    'bg-term-accent',
    'bg-term-accent2',
    'bg-term-success',
    'bg-term-warn',
    'bg-term-error',
    'bg-term-link',
    'bg-term-dim',
  ];
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:gap-6">
      <pre className="shrink-0 text-term-accent">{NEO_LOGO}</pre>
      <div className="font-mono">
        <div>
          <Accent>{resume.username}</Accent>
          <span className="text-term-fg">@</span>
          <Accent2>{resume.host}</Accent2>
        </div>
        <div className="text-term-dim">{'-'.repeat(21)}</div>
        {info.map(([k, v]) => (
          <div key={k}>
            <span className="text-term-accent">{k}</span>
            <span className="text-term-fg">: {v}</span>
          </div>
        ))}
        <div className="mt-2 flex">
          {palette.map((c, i) => (
            <span key={i} className={`inline-block h-4 w-4 ${c}`} />
          ))}
        </div>
      </div>
    </div>
  );
}

/* ------------------------------------- hack ------------------------------------- */

const HACK_LINES: { text: string; cls?: string }[] = [
  { text: '[+] Initializing exploit framework…' },
  { text: '[+] Scanning 65 Kubernetes worker nodes… done' },
  { text: '[+] Bypassing RBAC policies… token acquired', cls: 'text-term-warn' },
  { text: '[+] Injecting into Terraform state… ok' },
  { text: '[+] Escalating privileges via Argo Workflows…' },
  { text: '[+] Exfiltrating production secrets… 100%' },
  { text: '[!] JUST KIDDING 😄', cls: 'text-term-accent' },
  { text: '[+] No systems were harmed. Type `help` for real commands.', cls: 'text-term-dim' },
];

function Hack() {
  const visible = useStepReveal(HACK_LINES.length, 420);
  return (
    <div className="font-mono">
      {HACK_LINES.slice(0, visible).map((l, i) => (
        <div key={i} className={l.cls ?? 'text-term-success'}>
          {l.text}
        </div>
      ))}
      {visible < HACK_LINES.length && (
        <span className="inline-block h-[1em] w-[0.6ch] animate-blink bg-term-accent align-middle" />
      )}
    </div>
  );
}

/* ------------------------------------ exports ----------------------------------- */

export const funCommands: Command[] = [
  {
    name: 'coffee',
    description: 'Brew a virtual coffee ☕',
    category: 'fun',
    run: () => <Coffee />,
  },
  {
    name: 'fortune',
    description: 'A random DevOps aphorism',
    category: 'fun',
    run: () => <Fortune />,
  },
  {
    name: 'neofetch',
    description: 'System info, résumé-style',
    category: 'fun',
    run: () => <Neofetch />,
  },
  {
    name: 'matrix',
    description: 'Enter the matrix (ESC / tap to exit)',
    category: 'fun',
    run: (ctx) => {
      ctx.actions.startMatrix();
      return <Muted>Entering the matrix… press ESC or tap to exit.</Muted>;
    },
  },
  {
    name: 'hack',
    description: 'Totally-legit "hacking" sequence',
    category: 'fun',
    run: () => <Hack />,
  },
  {
    name: 'sudo',
    description: 'Run something with elevated privileges 😉',
    usage: 'sudo hire-me',
    category: 'fun',
    run: (ctx) => {
      const target = ctx.args.join(' ').trim().toLowerCase();
      if (target === 'hire-me' || target === 'hireme') return <HireMe />;
      if (target.startsWith('rm -rf') || target.includes('rm -rf /')) {
        return (
          <div className="text-term-error">
            Nice try. 🛡️ This filesystem is immutable (and version-controlled).
          </div>
        );
      }
      if (!target) {
        return (
          <div>
            <Warn>usage:</Warn> <Accent>sudo hire-me</Accent>
          </div>
        );
      }
      return (
        <div className="text-term-warn">
          {resume.username} is not in the sudoers file. This incident will be reported. 😏
        </div>
      );
    },
  },
];
