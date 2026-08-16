// Package catalogclient verifies and caches signed catalog manifests
// on the desktop side (central OpenSpec task 4.7, desktop half). It
// speaks the versioned wire contract only — the companion never
// imports a central package, by architecture gate.
package catalogclient

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// ObjectRef mirrors the manifest object reference on the wire.
type ObjectRef struct {
	Key    string `json:"key"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Manifest mirrors the signed manifest wire shape.
type Manifest struct {
	KeyID     string    `json:"key_id"`
	Signature string    `json:"signature"`
	Schema    string    `json:"schema"`
	Object    ObjectRef `json:"object"`
}

// canonicalPayload pins the signed format of manifest schema v1;
// central changing it is a breaking change that must bump the schema.
func canonicalPayload(ref ObjectRef) []byte {
	return fmt.Appendf(nil, "v1\n%s\n%s\n%d\n", ref.Key, ref.SHA256, ref.Size)
}

// Verifier holds the trusted public keys by key ID; removing a key
// revokes every manifest signed by it, immediately.
type Verifier struct {
	Trusted map[string]ed25519.PublicKey
}

// Verify accepts a manifest only when the key ID is trusted, the
// signature covers the canonical payload, the object bytes match the
// declared hash and size, and the schema is the expected one.
func (v Verifier) Verify(man Manifest, object []byte, wantSchema string) error {
	key, ok := v.Trusted[man.KeyID]
	if !ok {
		return fmt.Errorf("catalogclient: unknown key id %q", man.KeyID)
	}
	signature, err := base64.StdEncoding.DecodeString(man.Signature)
	if err != nil {
		return fmt.Errorf("catalogclient: malformed signature: %w", err)
	}
	if !ed25519.Verify(key, canonicalPayload(man.Object), signature) {
		return fmt.Errorf("catalogclient: invalid signature for %q", man.KeyID)
	}
	sum := sha256.Sum256(object)
	if hex.EncodeToString(sum[:]) != man.Object.SHA256 ||
		int64(len(object)) != man.Object.Size {
		return fmt.Errorf("catalogclient: object does not match its manifest")
	}
	if man.Schema != wantSchema {
		return fmt.Errorf("catalogclient: schema %q is not compatible with %q",
			man.Schema, wantSchema)
	}
	return nil
}
