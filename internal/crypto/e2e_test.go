package crypto

import (
	"bytes"
	"testing"
)

func TestNewE2ESession(t *testing.T) {
	session, err := NewE2ESession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session == nil {
		t.Fatal("Session is nil")
	}

	// Public key should be non-zero
	allZero := true
	for _, b := range session.publicKey {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("Public key is all zeros")
	}
}

func TestPublicKey(t *testing.T) {
	session, _ := NewE2ESession()

	pubKey := session.PublicKey()
	if len(pubKey) == 0 {
		t.Error("Public key string is empty")
	}

	pubKeyBytes := session.PublicKeyBytes()
	if len(pubKeyBytes) != 32 {
		t.Errorf("Public key bytes should be 32, got %d", len(pubKeyBytes))
	}
}

func TestKeyExchange(t *testing.T) {
	// Create two sessions (simulating agent and browser)
	alice, err := NewE2ESession()
	if err != nil {
		t.Fatalf("Failed to create Alice's session: %v", err)
	}

	bob, err := NewE2ESession()
	if err != nil {
		t.Fatalf("Failed to create Bob's session: %v", err)
	}

	// Exchange public keys
	if err := alice.SetRemotePublicKey(bob.PublicKey()); err != nil {
		t.Fatalf("Alice failed to set Bob's key: %v", err)
	}

	if err := bob.SetRemotePublicKey(alice.PublicKey()); err != nil {
		t.Fatalf("Bob failed to set Alice's key: %v", err)
	}

	// Both should be ready
	if !alice.IsReady() {
		t.Error("Alice should be ready")
	}
	if !bob.IsReady() {
		t.Error("Bob should be ready")
	}

	// Shared secrets should be equal
	if alice.sharedSecret != bob.sharedSecret {
		t.Error("Shared secrets don't match")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	alice, _ := NewE2ESession()
	bob, _ := NewE2ESession()

	alice.SetRemotePublicKey(bob.PublicKey())
	bob.SetRemotePublicKey(alice.PublicKey())

	testCases := []struct {
		name    string
		message []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello")},
		{"medium", []byte("this is a longer message for testing encryption")},
		{"with unicode", []byte("Hello 世界 🔐")},
		{"large", bytes.Repeat([]byte("A"), 10000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Alice encrypts
			ciphertext, err := alice.Encrypt(tc.message)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Bob decrypts
			plaintext, err := bob.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if !bytes.Equal(plaintext, tc.message) {
				t.Errorf("Message mismatch: got %q, want %q", plaintext, tc.message)
			}
		})
	}
}

func TestEncryptDecryptBytes(t *testing.T) {
	alice, _ := NewE2ESession()
	bob, _ := NewE2ESession()

	alice.SetRemotePublicKey(bob.PublicKey())
	bob.SetRemotePublicKey(alice.PublicKey())

	message := []byte("test message for bytes API")

	ciphertext, err := alice.EncryptBytes(message)
	if err != nil {
		t.Fatalf("EncryptBytes failed: %v", err)
	}

	plaintext, err := bob.DecryptBytes(ciphertext)
	if err != nil {
		t.Fatalf("DecryptBytes failed: %v", err)
	}

	if !bytes.Equal(plaintext, message) {
		t.Errorf("Message mismatch: got %q, want %q", plaintext, message)
	}
}

func TestBidirectionalEncryption(t *testing.T) {
	alice, _ := NewE2ESession()
	bob, _ := NewE2ESession()

	alice.SetRemotePublicKey(bob.PublicKey())
	bob.SetRemotePublicKey(alice.PublicKey())

	// Alice -> Bob
	msg1 := []byte("Hello Bob!")
	enc1, _ := alice.Encrypt(msg1)
	dec1, err := bob.Decrypt(enc1)
	if err != nil || !bytes.Equal(dec1, msg1) {
		t.Error("Alice -> Bob failed")
	}

	// Bob -> Alice
	msg2 := []byte("Hello Alice!")
	enc2, _ := bob.Encrypt(msg2)
	dec2, err := alice.Decrypt(enc2)
	if err != nil || !bytes.Equal(dec2, msg2) {
		t.Error("Bob -> Alice failed")
	}
}

func TestDecryptionBeforeKeyExchange(t *testing.T) {
	session, _ := NewE2ESession()

	_, err := session.Encrypt([]byte("test"))
	if err == nil {
		t.Error("Encryption should fail before key exchange")
	}

	_, err = session.Decrypt("dGVzdA==")
	if err == nil {
		t.Error("Decryption should fail before key exchange")
	}
}

func TestInvalidPublicKey(t *testing.T) {
	session, _ := NewE2ESession()

	// Invalid base64
	err := session.SetRemotePublicKey("not-valid-base64!!!")
	if err == nil {
		t.Error("Should reject invalid base64")
	}

	// Wrong length
	err = session.SetRemotePublicKeyBytes([]byte("too short"))
	if err == nil {
		t.Error("Should reject wrong length key")
	}
}

func TestTamperedCiphertext(t *testing.T) {
	alice, _ := NewE2ESession()
	bob, _ := NewE2ESession()

	alice.SetRemotePublicKey(bob.PublicKey())
	bob.SetRemotePublicKey(alice.PublicKey())

	ciphertext, _ := alice.EncryptBytes([]byte("secret message"))

	// Tamper with ciphertext
	if len(ciphertext) > 30 {
		ciphertext[30] ^= 0xFF
	}

	_, err := bob.DecryptBytes(ciphertext)
	if err == nil {
		t.Error("Should detect tampered ciphertext")
	}
}

func TestFingerprint(t *testing.T) {
	session, _ := NewE2ESession()

	fp := session.Fingerprint()
	if len(fp) != 16 { // 8 bytes * 2 hex chars
		t.Errorf("Fingerprint length should be 16, got %d", len(fp))
	}

	// Remote fingerprint should be empty before key exchange
	rfp := session.RemoteFingerprint()
	if rfp != "" {
		t.Error("Remote fingerprint should be empty before key exchange")
	}
}

func TestUniqueEncryptions(t *testing.T) {
	alice, _ := NewE2ESession()
	bob, _ := NewE2ESession()

	alice.SetRemotePublicKey(bob.PublicKey())

	message := []byte("same message")

	// Encrypt same message twice
	cipher1, _ := alice.Encrypt(message)
	cipher2, _ := alice.Encrypt(message)

	// Ciphertexts should be different (random nonce)
	if cipher1 == cipher2 {
		t.Error("Same message should produce different ciphertexts")
	}
}

func BenchmarkEncrypt(b *testing.B) {
	alice, _ := NewE2ESession()
	bob, _ := NewE2ESession()
	alice.SetRemotePublicKey(bob.PublicKey())

	message := bytes.Repeat([]byte("A"), 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alice.Encrypt(message)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	alice, _ := NewE2ESession()
	bob, _ := NewE2ESession()
	alice.SetRemotePublicKey(bob.PublicKey())
	bob.SetRemotePublicKey(alice.PublicKey())

	ciphertext, _ := alice.Encrypt(bytes.Repeat([]byte("A"), 1024))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bob.Decrypt(ciphertext)
	}
}
