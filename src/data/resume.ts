import type { ResumeData } from '@/types/resume';
import data from '@shared/resume.json';

/**
 * Single source of truth for BOTH frontends.
 *
 * The résumé content lives in `shared/resume.json` at the repo root and is
 * consumed verbatim by:
 *   - this React website (imported + bundled at build time), and
 *   - the Go SSH server (`ssh/`, loaded at runtime).
 *
 * Editing `shared/resume.json` updates both experiences — there is no
 * duplicated résumé data anywhere in the project.
 */
export const resume = data as ResumeData;

/** Flattened, de-duplicated technology list — used by `stats` and search. */
export const allTechnologies: string[] = Array.from(
  new Set(resume.skills.flatMap((c) => c.skills)),
);
