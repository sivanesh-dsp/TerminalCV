import { type ReactNode } from 'react';
import { resume, allTechnologies } from '@/data/resume';
import { experienceLabel } from '@/utils/format';
import type { Command, CommandContext, TerminalActions } from '@/commands/types';
import {
  Accent,
  Accent2,
  Bullet,
  Bold,
  Dashes,
  Ext,
  Field,
  Heading,
  KV,
  Markup,
  Muted,
  Ok,
  Warn,
} from '@/components/output/ui';

/* ============================ render helpers ============================ */

function AboutView() {
  return (
    <div>
      <Heading title="about" />
      <div className="max-w-3xl leading-relaxed text-term-fg">
        <Markup>{resume.summary}</Markup>
      </div>
      <div className="mt-3">
        <KV label="name" width={10}>
          <Bold>{resume.name}</Bold>
        </KV>
        <KV label="role" width={10}>
          {resume.title}
        </KV>
        <KV label="location" width={10}>
          {resume.contact.location}
        </KV>
        <KV label="focus" width={10}>
          Platform Engineering · DevOps · Kubernetes · IaC · CI/CD
        </KV>
      </div>
    </div>
  );
}

function ExperienceView() {
  return (
    <div>
      <Heading title="experience" />
      {resume.experience.map((e, i) => (
        <div key={i} className={i > 0 ? 'mt-4' : ''}>
          <div className="flex flex-wrap items-baseline justify-between gap-x-4">
            <Bold>{e.role}</Bold>
            <Muted>
              {e.start} – {e.end}
            </Muted>
          </div>
          <div className="text-term-accent2">
            {e.company} <Muted>· {e.location}</Muted>
          </div>
          <div className="mt-2 space-y-1">
            {e.highlights.map((h, j) => (
              <Bullet key={j}>
                <Markup>{h}</Markup>
              </Bullet>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

function ProjectsView() {
  return (
    <div>
      <Heading title="projects" />
      <Muted>Key initiatives drawn from professional experience.</Muted>
      <div className="mt-2 space-y-4">
        {resume.projects.map((p, i) => (
          <div key={i}>
            <div className="text-term-accent">
              ◈ <Bold>{p.name}</Bold>
            </div>
            <div className="ml-4 max-w-3xl text-term-fg">
              <Markup>{p.description}</Markup>
            </div>
            <div className="ml-4 mt-1 flex flex-wrap gap-x-2 gap-y-1 text-xs text-term-dim">
              {p.tech.map((t) => (
                <span key={t}>[{t}]</span>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function SkillsView() {
  return (
    <div>
      <Heading title="skills" />
      <div className="grid grid-cols-1 gap-x-10 sm:grid-cols-2">
        {resume.skills.map((cat) => (
          <Field key={cat.name} label={cat.name}>
            <div className="flex flex-wrap gap-x-2 gap-y-0.5">
              {cat.skills.map((s, i) => (
                <span key={s} className="text-term-fg">
                  {s}
                  {i < cat.skills.length - 1 && <span className="text-term-dim"> ·</span>}
                </span>
              ))}
            </div>
          </Field>
        ))}
      </div>
    </div>
  );
}

function TechStackView() {
  return (
    <div>
      <Heading title="techstack" />
      <Muted>{allTechnologies.length} technologies across the stack.</Muted>
      <div className="mt-2 font-mono text-sm">
        {resume.skills.map((cat, i) => {
          const last = i === resume.skills.length - 1;
          return (
            <div key={cat.name} className="mb-1">
              <div>
                <span className="text-term-dim">{last ? '└─ ' : '├─ '}</span>
                <Accent2>{cat.name}</Accent2>
              </div>
              <div className="flex flex-wrap gap-x-2">
                <span className="text-term-dim">{last ? '   ' : '│  '}</span>
                <span className="flex-1 text-term-fg">
                  {cat.skills.map((s, j) => (
                    <span key={s}>
                      {s}
                      {j < cat.skills.length - 1 && <span className="text-term-dim"> · </span>}
                    </span>
                  ))}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function CertificationsView() {
  return (
    <div>
      <Heading title="certifications" />
      <div className="space-y-2">
        {resume.certifications.map((c, i) => (
          <div key={i}>
            <div className="text-term-fg">
              <Ok>✔</Ok> <Bold>{c.name}</Bold>
            </div>
            {c.issuer && <div className="ml-5 text-term-dim">{c.issuer}</div>}
          </div>
        ))}
      </div>
    </div>
  );
}

function EducationView() {
  return (
    <div>
      <Heading title="education" />
      {resume.education.map((e, i) => (
        <div key={i}>
          <div className="flex flex-wrap items-baseline justify-between gap-x-4">
            <Bold>{e.degree}</Bold>
            <Muted>
              {e.start} – {e.end}
            </Muted>
          </div>
          <div className="text-term-accent2">
            {e.institution}
            {e.location && <Muted> · {e.location}</Muted>}
          </div>
        </div>
      ))}
    </div>
  );
}

function AchievementsView() {
  return (
    <div>
      <Heading title="achievements" />
      <div className="space-y-1">
        {resume.achievements.map((a, i) => (
          <div key={i} className="flex gap-2">
            <span className="text-term-warn" aria-hidden="true">
              ★
            </span>
            <span className="flex-1 text-term-fg">
              <Markup>{a}</Markup>
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function TimelineView() {
  return (
    <div>
      <Heading title="timeline" />
      <div className="font-mono">
        {resume.timeline.map((t, i) => {
          const last = i === resume.timeline.length - 1;
          return (
            <div key={i} className="flex gap-3">
              <div className="w-20 shrink-0 pt-[1px] text-right text-term-accent2">{t.date}</div>
              <div className="flex flex-col items-center">
                <span className="text-term-accent">●</span>
                {!last && <span className="my-0.5 w-px flex-1 bg-term-dim/50" />}
              </div>
              <div className={last ? 'pb-0' : 'pb-4'}>
                <div className="text-term-fg">
                  <Bold>{t.title}</Bold>
                </div>
                {t.subtitle && <div className="text-term-dim">{t.subtitle}</div>}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function StatsView() {
  const employers = new Set(resume.experience.map((e) => e.company)).size;
  const rows: [string, ReactNode][] = [
    ['Experience', `${experienceLabel(resume.careerStartISO)} (since Aug 2024)`],
    ['Technologies', String(allTechnologies.length)],
    ['Certifications', String(resume.certifications.length)],
    ['Projects', String(resume.projects.length)],
    ['Employers', String(employers)],
    ['GitHub repos', <Warn key="g">n/a — not listed on résumé</Warn>],
  ];
  return (
    <div>
      <Heading title="stats" />
      <div className="inline-block rounded-md border border-term-dim/40 p-3">
        {rows.map(([k, v]) => (
          <div key={k} className="flex gap-2 py-0.5">
            <span className="w-32 shrink-0 text-term-dim">{k}</span>
            <span className="text-term-fg">{v}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ---------- contact family ---------- */

function ContactView() {
  const c = resume.contact;
  return (
    <div>
      <Heading title="contact" />
      <div className="space-y-1">
        <KV label="email" width={11}>
          <Ext href={`mailto:${c.email}`}>{c.email}</Ext>
        </KV>
        {c.phone && (
          <KV label="phone" width={11}>
            <Ext href={`tel:${c.phone}`}>{c.phone}</Ext>
          </KV>
        )}
        <KV label="location" width={11}>
          {c.location}
        </KV>
        {c.linkedin && (
          <KV label="linkedin" width={11}>
            <Ext href={c.linkedin.url}>{c.linkedin.url}</Ext>
          </KV>
        )}
        {c.github && (
          <KV label="github" width={11}>
            <Ext href={c.github.url}>{c.github.url}</Ext>
          </KV>
        )}
      </div>
      <div className="mt-2 text-term-dim">
        Tip: <Accent>email</Accent>, <Accent>linkedin</Accent> open the relevant channel.
      </div>
    </div>
  );
}

/* ============================== commands =============================== */

const view = (node: ReactNode) => () => node;

export const infoCommands: Command[] = [
  {
    name: 'about',
    description: 'Who I am — the professional summary',
    category: 'info',
    run: view(<AboutView />),
  },
  {
    name: 'whoami',
    description: 'Print the current user',
    category: 'info',
    run: () => (
      <div>
        <Accent>{resume.username}</Accent> — {resume.name}, {resume.title}.{' '}
        <Muted>Type</Muted> <Accent>about</Accent> <Muted>for the full story.</Muted>
      </div>
    ),
  },
  {
    name: 'experience',
    aliases: ['exp', 'work'],
    description: 'Professional work history',
    category: 'sections',
    run: view(<ExperienceView />),
  },
  {
    name: 'projects',
    description: 'Key engineering initiatives',
    category: 'sections',
    run: view(<ProjectsView />),
  },
  {
    name: 'skills',
    description: 'Technical skills by category',
    category: 'sections',
    run: view(<SkillsView />),
  },
  {
    name: 'techstack',
    aliases: ['stack', 'tech'],
    description: 'Technologies grouped by category',
    category: 'sections',
    run: view(<TechStackView />),
  },
  {
    name: 'certifications',
    aliases: ['certs', 'cert'],
    description: 'Professional certifications',
    category: 'sections',
    run: view(<CertificationsView />),
  },
  {
    name: 'education',
    aliases: ['edu'],
    description: 'Academic background',
    category: 'sections',
    run: view(<EducationView />),
  },
  {
    name: 'achievements',
    aliases: ['awards'],
    description: 'Quantified wins & highlights',
    category: 'sections',
    run: view(<AchievementsView />),
  },
  {
    name: 'timeline',
    description: 'Career journey as an ASCII timeline',
    category: 'sections',
    run: view(<TimelineView />),
  },
  {
    name: 'stats',
    description: 'Snapshot dashboard of the numbers',
    category: 'sections',
    run: view(<StatsView />),
  },
  {
    name: 'contact',
    description: 'All the ways to reach me',
    category: 'social',
    run: view(<ContactView />),
  },
  {
    name: 'email',
    description: 'Email address (mailto)',
    category: 'social',
    run: () => (
      <div>
        <Ext href={`mailto:${resume.contact.email}`}>{resume.contact.email}</Ext>{' '}
        <Muted>— opens your mail client.</Muted>
      </div>
    ),
  },
  {
    name: 'linkedin',
    description: 'LinkedIn profile',
    category: 'social',
    run: () =>
      resume.contact.linkedin ? (
        <div>
          <Ext href={resume.contact.linkedin.url}>{resume.contact.linkedin.url}</Ext>{' '}
          <Muted>(@{resume.contact.linkedin.handle})</Muted>
        </div>
      ) : (
        <Warn>LinkedIn is not listed on the résumé.</Warn>
      ),
  },
  {
    name: 'github',
    description: 'GitHub profile',
    category: 'social',
    run: () =>
      resume.contact.github ? (
        <div>
          <Ext href={resume.contact.github.url}>{resume.contact.github.url}</Ext>{' '}
          <Muted>(@{resume.contact.github.handle})</Muted>
        </div>
      ) : (
        <div className="text-term-warn">
          GitHub is not listed on the résumé, so there is nothing to link here yet.
          <div className="mt-1 text-term-dim">
            Add it later in <Accent>shared/resume.json</Accent> under{' '}
            <Accent>contact.github</Accent>.
          </div>
        </div>
      ),
  },
  {
    name: 'blog',
    description: 'Writing / blog',
    category: 'social',
    run: () => (
      <div className="text-term-warn">
        No blog on the résumé (yet). <Muted>Watch this space.</Muted>
      </div>
    ),
  },
];

/**
 * `resume` renders the full résumé in one scroll and offers the PDF.
 * Kept here so it can reuse every section view above.
 */
function ResumeView({ actions }: { actions: TerminalActions }) {
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center gap-3">
        <div>
          <div className="text-lg font-bold text-term-fg">{resume.name}</div>
          <div className="text-term-dim">{resume.title}</div>
        </div>
        <button
          type="button"
          onClick={actions.downloadResume}
          className="no-print rounded border border-term-accent/60 px-2 py-1 text-sm text-term-accent transition-colors hover:bg-term-accent/10"
        >
          ↓ Download PDF
        </button>
        <button
          type="button"
          onClick={actions.printResume}
          className="no-print rounded border border-term-dim/50 px-2 py-1 text-sm text-term-dim transition-colors hover:border-term-accent hover:text-term-accent"
        >
          ⎙ Print
        </button>
      </div>
      <AboutView />
      <ExperienceView />
      <ProjectsView />
      <SkillsView />
      <CertificationsView />
      <EducationView />
      <div className="pt-1">
        <Dashes len={40} />
        <div className="text-term-dim">
          Run <Accent>download</Accent> to save the original PDF, or <Accent>contact</Accent> to
          reach me.
        </div>
      </div>
    </div>
  );
}

export const resumeCommand: Command = {
  name: 'resume',
  aliases: ['cv'],
  description: 'View the full résumé (with PDF download)',
  category: 'sections',
  run: (ctx: CommandContext) => <ResumeView actions={ctx.actions} />,
};

// Re-export the section views for reuse (e.g. `cat`).
export const sectionViews = {
  about: <AboutView />,
  experience: <ExperienceView />,
  projects: <ProjectsView />,
  skills: <SkillsView />,
  certifications: <CertificationsView />,
  education: <EducationView />,
  achievements: <AchievementsView />,
  timeline: <TimelineView />,
  stats: <StatsView />,
  contact: <ContactView />,
};
