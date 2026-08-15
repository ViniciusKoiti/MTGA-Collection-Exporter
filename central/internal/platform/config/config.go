// Package config faz o parsing estrito da configuração do serviço central.
//
// Regras (OpenSpec add-central-go-platform, tarefa 1.5):
//   - toda variável reconhecida tem prefixo CENTRAL_;
//   - variáveis CENTRAL_* desconhecidas são rejeitadas (parsing estrito);
//   - segredos entram apenas como referência externa e nunca em texto claro
//     nos diagnósticos;
//   - todos os limites de concorrência têm defaults seguros e validação.
package config

import (
	"fmt"
	"strings"
)

// Environment identifica o ambiente de execução do processo.
type Environment string

// Ambientes válidos; produção nunca é o default.
const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// Budgets concentra os limites de concorrência derivados da configuração.
// Nenhum código de runtime pode escolher concorrência fora destes valores.
type Budgets struct {
	FetchWorkers      int // pool de fetch de provedores (orçamento de I/O)
	NormalizeWorkers  int // pool de normalização (orçamento de CPU)
	ProviderBuffer    int // capacidade da fila de jobs de provedor (P)
	ObservationBuffer int // capacidade da fila de observações (O)
	NormalizedBuffer  int // capacidade da fila de cartas normalizadas (N)
	BatchSize         int // tamanho máximo de lote de escrita
}

// Config é a configuração validada de um processo central (api, worker, migrate).
type Config struct {
	Env          Environment
	HTTPAddr     string
	DBSecretRef  string // referência a segredo externo; nunca o valor em si
	SignerKeyRef string // referência à chave de assinatura; nunca a chave
	Budgets      Budgets
}

// Load valida o ambiente informado e devolve a configuração ou o primeiro erro.
// `environ` segue o formato de os.Environ (CHAVE=valor), o que mantém o
// parsing puro e testável sem tocar o processo.
func Load(environ []string) (Config, error) {
	vars, err := centralVars(environ)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		HTTPAddr:     take(vars, "CENTRAL_HTTP_ADDR", "127.0.0.1:8080"),
		DBSecretRef:  take(vars, "CENTRAL_DB_SECRET_REF", ""),
		SignerKeyRef: take(vars, "CENTRAL_SIGNER_KEY_REF", ""),
	}
	env := take(vars, "CENTRAL_ENV", string(EnvDevelopment))
	switch Environment(env) {
	case EnvDevelopment, EnvStaging, EnvProduction:
		cfg.Env = Environment(env)
	default:
		return Config{}, fmt.Errorf("config: CENTRAL_ENV inválido: %q", env)
	}
	if cfg.Budgets, err = loadBudgets(vars); err != nil {
		return Config{}, err
	}
	for name := range vars {
		return Config{}, fmt.Errorf("config: variável desconhecida: %s", name)
	}
	return cfg, nil
}

// centralVars separa as variáveis CENTRAL_* preservando apenas o escopo do serviço.
func centralVars(environ []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, kv := range environ {
		name, value, ok := strings.Cut(kv, "=")
		if ok && strings.HasPrefix(name, "CENTRAL_") {
			vars[name] = value
		}
	}
	return vars, nil
}

// take remove a chave do mapa para que sobras denunciem variáveis desconhecidas.
func take(vars map[string]string, name, fallback string) string {
	value, ok := vars[name]
	delete(vars, name)
	if !ok || value == "" {
		return fallback
	}
	return value
}
