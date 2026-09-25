// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import react from '@astrojs/react';
import rehypeMermaid from 'rehype-mermaid';

export default defineConfig({
  markdown: {
    // Render ```mermaid fences to inline SVG at build time (headless Chromium via
    // playwright) so static diagrams ship zero client JS. Interactive islands
    // (SystemDiagram, sims) stay as React components.
    rehypePlugins: [[rehypeMermaid, { strategy: 'inline-svg', mermaidConfig: { theme: 'dark', themeVariables: { edgeLabelBackground: '#24272f' } } }]],
  },
  integrations: [
    starlight({
      title: 'Learn System Design',
      description: 'Astro + Starlight migration proof slice with interactive diagrams.',
      sidebar: [
        {
          label: 'Networking & Edge',
          items: [{ label: 'Load Balancing', slug: 'hld/load-balancing' }],
        },
        {
          label: 'Scalability',
          items: [
            { label: 'Caching', slug: 'hld/caching' },
            { label: 'Rate Limiting', slug: 'hld/rate-limiting' },
            { label: 'Database Replication', slug: 'hld/db-replication' },
          ],
        },
        {
          label: 'Resilience',
          items: [{ label: 'Circuit Breaker', slug: 'hld/circuit-breaker' }],
        },
        {
          label: 'Data & Consistency',
          items: [{ label: 'Consistent Hashing', slug: 'hld/consistent-hashing' }],
        },
      ],
      customCss: ['./src/styles/custom.css'],
    }),
    react(),
  ],
});
