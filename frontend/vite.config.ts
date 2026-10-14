import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  server: {
    // The dev server uses the same /api paths as nginx in the Compose stack.
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/main.tsx", "src/test/**", "src/**/*.test.{ts,tsx}"],
      reporter: [["text", { skipFull: false }], "html", "json-summary"],
      reportsDirectory: "../coverage/frontend",
      thresholds: { lines: 85 },
    },
  },
});
