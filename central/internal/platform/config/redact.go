package config

import "log/slog"

// redacted substitui qualquer valor sensível nos diagnósticos de inicialização.
const redacted = "[redigido]"

// Diagnostics devolve atributos estruturados seguros para log de inicialização.
// Referências de segredo aparecem apenas como presentes/ausentes: o valor da
// referência pode conter caminhos internos e nunca é emitido.
func (c Config) Diagnostics() []slog.Attr {
	return []slog.Attr{
		slog.String("env", string(c.Env)),
		slog.String("http_addr", c.HTTPAddr),
		slog.String("db_secret_ref", presence(c.DBSecretRef)),
		slog.String("db_dsn", presence(c.DBDsn)),
		slog.String("signer_key_ref", presence(c.SignerKeyRef)),
		slog.Int("fetch_workers", c.Budgets.FetchWorkers),
		slog.Int("normalize_workers", c.Budgets.NormalizeWorkers),
		slog.Int("provider_buffer", c.Budgets.ProviderBuffer),
		slog.Int("observation_buffer", c.Budgets.ObservationBuffer),
		slog.Int("normalized_buffer", c.Budgets.NormalizedBuffer),
		slog.Int("batch_size", c.Budgets.BatchSize),
	}
}

func presence(secretRef string) string {
	if secretRef == "" {
		return "(ausente)"
	}
	return redacted
}
