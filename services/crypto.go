package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/scrypt"
)

// ──────────────────────────────────────────────────────────
// ADVANCED FEATURE: Encrypted Backup/Restore
// Demonstrates: AES-256-GCM, scrypt key derivation, crypto
// ──────────────────────────────────────────────────────────

const (
	saltSize  = 32 // 256-bit salt
	keySize   = 32 // AES-256
	scryptN   = 32768
	scryptR   = 8
	scryptP   = 1
)

// Encrypt encrypts data using AES-256-GCM with a password-derived key.
// Format: [32-byte salt][12-byte nonce][ciphertext+tag]
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	// Generate random salt for key derivation.
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Derive key from password using scrypt.
	key, err := scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, keySize)
	if err != nil {
		return nil, err
	}

	// Create AES cipher.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM (Galois/Counter Mode) for authenticated encryption.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and authenticate.
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Combine: salt + nonce + ciphertext
	result := make([]byte, 0, saltSize+gcm.NonceSize()+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt decrypts data that was encrypted with Encrypt.
func Decrypt(data []byte, password string) ([]byte, error) {
	if len(data) < saltSize+12 { // salt + minimum nonce size
		return nil, errors.New("ciphertext too short")
	}

	// Extract salt.
	salt := data[:saltSize]
	rest := data[saltSize:]

	// Derive the same key from password.
	key, err := scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, keySize)
	if err != nil {
		return nil, err
	}

	// Recreate cipher.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(rest) < nonceSize {
		return nil, errors.New("ciphertext too short for nonce")
	}

	// Extract nonce and ciphertext.
	nonce := rest[:nonceSize]
	ciphertext := rest[nonceSize:]

	// Decrypt and verify.
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: wrong password or corrupted data")
	}

	return plaintext, nil
}
