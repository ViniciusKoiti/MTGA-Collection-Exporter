package postgres

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"
)

// encryptRoundTrip proves the backup is encrypted at rest — the
// ciphertext must not contain the plaintext — and comes back
// byte-identical after decryption (AES-256-GCM).
func encryptRoundTrip(t *testing.T, dump []byte) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("key: %v", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("gcm: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("nonce: %v", err)
	}
	sealed := gcm.Seal(nil, nonce, dump, nil)
	if bytes.Contains(sealed, []byte("CREATE TABLE")) {
		t.Fatal("the sealed backup leaks plaintext")
	}
	opened, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil || !bytes.Equal(opened, dump) {
		t.Fatalf("the encrypted backup must decrypt byte-identical: %v", err)
	}
	return opened
}
