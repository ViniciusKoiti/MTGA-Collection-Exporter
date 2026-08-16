// Component tests for the primary flows (task 4.7): setup, successful
// import, stale collection, failed sync recovery and unresolved-card
// inspection — the Wails bindings are mocked, the real renderers run.
import { beforeEach, describe, expect, it, vi } from 'vitest';

const bindings = vi.hoisted(() => ({
  Views: vi.fn(),
  FirstRun: vi.fn(),
  Home: vi.fn(),
  Collection: vi.fn(),
  CompareSnapshots: vi.fn(),
}));

vi.mock('../../wailsjs/go/main/App', () => bindings);

import { renderFirstRun } from '../firstrun';
import { renderHome } from '../home';
import { renderCollection } from '../collection';

function container(): HTMLElement {
  const el = document.createElement('div');
  document.body.appendChild(el);
  return el;
}

const tick = (): Promise<void> => new Promise((r) => setTimeout(r, 0));

beforeEach(() => {
  vi.clearAllMocks();
  document.body.innerHTML = '';
});

describe('first-run setup', () => {
  it('renders the detected source, guidance and explicit privacy', async () => {
    bindings.FirstRun.mockResolvedValue({
      source: 'json_import', needs_guidance: true, verified: true,
      fell_back: false,
      privacy: { telemetry_opt_in: false, assistant_enabled: false },
    });
    const el = container();
    await renderFirstRun(el);
    expect(el.textContent).toContain('Explicit JSON import');
    expect(el.textContent).toContain('Detailed Logs guidance');
    expect(el.textContent).toContain('Nothing is shared unless you opt in');
    el.querySelector<HTMLInputElement>('#opt-telemetry')!.click();
    await tick();
    expect(bindings.FirstRun).toHaveBeenLastCalledWith(true, false);
  });
});

describe('home', () => {
  it('offers resync in context when the collection is stale', async () => {
    bindings.Home.mockResolvedValue({
      mtga_running: true, source: 'json_import', source_healthy: true,
      last_sync: '2026-08-14T10:00:00Z', totals: 900,
      state: { status: 'stale', detail: 'last sync is older than 24h' },
      actions: [{ id: 'resync', label: 'Sync now' }],
    });
    const el = container();
    await renderHome(el);
    expect(el.textContent).toContain('Showing an older snapshot');
    expect(el.querySelector('button[data-action="resync"]')!.textContent)
      .toBe('Sync now');
  });

  it('recovers from a failed sync through the retry action', async () => {
    bindings.Home.mockResolvedValue({
      mtga_running: false, source: 'json_import', source_healthy: false,
      totals: 0,
      state: { status: 'error', code: 'source_unavailable',
        detail: 'collection.sync (source_unavailable)' },
      actions: [{ id: 'retry_sync', label: 'Retry sync' }],
    });
    const el = container();
    await renderHome(el);
    expect(el.textContent).toContain('source_unavailable');
    el.querySelector<HTMLButtonElement>('button[data-action="retry_sync"]')!
      .click();
    await tick();
    expect(bindings.Home).toHaveBeenCalledTimes(2);
  });
});

describe('collection', () => {
  it('renders a successful import as rows', async () => {
    bindings.Collection.mockResolvedValue({
      rows: [
        { printing: 'p1', name: 'Lightning Strike', set: 'DMU',
          quantity: 4, unresolved: false, oracle: 'o-1', arena: 101 },
        { printing: 'p2', name: 'Counterspell', set: 'DMU',
          quantity: 2, unresolved: false, oracle: 'o-2', arena: 102 },
      ],
      state: { status: 'success' },
    });
    const el = container();
    renderCollection(el);
    await tick();
    expect(el.querySelectorAll('tbody tr')).toHaveLength(2);
    expect(el.textContent).toContain('Lightning Strike');
  });

  it('inspects unresolved cards and filters down to them', async () => {
    bindings.Collection.mockResolvedValue({
      rows: [{ printing: '', name: '', set: '', quantity: 1,
        unresolved: true, raw: '??? mystery card' }],
      state: { status: 'success' },
    });
    const el = container();
    renderCollection(el);
    await tick();
    expect(el.textContent).toContain('unresolved');
    expect(el.textContent).toContain('raw: ??? mystery card');
    el.querySelector<HTMLInputElement>('#col-unresolved')!.click();
    await tick();
    const lastCall = bindings.Collection.mock.lastCall![0];
    expect(lastCall.unresolved_only).toBe(true);
  });
});
