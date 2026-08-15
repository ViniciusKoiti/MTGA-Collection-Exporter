package signing

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/ViniciusKoiti/MTGA-Collection-Exporter/central/internal/domain/catalog"
)

// Verifier valida manifestos contra um conjunto de chaves públicas confiáveis
// indexadas por key ID. É o mesmo contrato que os clientes desktop aplicam:
// manifesto com key ID desconhecido é rejeitado e o snapshot anterior
// permanece válido (cenário "Signature is unknown" da spec).
type Verifier struct {
	confiaveis map[string]ed25519.PublicKey
}

// NewVerifier copia o conjunto confiável; remoção de uma chave do conjunto
// (revogação) invalida imediatamente os manifestos assinados por ela.
func NewVerifier(confiaveis map[string]ed25519.PublicKey) (*Verifier, error) {
	if len(confiaveis) == 0 {
		return nil, fmt.Errorf("signing: conjunto de chaves confiáveis vazio")
	}
	copia := make(map[string]ed25519.PublicKey, len(confiaveis))
	for id, chave := range confiaveis {
		if len(chave) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("signing: chave pública %q com tamanho inválido", id)
		}
		copia[id] = chave
	}
	return &Verifier{confiaveis: copia}, nil
}

// Verify aceita o manifesto somente se o key ID for confiável e a assinatura
// cobrir exatamente o payload canônico do objeto referenciado.
func (v *Verifier) Verify(man catalog.Manifest) error {
	chave, ok := v.confiaveis[man.KeyID]
	if !ok {
		return fmt.Errorf("signing: key ID desconhecido: %q", man.KeyID)
	}
	assinatura, err := base64.StdEncoding.DecodeString(man.Signature)
	if err != nil {
		return fmt.Errorf("signing: assinatura malformada: %w", err)
	}
	if !ed25519.Verify(chave, canonicalPayload(man.Object), assinatura) {
		return fmt.Errorf("signing: assinatura inválida para key ID %q", man.KeyID)
	}
	return nil
}
