package config

import (
	"fmt"
	"strconv"
)

// Defaults seguros: valores pequenos que nunca excedem orçamentos de banco ou
// de provedor por acidente. Ajustes maiores exigem configuração explícita.
const (
	defaultFetchWorkers      = 4
	defaultNormalizeWorkers  = 4
	defaultProviderBuffer    = 16
	defaultObservationBuffer = 64
	defaultNormalizedBuffer  = 64
	defaultBatchSize         = 500
	maxWorkers               = 64
	maxBuffer                = 4096
	maxBatchSize             = 5000
)

// loadBudgets valida cada limite de concorrência; consome as chaves lidas.
func loadBudgets(vars map[string]string) (Budgets, error) {
	b := Budgets{}
	fields := []struct {
		dst      *int
		name     string
		fallback int
		max      int
	}{
		{&b.FetchWorkers, "CENTRAL_FETCH_WORKERS", defaultFetchWorkers, maxWorkers},
		{&b.NormalizeWorkers, "CENTRAL_NORMALIZE_WORKERS", defaultNormalizeWorkers, maxWorkers},
		{&b.ProviderBuffer, "CENTRAL_PROVIDER_BUFFER", defaultProviderBuffer, maxBuffer},
		{&b.ObservationBuffer, "CENTRAL_OBSERVATION_BUFFER", defaultObservationBuffer, maxBuffer},
		{&b.NormalizedBuffer, "CENTRAL_NORMALIZED_BUFFER", defaultNormalizedBuffer, maxBuffer},
		{&b.BatchSize, "CENTRAL_BATCH_SIZE", defaultBatchSize, maxBatchSize},
	}
	for _, f := range fields {
		value, err := positiveInt(vars, f.name, f.fallback, f.max)
		if err != nil {
			return Budgets{}, err
		}
		*f.dst = value
	}
	return b, nil
}

// positiveInt exige inteiro em [1, max]; zero ou negativo nunca é aceito
// porque desligaria silenciosamente um estágio do pipeline.
func positiveInt(vars map[string]string, name string, fallback, max int) (int, error) {
	raw := take(vars, name, "")
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("config: %s não é inteiro: %q", name, raw)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("config: %s fora de [1, %d]: %d", name, max, value)
	}
	return value, nil
}
