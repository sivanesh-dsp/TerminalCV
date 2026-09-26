import type {
  Certification,
  ContactInfo,
  EducationItem,
  ExperienceItem,
  Project,
  ResumeData,
  SkillCategory,
  TimelineEvent,
} from '@/types/resume';

/**
 * Client for the résumé REST API (https://sivaneshbalaji.online/api/v1).
 *
 * The site ships with `shared/resume.json` bundled, so it works with no network
 * at all. When the API is reachable the in-memory résumé is hydrated from it,
 * so updating the data no longer requires a site rebuild. Both read the exact
 * same source file, so there is no second source of truth either way.
 *
 * Every failure is silent by design: the bundled data is always a valid
 * fallback, and the terminal must never show a network error.
 */

/** Same-origin by default — Caddy proxies `/api/*` to the FastAPI service. */
const API_BASE = (import.meta.env.VITE_API_BASE_URL ?? '/api/v1').replace(/\/+$/, '');

/** Hydration must never delay the terminal; give up quickly. */
const TIMEOUT_MS = 4000;

export type ApiStatus = 'idle' | 'live' | 'unavailable';

let status: ApiStatus = 'idle';

/** Last known API state — surfaced by the `api` command. */
export const getApiStatus = (): ApiStatus => status;

export const apiBaseUrl = (): string => API_BASE;

async function getJson<T>(path: string, signal: AbortSignal): Promise<T | null> {
  try {
    const res = await fetch(`${API_BASE}${path}`, {
      signal,
      headers: { Accept: 'application/json' },
    });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch {
    return null;
  }
}

/* ------------------------------ response shapes ----------------------------- */

interface ProfileResponse {
  name: string;
  title: string;
  summary: string;
}
interface SkillsResponse {
  categories: { name: string; skills: string[] }[];
}
interface ExperienceResponse {
  experience: {
    company: string;
    role: string;
    start: string;
    end: string;
    location: string | null;
    highlights: string[];
  }[];
}
interface ProjectsResponse {
  projects: { name: string; description: string; technologies: string[] }[];
}
interface CertificationsResponse {
  certifications: { name: string; issuer: string | null }[];
}
interface EducationResponse {
  education: {
    degree: string;
    institution: string;
    location: string | null;
    start: string | null;
    end: string | null;
  }[];
}
interface AchievementsResponse {
  achievements: string[];
}
interface TimelineResponse {
  timeline: { date: string; title: string; description: string | null }[];
}
interface ContactResponse {
  email: string;
  location: string;
  github: { handle: string; url: string } | null;
  linkedin: { handle: string; url: string } | null;
}

function nonEmpty<T>(value: T[] | undefined | null): value is T[] {
  return Array.isArray(value) && value.length > 0;
}

/**
 * Fetch the résumé from the API and map it back onto the site's data model.
 * Returns only the sections that came back well-formed.
 */
export async function fetchResumeFromApi(): Promise<Partial<ResumeData> | null> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);
  const { signal } = controller;

  try {
    const [
      profile,
      skills,
      experience,
      projects,
      certifications,
      education,
      achievements,
      timeline,
      contact,
    ] = await Promise.all([
      getJson<ProfileResponse>('/profile', signal),
      getJson<SkillsResponse>('/skills', signal),
      getJson<ExperienceResponse>('/experience', signal),
      getJson<ProjectsResponse>('/projects', signal),
      getJson<CertificationsResponse>('/certifications', signal),
      getJson<EducationResponse>('/education', signal),
      getJson<AchievementsResponse>('/achievements', signal),
      getJson<TimelineResponse>('/timeline', signal),
      getJson<ContactResponse>('/contact', signal),
    ]);

    if (!profile?.name) {
      status = 'unavailable';
      return null;
    }

    const patch: Partial<ResumeData> = {
      name: profile.name,
      title: profile.title,
      summary: profile.summary,
    };

    if (skills && nonEmpty(skills.categories)) {
      patch.skills = skills.categories.map<SkillCategory>((c) => ({
        name: c.name,
        skills: c.skills,
      }));
    }

    if (experience && nonEmpty(experience.experience)) {
      patch.experience = experience.experience.map<ExperienceItem>((e) => ({
        role: e.role,
        company: e.company,
        start: e.start,
        end: e.end,
        location: e.location ?? '',
        highlights: e.highlights,
      }));
    }

    if (projects && nonEmpty(projects.projects)) {
      patch.projects = projects.projects.map<Project>((p) => ({
        name: p.name,
        description: p.description,
        tech: p.technologies,
      }));
    }

    if (certifications && nonEmpty(certifications.certifications)) {
      patch.certifications = certifications.certifications.map<Certification>((c) => ({
        name: c.name,
        ...(c.issuer ? { issuer: c.issuer } : {}),
      }));
    }

    if (education && nonEmpty(education.education)) {
      patch.education = education.education.map<EducationItem>((e) => ({
        degree: e.degree,
        institution: e.institution,
        start: e.start ?? '',
        end: e.end ?? '',
        ...(e.location ? { location: e.location } : {}),
      }));
    }

    if (achievements && nonEmpty(achievements.achievements)) {
      patch.achievements = achievements.achievements;
    }

    if (timeline && nonEmpty(timeline.timeline)) {
      patch.timeline = timeline.timeline.map<TimelineEvent>((t) => ({
        date: t.date,
        title: t.title,
        ...(t.description ? { subtitle: t.description } : {}),
      }));
    }

    if (contact?.email) {
      const next: ContactInfo = {
        email: contact.email,
        location: contact.location,
        ...(contact.github ? { github: contact.github } : {}),
        ...(contact.linkedin ? { linkedin: contact.linkedin } : {}),
      };
      patch.contact = next;
    }

    status = 'live';
    return patch;
  } catch {
    status = 'unavailable';
    return null;
  } finally {
    clearTimeout(timer);
  }
}
