/**
 * Strongly-typed model for the résumé data that powers every command.
 * All values are sourced directly from the attached PDF — nothing is invented.
 * Optional fields are omitted when the source résumé does not provide them.
 */

export interface SocialLink {
  /** Handle as shown on the résumé, e.g. "sivaneshb". */
  handle: string;
  /** Fully-qualified URL. */
  url: string;
}

export interface ContactInfo {
  email: string;
  phone?: string;
  location: string;
  linkedin?: SocialLink;
  github?: SocialLink;
  portfolio?: SocialLink;
}

export interface SkillCategory {
  /** Category label as grouped on the résumé. */
  name: string;
  /** Short slug used for the ASCII underline + search. */
  skills: string[];
}

export interface ExperienceItem {
  role: string;
  company: string;
  /** Human-readable start, e.g. "08/2024". */
  start: string;
  /** Human-readable end, e.g. "Present". */
  end: string;
  location: string;
  highlights: string[];
}

export interface Certification {
  name: string;
  issuer?: string;
}

export interface EducationItem {
  degree: string;
  institution: string;
  start: string;
  end: string;
  location?: string;
}

export interface Project {
  name: string;
  description: string;
  tech: string[];
}

export interface TimelineEvent {
  /** Sort key + label, e.g. "2024" or "Aug 2024". */
  date: string;
  title: string;
  subtitle?: string;
  /**
   * Optional semantic category consumed by the REST API's /api/v1/timeline
   * endpoint ("experience" | "education" | "certification" | "milestone").
   * When omitted the API infers it. Unused by the website itself.
   */
  type?: string;
}

/** Published SSH host key, used to verify the server on first connect. */
export interface SshHostKey {
  /** Key algorithm, e.g. `ed25519`. */
  type: string;
  /** `ssh-keygen -lf` style fingerprint, e.g. `SHA256:...`. */
  fingerprint: string;
  /** Full `known_hosts` key material, e.g. `ssh-ed25519 AAAA...`. */
  publicKey: string;
}

export interface ResumeData {
  name: string;
  title: string;
  /** Shell username shown in the prompt: `<username>@<host>:~$`. */
  username: string;
  host: string;
  summary: string;
  contact: ContactInfo;
  skills: SkillCategory[];
  experience: ExperienceItem[];
  certifications: Certification[];
  education: EducationItem[];
  /** Derived strictly from experience bullets — real initiatives, not invented. */
  projects: Project[];
  /** Quantified wins pulled from experience bullets. */
  achievements: string[];
  timeline: TimelineEvent[];
  /** ISO date the current role started — used to compute years of experience. */
  careerStartISO: string;
  /** Path (relative to the site base) to the downloadable PDF. */
  resumeFile: string;
  /** Hostname the SSH portfolio is reachable at, e.g. `ssh <sshHost>`. */
  sshHost?: string;
  /**
   * Public host key of the SSH portfolio, published so visitors can verify the
   * fingerprint they are shown on first connect (or pre-seed `known_hosts`).
   */
  sshHostKey?: SshHostKey;
}
