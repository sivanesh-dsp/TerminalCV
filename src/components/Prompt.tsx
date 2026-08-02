import { resume } from '@/data/resume';

/**
 * The fake shell prompt: `sivanesh@portfolio:~$`.
 * Colour-segmented but rendered as a single inline element so it wraps well.
 */
export function Prompt({ path = '~' }: { path?: string }) {
  return (
    <span className="select-none whitespace-nowrap" aria-hidden="true">
      <span className="text-term-prompt">{resume.username}</span>
      <span className="text-term-dim">@</span>
      <span className="text-term-accent2">{resume.host}</span>
      <span className="text-term-dim">:</span>
      <span className="text-term-link">{path}</span>
      <span className="text-term-dim">$&nbsp;</span>
    </span>
  );
}

export const promptText = `${resume.username}@${resume.host}:~$`;
