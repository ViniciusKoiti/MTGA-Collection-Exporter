package devmcp

import "testing"

func TestClassificacaoNegaProducaoEAmbiguidade(t *testing.T) {
	casos := map[string]struct {
		alvo   Alvo
		classe Ambiente
	}{
		"dev local completo": {
			Alvo{Host: "localhost", DSN: "file:dev.db", Marcador: "development"}, AmbienteDev},
		"dev por loopback": {
			Alvo{Host: "127.0.0.1:8080", Marcador: "development"}, AmbienteDev},
		"staging declarado e evidente": {
			Alvo{Host: "api.staging.exemplo.dev", Marcador: "staging"}, AmbienteStaging},
		"dsn de producao": {
			Alvo{Host: "localhost", DSN: "postgres://prod-db/catalog", Marcador: "development"},
			AmbienteProducao},
		"host de producao": {
			Alvo{Host: "api.live.exemplo.com", Marcador: "staging"}, AmbienteProducao},
		"certificado de producao": {
			Alvo{Host: "localhost", CertCN: "cn=release.exemplo.com", Marcador: "development"},
			AmbienteProducao},
		"sem marcador": {
			Alvo{Host: "localhost"}, AmbienteAmbiguo},
		"marcador dev com host remoto": {
			Alvo{Host: "10.0.5.7", Marcador: "development"}, AmbienteAmbiguo},
		"marcador staging sem evidencia": {
			Alvo{Host: "api.exemplo.dev", Marcador: "staging"}, AmbienteAmbiguo},
	}
	for nome, caso := range casos {
		if got := Classifica(caso.alvo); got != caso.classe {
			t.Errorf("%s: esperava %s, veio %s", nome, caso.classe, got)
		}
	}
}

func TestAutorizacaoLiberaApenasDevEStaging(t *testing.T) {
	if err := AutorizaConexao(Alvo{Host: "localhost", Marcador: "development"}); err != nil {
		t.Fatalf("dev local deveria conectar: %v", err)
	}
	negados := []Alvo{
		{Host: "localhost", DSN: "postgres://prod/x", Marcador: "development"},
		{Host: "localhost"}, // ambíguo
	}
	for _, alvo := range negados {
		if err := AutorizaConexao(alvo); err == nil {
			t.Errorf("alvo deveria ser negado: %+v", alvo)
		}
	}
}
