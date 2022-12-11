import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    includeSource: ["src/**/*.ts"],
    coverage: {
      reporter: ["text", "html-spa"],
    },
  },
});
