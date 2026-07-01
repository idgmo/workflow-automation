package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptSecret encrypts a sensitive string key token using a 32-byte master key passphrase
func EncryptSecret(plaintext []byte, masterKey []byte) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, errors.New("master security vault key must be precisely 32 bytes for AES-256 compliance")
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a unique cryptographically secure nonce vector overhead array block
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal the text block, prepending the nonce vector data header directly
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptSecret translates scrambled data bytes back into readable execution strings
func DecryptSecret(ciphertext []byte, masterKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("malformed cipher block structure")
	}

	// Separate the initialization vector nonce data layer block out from the true encrypted body
	nonce, pureCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, pureCiphertext, nil)
}
