/** Duration between a start date and now, expressed in whole years + months. */
export function experienceDuration(startISO: string, now: Date = new Date()) {
  const start = new Date(startISO);
  let months =
    (now.getFullYear() - start.getFullYear()) * 12 + (now.getMonth() - start.getMonth());
  if (now.getDate() < start.getDate()) months -= 1;
  months = Math.max(0, months);
  return { years: Math.floor(months / 12), months: months % 12, totalMonths: months };
}

/** Human label, e.g. "2 yrs 1 mo". */
export function experienceLabel(startISO: string, now: Date = new Date()): string {
  const { years, months } = experienceDuration(startISO, now);
  return monthsLabel(years * 12 + months);
}

/** Formats a whole-month count as "2 yrs 9 mos". */
function monthsLabel(totalMonths: number): string {
  const years = Math.floor(totalMonths / 12);
  const months = totalMonths % 12;
  const parts: string[] = [];
  if (years > 0) parts.push(`${years} yr${years > 1 ? 's' : ''}`);
  parts.push(`${months} mo${months !== 1 ? 's' : ''}`);
  return parts.join(' ');
}

/** Parses a "MM/YYYY" string into a 1-based month and year. */
function parseMonthYear(s: string): { y: number; m: number } | null {
  const match = s.trim().match(/^(\d{1,2})\/(\d{4})$/);
  if (!match) return null;
  return { m: parseInt(match[1], 10), y: parseInt(match[2], 10) };
}

const PRESENT = new Set(['', 'present', 'current', 'now', 'ongoing']);

/** Whole months of one employment period ("MM/YYYY"; end "Present"/blank = now). */
function periodMonths(start: string, end: string, now: Date): number {
  const s = parseMonthYear(start);
  if (!s) return 0;
  let ey = now.getFullYear();
  let em = now.getMonth() + 1;
  if (!PRESENT.has(end.trim().toLowerCase())) {
    const e = parseMonthYear(end);
    if (e) {
      ey = e.y;
      em = e.m;
    }
  }
  return Math.max(0, (ey - s.y) * 12 + (em - s.m));
}

/**
 * Total professional experience in whole months, summed across every
 * employment period. Career gaps (e.g. for higher studies) are naturally
 * excluded because only actual employment periods are counted.
 */
export function totalExperienceMonths(
  entries: { start: string; end: string }[],
  now: Date = new Date(),
): number {
  return entries.reduce((sum, e) => sum + periodMonths(e.start, e.end, now), 0);
}

/** Human label for total professional experience, e.g. "2 yrs 9 mos". */
export function totalExperienceLabel(
  entries: { start: string; end: string }[],
  now: Date = new Date(),
): string {
  return monthsLabel(totalExperienceMonths(entries, now));
}
