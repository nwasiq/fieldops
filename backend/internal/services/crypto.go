package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

// ciphertextPrefix marks a stored value as encrypted so a value can never be
// encrypted twice or mistaken for plaintext.
const ciphertextPrefix = "enc:v1:"

// Cipher encrypts sensitive columns at rest with AES-256-GCM (§1.3).
type Cipher struct {
	aead cipher.AEAD
}

// NewCipher builds a Cipher from a base64-encoded 32-byte key.
func NewCipher(base64Key string) (*Cipher, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64Key))
	if err != nil {
		return nil, fmt.Errorf("decode encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("build cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("build gcm: %w", err)
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt returns the prefixed, base64 ciphertext for plaintext. An empty
// plaintext stays empty so "no value" is stored as NULL-equivalent, not as a
// ciphertext of nothing. Already-encrypted input is returned unchanged.
func (c *Cipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" || strings.HasPrefix(plaintext, ciphertextPrefix) {
		return plaintext, nil
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	sealed := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertextPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt. A value without the prefix is not ours and is an
// error: a plaintext licence number in the column means encryption was
// bypassed somewhere, and that must surface rather than be papered over.
func (c *Cipher) Decrypt(stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	if !strings.HasPrefix(stored, ciphertextPrefix) {
		return "", errors.New("stored value is not encrypted")
	}
	sealed, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, ciphertextPrefix))
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	nonceSize := c.aead.NonceSize()
	if len(sealed) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	plaintext, err := c.aead.Open(nil, sealed[:nonceSize], sealed[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}

// IsEncrypted reports whether a stored value carries the ciphertext prefix.
func IsEncrypted(stored string) bool {
	return strings.HasPrefix(stored, ciphertextPrefix)
}

// encryptUserSensitive encrypts the sensitive user columns in place. Call it
// before every save.
func encryptUserSensitive(c *Cipher, u *models.User) error {
	encrypted, err := c.Encrypt(u.LicenceNumber)
	if err != nil {
		return fmt.Errorf("encrypt licence number: %w", err)
	}
	u.LicenceNumber = encrypted
	return nil
}

// decryptUserSensitive decrypts the sensitive user columns in place. Call it
// after every read that needs them in the clear.
func decryptUserSensitive(c *Cipher, u *models.User) error {
	plaintext, err := c.Decrypt(u.LicenceNumber)
	if err != nil {
		return fmt.Errorf("decrypt licence number for user %d: %w", u.ID, err)
	}
	u.LicenceNumber = plaintext
	return nil
}

// fingerprint returns a short, stable digest of a sensitive value so the audit
// trail can show that it changed without recording it in the clear.
func fingerprint(value string) string {
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:6])
}
