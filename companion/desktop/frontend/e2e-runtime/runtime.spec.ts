// Runtime end-to-end (companion 4.7 / harness 8.4): the named flows
// against the REAL app — live Wails bindings, real SQLite store, real
// graph runs — served by `wails dev` to an external browser.
import { test, expect } from '@playwright/test';
import { resolve } from 'node:path';

const fixture = resolve(__dirname, '..', '..', '..', 'internal', 'adapters',
  'legacyjson', 'testdata', 'valid_export.json').replaceAll('\\', '/');

test.describe.configure({ mode: 'serial' });

async function bindingsReady(page: import('@playwright/test').Page): Promise<void> {
  await page.waitForFunction(
    () => Boolean((window as never as { go?: { main?: { App?: object } } })
      .go?.main?.App), undefined, { timeout: 30_000 });
}

test('first-run setup detects the real source of this machine', async ({ page }) => {
  await page.goto('/#settings');
  await bindingsReady(page);
  await expect(page.getByText('First-run setup')).toBeVisible({ timeout: 15_000 });
  await expect(page.getByText('Explicit JSON import')).toBeVisible();
  await expect(page.locator('.state-code', { hasText: 'verified' })).toBeVisible();
});

test('a real import runs the graph via the activity registry', async ({ page }) => {
  await page.goto('/');
  await bindingsReady(page);
  const result = await page.evaluate(async (path) => {
    const bridge = window as never as {
      go: { main: { App: {
        ImportCollectionFrom(p: string): Promise<{ outcome: string;
          resolved: number; unresolved: number }> } } };
    };
    return bridge.go.main.App.ImportCollectionFrom(path);
  }, fixture);
  expect(result.outcome).toBe('done');
  expect(result.resolved + result.unresolved).toBeGreaterThan(0);
});

test('home shows the imported totals (successful import flow)', async ({ page }) => {
  await page.goto('/');
  await bindingsReady(page);
  const totals = page.locator('.card-value').nth(3);
  await expect(totals).not.toHaveText('0', { timeout: 15_000 });
});

test('unresolved records stay inspectable in the collection', async ({ page }) => {
  await page.goto('/#collection');
  await bindingsReady(page);
  await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15_000 });
  await page.getByLabel('Search by name or raw text').fill('zzz-no-such-card');
  await expect(page.locator('tbody tr')).toHaveCount(0);
});
