import { resume } from '@/data/resume';

export interface SearchHit {
  section: string;
  text: string;
}

/** Removes inline `**highlight**` markers so they never show in search text. */
const clean = (s: string): string => s.replace(/\*\*/g, '');

/** Build a flat, searchable corpus from every résumé section. */
function corpus(): SearchHit[] {
  const hits: SearchHit[] = [{ section: 'summary', text: clean(resume.summary) }];

  resume.experience.forEach((e) => {
    hits.push({
      section: `experience · ${e.company}`,
      text: `${e.role} at ${e.company}, ${e.location} (${e.start}–${e.end})`,
    });
    e.highlights.forEach((h) => hits.push({ section: `experience · ${e.company}`, text: clean(h) }));
  });

  resume.skills.forEach((c) =>
    c.skills.forEach((s) => hits.push({ section: `skills · ${c.name}`, text: s })),
  );

  resume.projects.forEach((p) =>
    hits.push({
      section: `project · ${p.name}`,
      text: `${p.name} — ${clean(p.description)} [${p.tech.join(', ')}]`,
    }),
  );

  resume.certifications.forEach((c) =>
    hits.push({ section: 'certifications', text: c.issuer ? `${c.name} — ${c.issuer}` : c.name }),
  );

  resume.education.forEach((e) =>
    hits.push({
      section: 'education',
      text: `${e.degree} — ${e.institution}${e.location ? ', ' + e.location : ''} (${e.start}–${e.end})`,
    }),
  );

  resume.achievements.forEach((a) => hits.push({ section: 'achievements', text: clean(a) }));

  return hits;
}

/** Case-insensitive substring search across the whole résumé. */
export function searchResume(query: string): SearchHit[] {
  const q = query.trim().toLowerCase();
  if (!q) return [];
  return corpus().filter((c) => c.text.toLowerCase().includes(q));
}

/**
 * Split `text` around every case-insensitive occurrence of `query`.
 * Returns segments flagged as matches so the renderer can highlight them.
 */
export function splitHighlights(
  text: string,
  query: string,
): { text: string; match: boolean }[] {
  const q = query.trim();
  if (!q) return [{ text, match: false }];
  const out: { text: string; match: boolean }[] = [];
  const lower = text.toLowerCase();
  const needle = q.toLowerCase();
  let i = 0;
  while (i < text.length) {
    const idx = lower.indexOf(needle, i);
    if (idx === -1) {
      out.push({ text: text.slice(i), match: false });
      break;
    }
    if (idx > i) out.push({ text: text.slice(i, idx), match: false });
    out.push({ text: text.slice(idx, idx + needle.length), match: true });
    i = idx + needle.length;
  }
  return out;
}
