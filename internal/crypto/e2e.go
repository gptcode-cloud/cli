// Package crypto provides end-to-end encryption for agent-browser communication.
// Uses X25519 for key exchange and ChaCha20-Poly1305 for symmetric encryption.
package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

// E2ESession manages an encrypted session between agent and browser.
type E2ESession struct {
	privateKey   [32]byte
	publicKey    [32]byte
	remotePublic [32]byte
	sharedSecret [32]byte
	cipher       cipher.AEAD
	keyExchanged bool
}

// NewE2ESession creates a new E2E encryption session with a fresh key pair.
func NewE2ESession() (*E2ESession, error) {
	s := &E2ESession{}

	// Generate random private key
	if _, err := rand.Read(s.privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Derive public key from private key (X25519)
	pub, err := curve25519.X25519(s.privateKey[:], curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}
	copy(s.publicKey[:], pub)

	return s, nil
}

// PublicKey returns the session's public key as base64.
func (s *E2ESession) PublicKey() string {
	return base64.StdEncoding.EncodeToString(s.publicKey[:])
}

// PublicKeyBytes returns the raw public key bytes.
func (s *E2ESession) PublicKeyBytes() []byte {
	return s.publicKey[:]
}

// SetRemotePublicKey sets the remote party's public key and derives the shared secret.
func (s *E2ESession) SetRemotePublicKey(keyBase64 string) error {
	keyBytes, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return fmt.Errorf("invalid public key encoding: %w", err)
	}

	return s.SetRemotePublicKeyBytes(keyBytes)
}

// SetRemotePublicKeyBytes sets the remote public key from raw bytes.
func (s *E2ESession) SetRemotePublicKeyBytes(keyBytes []byte) error {
	if len(keyBytes) != 32 {
		return errors.New("public key must be 32 bytes")
	}

	copy(s.remotePublic[:], keyBytes)

	// Derive shared secret using X25519
	shared, err := curve25519.X25519(s.privateKey[:], s.remotePublic[:])
	if err != nil {
		return fmt.Errorf("X25519 key exchange failed: %w", err)
	}
	copy(s.sharedSecret[:], shared)

	// Create ChaCha20-Poly1305 cipher with shared secret
	cipher, err := chacha20poly1305.NewX(s.sharedSecret[:])
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}
	s.cipher = cipher
	s.keyExchanged = true

	return nil
}

// IsReady returns true if key exchange is complete and encryption is ready.
func (s *E2ESession) IsReady() bool {
	return s.keyExchanged && s.cipher != nil
}

// Encrypt encrypts plaintext and returns base64-encoded ciphertext.
func (s *E2ESession) Encrypt(plaintext []byte) (string, error) {
	if !s.IsReady() {
		return "", errors.New("key exchange not complete")
	}

	// Generate random nonce (24 bytes for XChaCha20)
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt with authenticated encryption
	ciphertext := s.cipher.Seal(nonce, nonce, plaintext, nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// EncryptBytes encrypts and returns raw bytes (nonce + ciphertext).
func (s *E2ESession) EncryptBytes(plaintext []byte) ([]byte, error) {
	if !s.IsReady() {
		return nil, errors.New("key exchange not complete")
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return s.cipher.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts base64-encoded ciphertext and returns plaintext.
func (s *E2ESession) Decrypt(ciphertextBase64 string) ([]byte, error) {
	if !s.IsReady() {
		return nil, errors.New("key exchange not complete")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid ciphertext encoding: %w", err)
	}

	return s.DecryptBytes(ciphertext)
}

// DecryptBytes decrypts raw bytes (nonce + ciphertext).
func (s *E2ESession) DecryptBytes(ciphertext []byte) ([]byte, error) {
	if !s.IsReady() {
		return nil, errors.New("key exchange not complete")
	}

	if len(ciphertext) < chacha20poly1305.NonceSizeX {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce from beginning
	nonce := ciphertext[:chacha20poly1305.NonceSizeX]
	ciphertext = ciphertext[chacha20poly1305.NonceSizeX:]

	// Decrypt and verify
	plaintext, err := s.cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// Fingerprint returns a short hash of the public key for verification.
func (s *E2ESession) Fingerprint() string {
	// First 8 bytes of public key in hex
	return fmt.Sprintf("%X", s.publicKey[:8])
}

// RemoteFingerprint returns the fingerprint of the remote public key.
func (s *E2ESession) RemoteFingerprint() string {
	if !s.keyExchanged {
		return ""
	}
	return fmt.Sprintf("%X", s.remotePublic[:8])
}
