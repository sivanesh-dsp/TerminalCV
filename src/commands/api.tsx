import { apiBaseUrl, getApiStatus } from '@/services/resumeApi';
import type { Command } from '@/commands/types';
import {
  Accent,
  Accent2,
  Bold,
  CopyButton,
  Ext,
  Heading,
  KV,
  Muted,
  Ok,
  Warn,
} from '@/components/output/ui';

/**
 * `api` — documents the REST API that backs this terminal.
 *
 * The website, the SSH terminal and the API all read the same
 * `shared/resume.json`; this command makes that architecture discoverable from
 * inside the portfolio itself.
 */

/** Endpoint catalogue, mirroring the API's own `GET /api/v1` index. */
const ENDPOINTS: [path: string, description: string][] = [
  ['/profile', 'name, title, summary, location, highlights'],
  ['/skills', 'skills grouped by category'],
  ['/experience', 'roles with computed durations (+ /{id})'],
  ['/projects', 'projects and their tech stack (+ /{id})'],
  ['/certifications', 'certifications (+ /{id})'],
  ['/education', 'formal education'],
  ['/achievements', 'quantified achievements'],
  ['/contact', 'public contact channels'],
  ['/timeline', 'chronological career timeline'],
  ['/search?q=', 'search everything, case-insensitive'],
  ['/stats', 'statistics computed from the résumé'],
  ['/resume', 'the PDF résumé'],
  ['/health', 'liveness / readiness'],
  ['/version', 'service + contract version'],
];

function absolute(path: string): string {
  const base = apiBaseUrl();
  if (base.startsWith('http')) return `${base}${path}`;
  return typeof window === 'undefined' ? path : `${window.location.origin}${base}${path}`;
}

function ApiView() {
  const status = getApiStatus();
  const example = absolute('/profile');

  return (
    <div>
      <Heading title="rest api" />

      <div className="max-w-3xl leading-relaxed text-term-fg">
        This portfolio is API-first: the same <Accent2>shared/resume.json</Accent2> powers the
        website, the SSH terminal and a versioned REST API — so nothing is ever duplicated.
      </div>

      <div className="mt-3">
        <KV label="base" width={10}>
          <Bold>{absolute('')}</Bold>
        </KV>
        <KV label="docs" width={10}>
          <Ext href={absolute('').replace('/v1', '/docs')}>{absolute('').replace('/v1', '/docs')}</Ext>
        </KV>
        <KV label="openapi" width={10}>
          <Ext href={absolute('').replace('/v1', '/openapi.json')}>
            {absolute('').replace('/v1', '/openapi.json')}
          </Ext>
        </KV>
        <KV label="status" width={10}>
          {status === 'live' ? (
            <Ok>live · this page is hydrated from the API</Ok>
          ) : status === 'unavailable' ? (
            <Warn>unreachable · showing the bundled résumé data</Warn>
          ) : (
            <Muted>checking…</Muted>
          )}
        </KV>
      </div>

      <div className="mt-4">
        <Accent>endpoints</Accent>
        <div className="mt-1 space-y-0.5">
          {ENDPOINTS.map(([path, description]) => (
            <div key={path} className="flex flex-wrap gap-x-2">
              <span className="text-term-accent2">
                GET /api/v1{path.padEnd(18, ' ')}
              </span>
              <Muted>{description}</Muted>
            </div>
          ))}
        </div>
      </div>

      <div className="mt-4">
        <Accent>try it</Accent>
        <div className="mt-1 flex flex-wrap items-center gap-2">
          <code className="text-term-fg">curl -s {example} | jq</code>
          <CopyButton text={`curl -s ${example} | jq`} label="copy curl" />
        </div>
      </div>

      <div className="mt-4 max-w-3xl">
        <Muted>
          Read-only, versioned and strongly typed — designed to be consumed by an MCP server so AI
          agents can answer questions about this résumé directly.
        </Muted>
      </div>
    </div>
  );
}

export const apiCommand: Command = {
  name: 'api',
  aliases: ['rest', 'endpoints'],
  description: 'REST API powering this portfolio (endpoints, docs, status)',
  category: 'info',
  run: () => <ApiView />,
};

export const apiCommands: Command[] = [apiCommand];
