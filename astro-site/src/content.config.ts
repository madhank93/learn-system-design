import { readdirSync } from 'node:fs';
import { defineCollection, z } from 'astro:content';
import { file } from 'astro/loaders';
import { docsLoader } from '@astrojs/starlight/loaders';
import { docsSchema } from '@astrojs/starlight/schema';

// One collection per syslings topic, read straight from its exercises.yaml so
// topic pages list exactly the exercises the trainer ships.
const EXERCISES_DIR = '../syslings/exercises';

const exercise = z.object({
  title: z.string(),
  desc: z.string(),
  type: z.enum(['fix', 'scenario']),
});

const syslings = Object.fromEntries(
  readdirSync(EXERCISES_DIR, { withFileTypes: true })
    .filter((d) => d.isDirectory())
    .map((d) => [
      `syslings-${d.name}`,
      defineCollection({ loader: file(`${EXERCISES_DIR}/${d.name}/exercises.yaml`), schema: exercise }),
    ]),
);

export const collections = {
  docs: defineCollection({ loader: docsLoader(), schema: docsSchema() }),
  ...syslings,
};
