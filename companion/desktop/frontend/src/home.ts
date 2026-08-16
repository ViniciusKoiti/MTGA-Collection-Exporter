// Home view (task 4.3): renders the model composed by Go
// (application/homesvc) — presence, source health, last sync,
// totals, delta and the in-context recovery actions.
import { Home } from '../wailsjs/go/main/App';
import { renderState, type ViewState } from './state';

type HomeModel = {
  mtga_running: boolean;
  source: string;
  source_healthy: boolean;
  last_sync?: string;
  totals: number;
  delta?: number;
  state: ViewState;
  actions?: { id: string; label: string }[];
};

function card(label: string, value: string): string {
  return `<div class="card"><span class="card-label">${label}</span>` +
    `<span class="card-value">${value}</span></div>`;
}

export async function renderHome(container: HTMLElement): Promise<void> {
  container.innerHTML = '<p class="placeholder">Composing your overview…</p>';
  let model: HomeModel;
  try {
    // The generated binding types status as plain string; the Go
    // contract guarantees the closed vocabulary.
    model = (await Home()) as unknown as HomeModel;
  } catch {
    container.innerHTML =
      '<p class="placeholder">Home needs the desktop runtime.</p>';
    return;
  }
  const delta = model.delta === undefined || model.delta === null
    ? '—'
    : model.delta >= 0 ? `+${model.delta}` : `${model.delta}`;
  const actions = (model.actions ?? [])
    .map((action) =>
      `<button type="button" data-action="${action.id}">${action.label}</button>`)
    .join(' ');
  container.innerHTML =
    renderState(model.state) +
    `<div class="cards">` +
    card('MTGA', model.mtga_running ? 'Running' : 'Not running') +
    card('Source', `${model.source}${model.source_healthy ? ' ✓' : ' ✗'}`) +
    card('Last sync', model.last_sync ?? 'never') +
    card('Cards', String(model.totals)) +
    card('Delta', delta) +
    `</div>` +
    (actions ? `<div class="actions">${actions}</div>` : '');
  container.querySelectorAll<HTMLButtonElement>('button[data-action]')
    .forEach((button) => {
      button.addEventListener('click', () => {
        if (button.dataset.action === 'import_collection') {
          location.hash = 'settings';
          return;
        }
        void renderHome(container); // resync / retry re-compose
      });
    });
}
