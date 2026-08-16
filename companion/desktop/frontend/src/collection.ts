// Collection view (task 4.4): search, sorting, compound filters,
// per-card details, unresolved records and snapshot comparison — the
// engine lives in Go (collectionsvc), this module renders it.
import { Collection, CompareSnapshots } from '../wailsjs/go/main/App';
import { renderState, type ViewState } from './state';

type Row = {
  printing: string; name: string; set: string; quantity: number;
  unresolved: boolean; raw?: string; oracle?: string; arena?: number;
};
type PageResult = { rows?: Row[]; state: ViewState };
type DiffResult = {
  diff: { total_before: number; total_after: number;
    changes?: { name: string; set: string; before: number; after: number }[] };
  state: ViewState;
};

const controlsHTML =
  `<div class="controls">` +
  `<input id="col-text" type="search" placeholder="Search name or raw text"/>` +
  `<input id="col-set" type="text" placeholder="Set" size="5"/>` +
  `<select id="col-sort"><option value="name">Name</option>` +
  `<option value="set">Set</option><option value="quantity">Quantity</option></select>` +
  `<label><input id="col-unresolved" type="checkbox"/> Unresolved only</label>` +
  `<button type="button" id="col-compare">Compare snapshots</button></div>`;

function rowHTML(row: Row): string {
  const badge = row.unresolved ? ' <span class="state-code">unresolved</span>' : '';
  const details = row.unresolved
    ? `raw: ${row.raw ?? ''}`
    : `printing ${row.printing} · oracle ${row.oracle ?? '—'} · arena ${row.arena ?? '—'}`;
  return `<tr><td>${row.name || row.raw || '?'}${badge}</td>` +
    `<td>${row.set || '—'}</td><td>${row.quantity}</td>` +
    `<td class="detail">${details}</td></tr>`;
}

async function refresh(container: HTMLElement): Promise<void> {
  const text = container.querySelector<HTMLInputElement>('#col-text')!.value;
  const set = container.querySelector<HTMLInputElement>('#col-set')!.value;
  const sortBy = container.querySelector<HTMLSelectElement>('#col-sort')!.value;
  const unresolved =
    container.querySelector<HTMLInputElement>('#col-unresolved')!.checked;
  const result = (await Collection({
    text, set, unresolved_only: unresolved, min_quantity: 0, sort: sortBy,
  } as never)) as unknown as PageResult;
  const target = container.querySelector<HTMLDivElement>('#col-results')!;
  const rows = result.rows ?? [];
  target.innerHTML =
    renderState(result.state) +
    (rows.length
      ? `<table><thead><tr><th>Card</th><th>Set</th><th>Qty</th>` +
        `<th>Details</th></tr></thead><tbody>` +
        rows.map(rowHTML).join('') + `</tbody></table>`
      : '');
}

async function compare(container: HTMLElement): Promise<void> {
  const target = container.querySelector<HTMLDivElement>('#col-results')!;
  const result = (await CompareSnapshots()) as unknown as DiffResult;
  if (result.state.status !== 'success') {
    target.innerHTML = renderState(result.state) +
      '<p class="placeholder">Nothing to compare yet — import twice first.</p>';
    return;
  }
  const changes = result.diff.changes ?? [];
  target.innerHTML =
    `<p>Total ${result.diff.total_before} → ${result.diff.total_after}</p>` +
    (changes.length
      ? `<table><thead><tr><th>Card</th><th>Set</th><th>Before</th>` +
        `<th>After</th></tr></thead><tbody>` +
        changes.map((c) => `<tr><td>${c.name || '?'}</td><td>${c.set || '—'}</td>` +
          `<td>${c.before}</td><td>${c.after}</td></tr>`).join('') +
        `</tbody></table>`
      : '<p class="placeholder">No card counts changed.</p>');
}

export function renderCollection(container: HTMLElement): void {
  container.innerHTML = controlsHTML + '<div id="col-results"></div>';
  const rerun = (): void => { void refresh(container); };
  container.querySelector('#col-text')!.addEventListener('input', rerun);
  container.querySelector('#col-set')!.addEventListener('input', rerun);
  container.querySelector('#col-sort')!.addEventListener('change', rerun);
  container.querySelector('#col-unresolved')!.addEventListener('change', rerun);
  container.querySelector('#col-compare')!.addEventListener('click', () => {
    void compare(container);
  });
  rerun();
}
