package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

func par(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("geração de chave: %v", err)
	}
	return pub, priv
}

func ref() catalog.ObjectRef {
	return catalog.ObjectRef{Key: "catalog/abc.json.gz", SHA256: "deadbeef", Size: 42}
}

func TestAssinaturaValidaEhAceita(t *testing.T) {
	pub, priv := par(t)
	signer, _ := NewSigner("k1", priv)
	man, err := signer.Sign(t.Context(), ref())
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	verifier, _ := NewVerifier(map[string]ed25519.PublicKey{"k1": pub})
	if err := verifier.Verify(man); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestKeyIDDesconhecidoEhRejeitado(t *testing.T) {
	_, priv := par(t)
	outraPub, _ := par(t)
	signer, _ := NewSigner("k1", priv)
	man, _ := signer.Sign(t.Context(), ref())
	verifier, _ := NewVerifier(map[string]ed25519.PublicKey{"k2": outraPub})
	if err := verifier.Verify(man); err == nil {
		t.Fatal("manifesto com key ID fora do conjunto confiável deveria falhar")
	}
}

func TestRotacaoDualTrustERevogacao(t *testing.T) {
	pubAntiga, privAntiga := par(t)
	pubNova, privNova := par(t)
	antiga, _ := NewSigner("k-old", privAntiga)
	nova, _ := NewSigner("k-new", privNova)
	manAntigo, _ := antiga.Sign(t.Context(), ref())
	manNovo, _ := nova.Sign(t.Context(), ref())

	transicao, _ := NewVerifier(map[string]ed25519.PublicKey{
		"k-old": pubAntiga, "k-new": pubNova,
	})
	if transicao.Verify(manAntigo) != nil || transicao.Verify(manNovo) != nil {
		t.Fatal("durante a transição, as duas chaves devem ser aceitas")
	}
	posRevogacao, _ := NewVerifier(map[string]ed25519.PublicKey{"k-new": pubNova})
	if posRevogacao.Verify(manAntigo) == nil {
		t.Fatal("após revogação, manifesto da chave antiga deve ser rejeitado")
	}
	if posRevogacao.Verify(manNovo) != nil {
		t.Fatal("chave nova segue válida após a revogação da antiga")
	}
}

func TestAssinaturaAdulteradaEhRejeitada(t *testing.T) {
	pub, priv := par(t)
	signer, _ := NewSigner("k1", priv)
	man, _ := signer.Sign(t.Context(), ref())
	man.Object.SHA256 = "0000beef" // objeto trocado após a assinatura
	verifier, _ := NewVerifier(map[string]ed25519.PublicKey{"k1": pub})
	if err := verifier.Verify(man); err == nil {
		t.Fatal("hash adulterado deveria invalidar a assinatura")
	}
}
