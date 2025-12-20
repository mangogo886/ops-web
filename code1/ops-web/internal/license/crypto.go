package license

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// 加密密钥（使用固定的密钥，通过SHA256哈希后用于AES-256）
// 注意：密钥硬编码在代码中，编译后无法轻易修改
var encryptionKey = []byte("ops-web-license-encryption-key-2024-v1")

// EncryptLicenseData 加密许可数据
func EncryptLicenseData(plaintext []byte) ([]byte, error) {
	// 使用SHA256哈希密钥，确保密钥长度为32字节（AES-256）
	keyHash := sha256.Sum256(encryptionKey)
	key := keyHash[:]

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %v", err)
	}

	// 生成随机IV（16字节，AES块大小）
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("生成IV失败: %v", err)
	}

	// 使用PKCS7填充（AES块大小对齐）
	plaintext = pkcs7Padding(plaintext, aes.BlockSize)

	// 创建CBC模式的加密器
	mode := cipher.NewCBCEncrypter(block, iv)

	// 加密数据（IV + 加密数据）
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)

	// 将IV和加密数据组合在一起（IV在前，加密数据在后）
	encrypted := make([]byte, len(iv)+len(ciphertext))
	copy(encrypted[:len(iv)], iv)
	copy(encrypted[len(iv):], ciphertext)

	// Base64编码
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(encrypted)))
	base64.StdEncoding.Encode(encoded, encrypted)

	return encoded, nil
}

// DecryptLicenseData 解密许可数据
func DecryptLicenseData(ciphertext []byte) ([]byte, error) {
	// Base64解码
	encrypted, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %v", err)
	}

	// 使用SHA256哈希密钥
	keyHash := sha256.Sum256(encryptionKey)
	key := keyHash[:]

	// 检查数据长度（至少要有IV）
	if len(encrypted) < aes.BlockSize {
		return nil, fmt.Errorf("加密数据长度不足")
	}

	// 提取IV和加密数据
	iv := encrypted[:aes.BlockSize]
	cipherData := encrypted[aes.BlockSize:]

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %v", err)
	}

	// 检查加密数据长度必须是块大小的倍数
	if len(cipherData)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("加密数据长度不正确")
	}

	// 创建CBC模式的解密器
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密数据
	plaintext := make([]byte, len(cipherData))
	mode.CryptBlocks(plaintext, cipherData)

	// 去除PKCS7填充
	plaintext, err = pkcs7UnPadding(plaintext)
	if err != nil {
		return nil, fmt.Errorf("去除填充失败: %v", err)
	}

	return plaintext, nil
}

// pkcs7Padding PKCS7填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7UnPadding 去除PKCS7填充
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("数据为空")
	}

	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, fmt.Errorf("填充长度无效")
	}

	return data[:(length - unpadding)], nil
}

