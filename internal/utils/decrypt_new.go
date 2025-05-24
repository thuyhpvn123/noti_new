package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	secp "github.com/meta-node-blockchain/noti-contract/internal/secp256k1-cgo/secp"
)

// ECDH + SHA256 with version byte 0x02
func ECDHSharedSecretHex(privBytes, pubBytes []byte) (string, error) {
	fmt.Println("privBytes:",hex.EncodeToString(privBytes))
	fmt.Println("pubBytes:",hex.EncodeToString(pubBytes))

	shared, err := secp.CreateECDH(hex.EncodeToString(privBytes), hex.EncodeToString(pubBytes))
	if err != nil {
		return "", err
	}

	return shared, nil
}


func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func unpadPKCS7(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	padding := int(data[length-1])
	if padding > length {
		return nil, fmt.Errorf("invalid padding")
	}
	return data[:length-padding], nil
}

func EncryptAESCBC(key []byte, plaintext []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext = padPKCS7(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

func DecryptAESCBC(ciphertext []byte, privHexB, pubHexA []byte, iv []byte) ([]byte, error) {
	sharedBHex, err := ECDHSharedSecretHex(privHexB, pubHexA)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECDH shared secret: %v", err)
	}
	fmt.Println("sharedBHex:",sharedBHex)
	fmt.Println("iv:",iv)
	sharedBBytes, err := hex.DecodeString(sharedBHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode shared secret: %v", err)
	}

	block, err := aes.NewCipher(sharedBBytes)
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of the block size: got %d", len(ciphertext))
	}
	mode.CryptBlocks(plaintext, ciphertext)
	return unpadPKCS7(plaintext)
}
