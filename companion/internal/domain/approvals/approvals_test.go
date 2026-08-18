package approvals

import (
	"testing"
	"time"
)

func TestTokenDeUsoUnicoVinculadoAosArgumentos(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	token, err := NewToken("t1", "export-deck", "hash-exato", base.Add(time.Minute))
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if err := token.Usavel(base, "export-deck", "hash-exato"); err != nil {
		t.Fatalf("token vigente deveria autorizar: %v", err)
	}
	if err := token.Usavel(base, "export-deck", "outro-hash"); err == nil {
		t.Fatal("argumentos diferentes não podem reutilizar a aprovação")
	}
	if err := token.Usavel(base, "sync-collection", "hash-exato"); err == nil {
		t.Fatal("outra ferramenta não pode reutilizar a aprovação")
	}
	if err := token.Usavel(base.Add(2*time.Minute), "export-deck", "hash-exato"); err == nil {
		t.Fatal("token expirado não autoriza")
	}
	usado := token.Consumir(base)
	if err := usado.Usavel(base, "export-deck", "hash-exato"); err == nil {
		t.Fatal("token consumido é de uso único")
	}
	if err := token.Usavel(base, "export-deck", "hash-exato"); err != nil {
		t.Fatal("Consumir deve ser imutável sobre o original")
	}
}

func TestNewTokenRejeitaIncompletos(t *testing.T) {
	prazo := time.Unix(1_700_000_000, 0)
	for nome, caso := range map[string][4]string{
		"sem id":         {"", "tool", "hash", ""},
		"sem ferramenta": {"t1", "", "hash", ""},
		"sem hash":       {"t1", "tool", "", ""},
	} {
		if _, err := NewToken(caso[0], caso[1], caso[2], prazo); err == nil {
			t.Errorf("%s: deveria falhar", nome)
		}
	}
	if _, err := NewToken("t1", "tool", "hash", time.Time{}); err == nil {
		t.Error("token sem expiração deveria falhar")
	}
}
