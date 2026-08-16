// First-run setup panel (task 4.2), rendered inside Settings. The
// flow itself lives in Go (application/setup); this module only
// renders its Result and re-runs it when a privacy choice flips.
import { FirstRun } from '../wailsjs/go/main/App';

type SetupResult = {
  source: string;
  needs_guidance: boolean;
  verified: boolean;
  fell_back: boolean;
  privacy: { telemetry_opt_in: boolean; assistant_enabled: boolean };
};

const sourceLabels: Record<string, string> = {
  json_import: 'Explicit JSON import',
  detailed_logs: 'MTGA Detailed Logs',
  legacy_bridge: 'Legacy Python bridge',
};

const guidanceHTML =
  '<div class="state state-stale"><strong>Detailed Logs guidance</strong>' +
  '<span class="state-detail">The current MTG Arena client does not expose ' +
  'your collection through Detailed Logs, so the explicit import is the ' +
  'reliable path. To keep logs enabled anyway: Options → Account → ' +
  'Detailed Logs (Plugin Support), then restart the game.</span></div>';

export async function renderFirstRun(container: HTMLElement,
  optIn = false, assistant = false): Promise<void> {
  container.innerHTML = '<p class="placeholder">Detecting collection source…</p>';
  let result: SetupResult;
  try {
    result = await FirstRun(optIn, assistant);
  } catch {
    container.innerHTML =
      '<p class="placeholder">First-run setup needs the desktop runtime.</p>';
    return;
  }
  const badge = result.verified
    ? '<span class="state-code">verified</span>'
    : '<span class="state-code">unverified</span>';
  const fellBack = result.fell_back
    ? '<div class="state state-partial"><strong>Fallback engaged</strong>' +
      '<span class="state-detail">Detailed Logs could not be verified; ' +
      'switched to the explicit import.</span></div>'
    : '';
  container.innerHTML =
    `<h2>First-run setup</h2>` +
    `<p>Collection source: <strong>${sourceLabels[result.source] ?? result.source}</strong> ${badge}</p>` +
    fellBack +
    (result.needs_guidance ? guidanceHTML : '') +
    `<h3>Privacy choices</h3>` +
    `<p class="placeholder">Nothing is shared unless you opt in.</p>` +
    `<label><input type="checkbox" id="opt-telemetry"` +
    ` ${result.privacy.telemetry_opt_in ? 'checked' : ''}/> Share anonymous usage telemetry</label><br/>` +
    `<label><input type="checkbox" id="opt-assistant"` +
    ` ${result.privacy.assistant_enabled ? 'checked' : ''}/> Enable the assistant</label>`;
  const telemetry = container.querySelector<HTMLInputElement>('#opt-telemetry')!;
  const assistantBox = container.querySelector<HTMLInputElement>('#opt-assistant')!;
  const rerun = (): void => {
    void renderFirstRun(container, telemetry.checked, assistantBox.checked);
  };
  telemetry.addEventListener('change', rerun);
  assistantBox.addEventListener('change', rerun);
}
