// Decks workspace (task 5.6): structured editing over Arena text,
// legality verdict, ownership, revision history and export preview —
// the engine is Go (domain/decks + decksvc), this module renders it.
import { AnalyzeDeck, SaveDeckRevision, DeckHistory } from '../wailsjs/go/main/App';
import { renderState, type ViewState } from './state';

type Analysis = {
  preview: string;
  verdict: { Legal: boolean; CatalogoStale: boolean;
    Problemas?: { Code: string; Carta?: string; Detail: string }[] };
  ownership: { Lines?: { Name: string; Required: number; Owned: number;
    Missing: number }[]; TotalMissing: number; Complete: boolean };
  state: ViewState;
};

const editorHTML =
  `<div class="controls">` +
  `<input id="deck-name" type="text" aria-label="Deck name" placeholder="Deck name"/>` +
  `<button type="button" id="deck-analyze">Analyze</button>` +
  `<button type="button" id="deck-save">Save revision</button></div>` +
  `<textarea id="deck-text" aria-label="Deck list in Arena format" rows="10"` +
  ` placeholder="4 Lightning Strike (DMU) 123&#10;56 Mountain"></textarea>` +
  `<div id="deck-result"></div><div id="deck-history"></div>`;

function verdictHTML(a: Analysis): string {
  const stale = a.verdict.CatalogoStale
    ? `<div class="state state-stale"><strong>Legality unknown</strong>` +
      `<span class="state-detail">The legality catalog is not installed ` +
      `yet; card-legality problems below are pending it.</span></div>`
    : '';
  const problems = (a.verdict.Problemas ?? [])
    .map((p) => `<li><code class="state-code">${p.Code}</code> ` +
      `${p.Carta ?? ''} ${p.Detail}</li>`).join('');
  const badge = a.verdict.Legal ? 'legal' : 'not legal';
  return `${stale}<p>Verdict: <strong>${badge}</strong></p>` +
    (problems ? `<ul>${problems}</ul>` : '');
}

function ownershipHTML(a: Analysis): string {
  const missing = (a.ownership.Lines ?? []).filter((l) => l.Missing > 0);
  if (a.ownership.Complete) return '<p>You own every card in this deck.</p>';
  if (!missing.length) return '';
  return `<p>Missing ${a.ownership.TotalMissing} cards:</p>` +
    `<table><thead><tr><th>Card</th><th>Need</th><th>Own</th>` +
    `<th>Missing</th></tr></thead><tbody>` +
    missing.map((l) => `<tr><td>${l.Name}</td><td>${l.Required}</td>` +
      `<td>${l.Owned}</td><td>${l.Missing}</td></tr>`).join('') +
    `</tbody></table>`;
}

async function analyze(container: HTMLElement): Promise<void> {
  const name = container.querySelector<HTMLInputElement>('#deck-name')!.value;
  const text = container.querySelector<HTMLTextAreaElement>('#deck-text')!.value;
  const target = container.querySelector<HTMLDivElement>('#deck-result')!;
  const a = (await AnalyzeDeck(name, text)) as unknown as Analysis;
  if (a.state.status === 'error') {
    target.innerHTML = renderState(a.state);
    return;
  }
  target.innerHTML =
    renderState(a.state) + verdictHTML(a) + ownershipHTML(a) +
    `<h3>Arena export preview</h3><pre>${a.preview}</pre>`;
}

async function history(container: HTMLElement): Promise<void> {
  const name = container.querySelector<HTMLInputElement>('#deck-name')!.value;
  const target = container.querySelector<HTMLDivElement>('#deck-history')!;
  const revisions = ((await DeckHistory(name || 'Untitled')) ?? []) as
    { id: string; saved_at: string; ruleset: string }[];
  target.innerHTML = revisions.length
    ? `<h3>Revision history</h3><ul>` +
      revisions.map((r) =>
        `<li><code class="state-code">${r.id}</code> ${r.saved_at} · ${r.ruleset}</li>`)
        .join('') + `</ul>`
    : '';
}

export function renderDecks(container: HTMLElement): void {
  container.innerHTML = editorHTML;
  container.querySelector('#deck-analyze')!.addEventListener('click', () => {
    void analyze(container);
  });
  container.querySelector('#deck-save')!.addEventListener('click', () => {
    void (async () => {
      const name = container.querySelector<HTMLInputElement>('#deck-name')!.value;
      const text = container.querySelector<HTMLTextAreaElement>('#deck-text')!.value;
      await SaveDeckRevision(name, text);
      await analyze(container);
      await history(container);
    })();
  });
  void history(container);
}
