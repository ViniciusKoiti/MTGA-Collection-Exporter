// Package signing implementa a assinatura Ed25519 de manifestos de catálogo
// (tarefa 4.4 do OpenSpec add-central-go-platform). A chave privada chega ao
// processo por referência de segredo externa; este pacote nunca a registra
// em log nem a serializa.
package signing

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

// Signer assina manifestos com uma chave Ed25519 identificada por key ID.
// Durante uma rotação, publica-se com a chave nova enquanto os clientes
// ainda confiam na antiga (transição dual-trust no Verifier).
type Signer struct {
	keyID string
	key   ed25519.PrivateKey
}

// NewSigner valida o tamanho da chave e o key ID antes de aceitar.
func NewSigner(keyID string, key ed25519.PrivateKey) (*Signer, error) {
	if keyID == "" {
		return nil, fmt.Errorf("signing: key ID vazio")
	}
	if len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("signing: chave privada com tamanho inválido")
	}
	return &Signer{keyID: keyID, key: key}, nil
}

// Sign produz o manifesto assinado sobre o payload canônico do objeto.
func (s *Signer) Sign(ctx context.Context, ref catalog.ObjectRef) (catalog.Manifest, error) {
	if err := ctx.Err(); err != nil {
		return catalog.Manifest{}, err
	}
	if ref.Key == "" || ref.SHA256 == "" {
		return catalog.Manifest{}, fmt.Errorf("signing: referência de objeto incompleta")
	}
	assinatura := ed25519.Sign(s.key, canonicalPayload(ref))
	return catalog.Manifest{
		KeyID:     s.keyID,
		Signature: base64.StdEncoding.EncodeToString(assinatura),
		Object:    ref,
	}, nil
}

// canonicalPayload fixa o formato assinado; qualquer mudança aqui é quebra
// de compatibilidade e exige nova versão de schema de manifesto.
func canonicalPayload(ref catalog.ObjectRef) []byte {
	return fmt.Appendf(nil, "v1\n%s\n%s\n%d\n", ref.Key, ref.SHA256, ref.Size)
}
