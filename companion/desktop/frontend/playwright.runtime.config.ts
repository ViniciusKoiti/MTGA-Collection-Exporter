import { defineConfig } from '@playwright/test';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

// Runtime e2e: `wails dev` runs the REAL app (live bindings, real
// engine, real store) and bridges it to external browsers on :34115.
// APPDATA is pointed at a temp dir so the run never touches the
// user's real companion.db.
export default defineConfig({
  testDir: 'e2e-runtime',
  workers: 1,
  timeout: 60_000,
  use: { baseURL: 'http://localhost:34115' },
  webServer: {
    command: 'wails dev',
    cwd: join(__dirname, '..'),
    url: 'http://localhost:34115',
    reuseExistingServer: !process.env.CI,
    timeout: 180_000,
    env: {
      ...process.env,
      // The app-level override keeps the test run away from the real
      // companion.db without disturbing APPDATA (wails needs it).
      COMPANION_DATA_DIR: mkdtempSync(join(tmpdir(), 'companion-e2e-')),
      GOPROXY: 'direct',
    },
  },
});
