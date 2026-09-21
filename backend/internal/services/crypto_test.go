package services

import (
	"strings"
	"testing"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

const testKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

// rule: §1.3 — licence numbers round-trip through the cipher and never rest in the clear
func TestCipher_RoundTrip(t *testing.T) {
	c, err := NewCipher(testKey)
	if err != nil {
		t.Fatal(err)
	}
	user := models.User{ID: 7, LicenceNumber: "GS-284611"}
	if err := encryptUserSensitive(c, &user); err != nil {
		t.Fatal(err)
	}
	if !IsEncrypted(user.LicenceNumber) || strings.Contains(user.LicenceNumber, "GS-284611") {
		t.Fatalf("stored value %q is not ciphertext", user.LicenceNumber)
	}
	stored := user.LicenceNumber
	if err := encryptUserSensitive(c, &user); err != nil {
		t.Fatal(err)
	}
	if user.LicenceNumber != stored {
		t.Error("encrypting twice changed an already-encrypted value")
	}
	if err := decryptUserSensitive(c, &user); err != nil {
		t.Fatal(err)
	}
	if user.LicenceNumber != "GS-284611" {
		t.Errorf("decrypted %q, want GS-284611", user.LicenceNumber)
	}
}

func TestCipher_EmptyAndPlaintext(t *testing.T) {
	c, err := NewCipher(testKey)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := c.Encrypt(""); got != "" {
		t.Errorf("empty plaintext encrypted to %q", got)
	}
	if got, _ := c.Decrypt(""); got != "" {
		t.Errorf("empty ciphertext decrypted to %q", got)
	}
	if _, err := c.Decrypt("GS-284611"); err == nil {
		t.Error("plaintext in the column decrypted without error")
	}
}

func TestNewCipher_RejectsBadKeys(t *testing.T) {
	for _, bad := range []string{"", "not-base64!", "c2hvcnQ="} {
		if _, err := NewCipher(bad); err == nil {
			t.Errorf("NewCipher(%q) accepted, want error", bad)
		}
	}
}

func TestFingerprint(t *testing.T) {
	if fingerprint("") != "" {
		t.Error("empty value has a fingerprint")
	}
	a, b := fingerprint("GS-1"), fingerprint("GS-2")
	if a == b || !strings.HasPrefix(a, "sha256:") || strings.Contains(a, "GS-1") {
		t.Errorf("fingerprints %q / %q are not distinct opaque digests", a, b)
	}
}
