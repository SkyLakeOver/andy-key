package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const KEY_PATH = ".andy-key.key" // ИЗМЕНЕНО: .master.key → .andy-key.key

type Crypto struct {
	key []byte
}

func NewCrypto() (*Crypto, error) {
	if _, err := os.Stat(KEY_PATH); err == nil {
		data, err := os.ReadFile(KEY_PATH)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения ключа: %v", err)
		}
		key, err := hex.DecodeString(string(data))
		if err != nil {
			return nil, fmt.Errorf("ошибка декодирования ключа: %v", err)
		}
		return &Crypto{key: key}, nil
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("ошибка генерации ключа: %v", err)
	}

	if err := os.WriteFile(KEY_PATH, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return nil, fmt.Errorf("ошибка сохранения ключа: %v", err)
	}

	return &Crypto{key: key}, nil
}

func (c *Crypto) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (c *Crypto) Decrypt(ciphertextHex string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}