import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      provider: "v8",
      include: ["src/**/*.ts"],
      // Entry points are wiring: index.ts re-exports and main.ts runs on import.
      exclude: ["src/index.ts", "src/main.ts"],
      thresholds: { lines: 85, branches: 85, functions: 85, statements: 85 },
    },
  },
});
