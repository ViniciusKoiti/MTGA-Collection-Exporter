import './style.css';
import { Views } from '../wailsjs/go/main/App';
import { renderState, type ViewState } from './state';
import { renderFirstRun } from './firstrun';
import { renderHome } from './home';
import { renderCollection } from './collection';

type View = { id: string; label: string };

// Every view starts empty; real states flow from the application
// services as their tasks land (4.3/4.4/5.x).
const initialStates: Record<string, ViewState> = {
  home: { status: 'success' }, // Home renders its own state banner
  collection: { status: 'empty' },
  decks: { status: 'empty' },
  assistant: { status: 'empty' },
  settings: { status: 'empty' },
};

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
    renderState(initialStates[active] ?? { status: 'empty' }) +
    `<p class="placeholder">${placeholders[active] ?? ''}</p>` +
    `<div id="view-body"></div></main>`;
  const body = app.querySelector<HTMLDivElement>('#view-body')!;
  if (active === 'settings') {
    void renderFirstRun(body);
  } else if (active === 'home') {
    void renderHome(body);
  } else if (active === 'collection') {
    renderCollection(body);
  }
  app.querySelectorAll<HTMLButtonElement>('button[data-view]').forEach((button) => {
    button.addEventListener('click', () => {
      location.hash = button.dataset.view!;
    });
  });
  // Keyboard navigation (task 4.6): arrows move through the nav,
  // Home/End jump, Enter/Space activate (native button behavior).
  const navElement = app.querySelector<HTMLElement>('nav')!;
  navElement.addEventListener('keydown', (event) => {
    const buttons = Array.from(
      navElement.querySelectorAll<HTMLButtonElement>('button[data-view]'));
    const index = buttons.indexOf(document.activeElement as HTMLButtonElement);
    if (index < 0) return;
    const move = (to: number): void => {
      event.preventDefault();
      buttons[(to + buttons.length) % buttons.length].focus();
    };
    if (event.key === 'ArrowDown' || event.key === 'ArrowRight') move(index + 1);
    else if (event.key === 'ArrowUp' || event.key === 'ArrowLeft') move(index - 1);
    else if (event.key === 'Home') move(0);
    else if (event.key === 'End') move(buttons.length - 1);
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
