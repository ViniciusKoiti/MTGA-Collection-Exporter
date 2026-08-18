package config

import (
	"strings"
	"testing"
)

func TestLoadDefaultsSeguras(t *testing.T) {
	cfg, err := Load([]string{"PATH=/x", "CENTRAL_ENV=development"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.Env != EnvDevelopment || cfg.HTTPAddr != "127.0.0.1:8080" {
		t.Fatalf("defaults incorretos: %+v", cfg)
	}
	if cfg.Budgets.FetchWorkers < 1 || cfg.Budgets.BatchSize < 1 {
		t.Fatalf("orçamentos devem ser positivos: %+v", cfg.Budgets)
	}
}

func TestLoadRejeitaEntradasInvalidas(t *testing.T) {
	casos := map[string][]string{
		"ambiente inválido":     {"CENTRAL_ENV=prod"},
		"variável desconhecida": {"CENTRAL_TYPO=1"},
		"inteiro inválido":      {"CENTRAL_FETCH_WORKERS=muitos"},
		"worker zero":           {"CENTRAL_FETCH_WORKERS=0"},
		"buffer acima do teto":  {"CENTRAL_PROVIDER_BUFFER=999999"},
	}
	for nome, environ := range casos {
		if _, err := Load(environ); err == nil {
			t.Errorf("%s: Load deveria falhar para %v", nome, environ)
		}
	}
}

func TestLoadAceitaOverridesValidos(t *testing.T) {
	cfg, err := Load([]string{
		"CENTRAL_ENV=staging",
		"CENTRAL_FETCH_WORKERS=8",
		"CENTRAL_BATCH_SIZE=100",
		"CENTRAL_DB_SECRET_REF=vault://central/db",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.Env != EnvStaging || cfg.Budgets.FetchWorkers != 8 || cfg.Budgets.BatchSize != 100 {
		t.Fatalf("overrides não aplicados: %+v", cfg)
	}
}

func TestDiagnosticsRedigemSegredos(t *testing.T) {
	cfg, err := Load([]string{
		"CENTRAL_ENV=production",
		"CENTRAL_DB_SECRET_REF=vault://central/db#token",
		"CENTRAL_SIGNER_KEY_REF=kms://key/ed25519",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	for _, attr := range cfg.Diagnostics() {
		if strings.Contains(attr.Value.String(), "vault://") ||
			strings.Contains(attr.Value.String(), "kms://") {
			t.Fatalf("diagnóstico vazou segredo: %s=%s", attr.Key, attr.Value)
		}
	}
}
