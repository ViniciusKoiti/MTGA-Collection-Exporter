import { defineConfig } from '@playwright/test';

// End-to-end suite of the shell (tasks 4.7/7.3): runs against the
// built frontend served by vite preview — no MTGA, no network beyond
// localhost, no administrator privileges, no LLM.
export default defineConfig({
  testDir: 'e2e',
  use: { baseURL: 'http://127.0.0.1:4173' },
  webServer: {
    command:
      'npm run build && npm run preview -- --host 127.0.0.1 --port 4173 --strictPort',
    url: 'http://127.0.0.1:4173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
