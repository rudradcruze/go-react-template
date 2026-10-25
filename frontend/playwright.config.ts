import { defineConfig, devices } from "@playwright/test";

// scripts/test-web-e2e.sh starts the stack on this port, in a Compose project of its own.
// To run the tests against a stack that already runs, set E2E_WEB_PORT and run: pnpm --filter frontend run test:e2e
const port = process.env.E2E_WEB_PORT ?? "5655";

export default defineConfig({
  testDir: "test/e2e",
  // Each test makes its own data, so the tests can run at the same time.
  fullyParallel: true,
  retries: 0,
  reporter: [["list"], ["html", { open: "never" }]],
  use: {
    baseURL: `http://localhost:${port}`,
    trace: "retain-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
