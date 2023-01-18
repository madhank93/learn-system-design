import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    includeSource: ["src/**/*.ts", "test/**/*.ts"],
    coverage: {
      reporter: ["text", "html"],
    },
  },
});
