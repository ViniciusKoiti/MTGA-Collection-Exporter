import './style.css';
import { Views } from '../wailsjs/go/main/App';

type View = { id: string; label: string };

// fallbackViews mirrors the Go contract for dev outside the runtime;
// inside Wails the bound Views() is the single source of truth.
const fallbackViews: View[] = [
  { id: 'home', label: 'Home' },
  { id: 'collection', label: 'Collection' },
  { id: 'decks', label: 'Decks' },
  { id: 'assistant', label: 'Assistant' },
  { id: 'settings', label: 'Settings' },
];

const placeholders: Record<string, string> = {
  home: 'MTGA presence, source health, last sync and snapshot totals will live here (task 4.3).',
  collection: 'Search, sorting, compound filters and snapshot comparison arrive with task 4.4.',
  decks: 'Structured deck editing, ownership and legality arrive with task 5.6.',
  assistant: 'The approval-gated assistant arrives with the assistant tasks.',
  settings: 'Source detection, Detailed Logs guidance and privacy choices arrive with task 4.2.',
};

const app = document.querySelector<HTMLDivElement>('#app')!;

function render(views: View[], active: string): void {
  const nav = views
    .map(
      (view) =>
        `<button type="button" data-view="${view.id}"` +
        ` aria-current="${view.id === active ? 'page' : 'false'}">${view.label}</button>`,
    )
    .join('');
  const current = views.find((view) => view.id === active);
  app.innerHTML =
    `<nav aria-label="Main navigation">${nav}</nav>` +
    `<main><h1>${current?.label ?? ''}</h1>` +
    `<p class="placeholder">${placeholders[active] ?? ''}</p></main>`;
  app.querySelectorAll<HTMLButtonElement>('button[data-view]').forEach((button) => {
    button.addEventListener('click', () => {
      location.hash = button.dataset.view!;
    });
  });
}

async function start(): Promise<void> {
  let views: View[];
  try {
    views = await Views();
  } catch {
    views = fallbackViews;
  }
  const sync = (): void => {
    const requested = location.hash.replace('#', '') || 'home';
    const active = views.some((view) => view.id === requested) ? requested : 'home';
    render(views, active);
  };
  window.addEventListener('hashchange', sync);
  sync();
}

void start();
