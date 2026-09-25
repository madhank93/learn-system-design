import { execFileSync } from 'node:child_process';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

// Every code sample under "## This implementation" must be copied from the
// repo dir its <BreakIt path> points at. Illustrative fences (text/bash) and
// sketches inside <details> are exempt.
const DOCS = new URL('.', import.meta.url).pathname;
const HLD = join(DOCS, '../../../../../high-level-design');
const EXEMPT = new Set(['', 'text', 'bash', 'sh', 'console']);

const norm = (s: string) => s.trim().replace(/\s+/g, ' ');

// Tracked files only: the page links to GitHub, so that's the code readers see.
function repoLines(dir: string): Set<string> {
  const files = execFileSync('git', ['ls-files', '-z'], { cwd: dir, encoding: 'utf8' }).split('\0').filter(Boolean);
  return new Set(files.flatMap((f) => readFileSync(join(dir, f), 'utf8').split('\n').map(norm)));
}

function snippets(mdx: string): { lang: string; lines: string[] }[] {
  const section = mdx.split(/^## This implementation$/m)[1]?.split(/^## /m)[0] ?? '';
  const outsideDetails = section.replace(/<details>[\s\S]*?<\/details>/g, '');
  return [...outsideDetails.matchAll(/^```(\w*)\n([\s\S]*?)^```/gm)]
    .filter(([, lang]) => !EXEMPT.has(lang))
    .map(([, lang, body]) => ({ lang, lines: body.split('\n').map(norm).filter(Boolean) }));
}

describe('"This implementation" snippets match the repo', () => {
  for (const file of readdirSync(DOCS).filter((f) => f.endsWith('.mdx'))) {
    const mdx = readFileSync(join(DOCS, file), 'utf8');
    const path = mdx.match(/<BreakIt path="([^"]+)"/)?.[1];
    const blocks = snippets(mdx);
    if (!path || blocks.length === 0) continue;

    it(file, () => {
      const repo = repoLines(join(HLD, path));
      for (const { lang, lines } of blocks) {
        const missing = lines.filter((l) => !repo.has(l));
        expect(missing, `${lang} lines not found under high-level-design/${path}`).toEqual([]);
      }
    });
  }
});
