// Consistent view-state rendering (task 4.5). The vocabulary and the
// stable error codes are owned by Go (internal/application/viewstate
// and apperr); this module only renders them — one markup shape per
// status, everywhere.

export type ViewState = {
  status: 'empty' | 'loading' | 'stale' | 'partial' | 'error' | 'success';
  code?: string;
  detail?: string;
};

const labels: Record<ViewState['status'], string> = {
  empty: 'Nothing here yet',
  loading: 'Loading…',
  stale: 'Showing an older snapshot',
  partial: 'Partial data',
  error: 'Something went wrong',
  success: '',
};

export function renderState(state: ViewState): string {
  if (state.status === 'success') {
    return '';
  }
  const code = state.code
    ? `<code class="state-code">${state.code}</code>`
    : '';
  const detail = state.detail
    ? `<span class="state-detail">${state.detail}</span>`
    : '';
  return (
    `<div class="state state-${state.status}" role="status">` +
    `<strong>${labels[state.status]}</strong> ${code} ${detail}</div>`
  );
}
