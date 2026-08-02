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
  const parts: string[] = [];
  if (years > 0) parts.push(`${years} yr${years > 1 ? 's' : ''}`);
  parts.push(`${months} mo${months !== 1 ? 's' : ''}`);
  return parts.join(' ');
}
