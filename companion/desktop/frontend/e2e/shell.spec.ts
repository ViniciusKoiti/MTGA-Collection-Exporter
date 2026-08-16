// Shell end-to-end (task 7.3): a real browser drives the five-view
// navigation, the keyboard contract and the out-of-runtime fallbacks.
// The Wails bindings are absent here by design, so every view must
// still render something honest instead of breaking.
import { test, expect } from '@playwright/test';

test('navigates the five views by mouse', async ({ page }) => {
  await page.goto('/');
  const nav = page.getByRole('navigation', { name: 'Main navigation' });
  await expect(nav.getByRole('button')).toHaveCount(5);
  for (const label of ['Collection', 'Decks', 'Assistant', 'Settings', 'Home']) {
    await nav.getByRole('button', { name: label }).click();
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(label);
  }
});

test('walks the navigation with the keyboard', async ({ page }) => {
  await page.goto('/');
  const nav = page.getByRole('navigation', { name: 'Main navigation' });
  await nav.getByRole('button', { name: 'Home' }).focus();
  await page.keyboard.press('ArrowDown');
  await page.keyboard.press('ArrowDown');
  await page.keyboard.press('Enter');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Decks');
  await page.keyboard.press('End');
  await expect(nav.getByRole('button', { name: 'Settings' })).toBeFocused();
});

test('views degrade honestly without the desktop runtime', async ({ page }) => {
  await page.goto('/#settings');
  await expect(page.getByText('needs the desktop runtime')).toBeVisible();
  await page.goto('/#decks');
  await expect(page.getByLabel('Deck list in Arena format')).toBeVisible();
  await page.goto('/#collection');
  await expect(page.getByLabel('Search by name or raw text')).toBeVisible();
});
