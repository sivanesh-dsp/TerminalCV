import type { ResumeData } from '@/types/resume';
import data from '@shared/resume.json';
import { fetchResumeFromApi } from '@/services/resumeApi';

/**
 * Single source of truth for EVERY surface.
 *
 * The résumé content lives in `shared/resume.json` at the repo root and is
 * consumed verbatim by:
 *   - this React website (imported + bundled at build time),
 *   - the Go SSH server (`ssh/`, loaded at runtime), and
 *   - the résumé REST API (`resume-api`, mounted read-only at runtime).
 *
 * Editing `shared/resume.json` updates all three — there is no duplicated
 * résumé data anywhere in the project.
 *
 * At runtime the site additionally hydrates from `GET /api/v1/*` (see
 * `hydrateFromApi`), so a data change is visible without rebuilding the site.
 * The bundled copy is always the fallback, so the terminal works offline and
 * whenever the API is unavailable.
 */
export const resume: ResumeData = { ...(data as ResumeData) };

function flattenTechnologies(source: ResumeData): string[] {
  return Array.from(new Set(source.skills.flatMap((c) => c.skills)));
}

/** Flattened, de-duplicated technology list — used by `stats` and search. */
export const allTechnologies: string[] = flattenTechnologies(resume);

/**
 * Merge a patch into the shared résumé object *in place* so every module that
 * imported `resume` (all commands do) sees the update without re-importing.
 * Returns true when something actually changed.
 */
export function applyResumeUpdate(patch: Partial<ResumeData>): boolean {
  const before = JSON.stringify(resume);
  Object.assign(resume, patch);

  allTechnologies.splice(0, allTechnologies.length, ...flattenTechnologies(resume));

  return JSON.stringify(resume) !== before;
}

/**
 * Refresh the résumé from the REST API. Never throws and never blocks the UI —
 * on any failure the bundled data simply stays in place.
 */
export async function hydrateFromApi(): Promise<boolean> {
  const patch = await fetchResumeFromApi();
  return patch ? applyResumeUpdate(patch) : false;
}
