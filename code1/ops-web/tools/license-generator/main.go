package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// LicenseInfo 许可信息结构
type LicenseInfo struct {
	ExpireDate string `json:"expireDate"`
	MacAddress string `json:"macAddress"`
	IssueDate  string `json:"issueDate"`
	LicenseKey string `json:"licenseKey,omitempty"`
}

const (
	DateFormat = "2006-01-02"
)

func main() {
	// 定义命令行参数
	expireDate := flag.String("expire", "", "到期日期（必填，格式：YYYY-MM-DD）")
	macAddress := flag.String("mac", "", "MAC地址（必填，格式：XX:XX:XX:XX:XX:XX）")
	outputFile := flag.String("output", "license.json", "输出文件路径（可选，默认：license.json）")
	issueDate := flag.String("issue", "", "签发日期（可选，默认：当前日期，格式：YYYY-MM-DD）")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "许可生成工具\n")
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s -expire 2026-12-30 -mac 00:11:22:33:44:55\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -expire 2026-12-30 -mac 00:11:22:33:44:55 -output config/license.json\n", os.Args[0])
	}

	flag.Parse()

	// 验证必填参数
	if *expireDate == "" {
		fmt.Fprintf(os.Stderr, "错误: 必须指定到期日期 (-expire)\n")
		flag.Usage()
		os.Exit(1)
	}

	if *macAddress == "" {
		fmt.Fprintf(os.Stderr, "错误: 必须指定MAC地址 (-mac)\n")
		flag.Usage()
		os.Exit(1)
	}

	// 验证日期格式
	_, err := time.Parse(DateFormat, *expireDate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 到期日期格式不正确，应为 YYYY-MM-DD，实际: %s\n", *expireDate)
		os.Exit(1)
	}

	// 标准化MAC地址
	mac := normalizeMacAddress(*macAddress)
	if mac == "" {
		fmt.Fprintf(os.Stderr, "错误: MAC地址格式不正确，应为 XX:XX:XX:XX:XX:XX，实际: %s\n", *macAddress)
		os.Exit(1)
	}

	// 设置签发日期
	issue := *issueDate
	if issue == "" {
		issue = time.Now().Format(DateFormat)
	} else {
		// 验证签发日期格式
		_, err := time.Parse(DateFormat, issue)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 签发日期格式不正确，应为 YYYY-MM-DD，实际: %s\n", issue)
			os.Exit(1)
		}
	}

	// 创建许可信息
	licenseInfo := LicenseInfo{
		ExpireDate: *expireDate,
		MacAddress: mac,
		IssueDate:  issue,
	}

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(licenseInfo, "", "    ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 序列化许可信息失败: %v\n", err)
		os.Exit(1)
	}

	// 加密数据
	encryptedData, err := encryptLicenseData(jsonData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 加密许可数据失败: %v\n", err)
		os.Exit(1)
	}

	// 写入文件（加密后的数据）
	err = os.WriteFile(*outputFile, encryptedData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("许可文件生成成功！\n")
	fmt.Printf("文件路径: %s\n", *outputFile)
	fmt.Printf("到期日期: %s\n", licenseInfo.ExpireDate)
	fmt.Printf("MAC地址: %s\n", licenseInfo.MacAddress)
	fmt.Printf("签发日期: %s\n", licenseInfo.IssueDate)
}

// normalizeMacAddress 标准化MAC地址格式
func normalizeMacAddress(mac string) string {
	if mac == "" {
		return ""
	}

	// 先将所有字符转为大写
	mac = strings.ToUpper(mac)

	// 将连字符替换为冒号
	mac = strings.ReplaceAll(mac, "-", ":")

	// 移除多余的空格
	mac = strings.ReplaceAll(mac, " ", "")

	// 验证MAC地址格式（应该是 XX:XX:XX:XX:XX:XX，共17个字符）
	if len(mac) != 17 {
		return ""
	}

	// 验证是否包含6个冒号分隔的十六进制数
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return ""
	}

	// 验证每部分是否为2位十六进制数
	for _, part := range parts {
		if len(part) != 2 {
			return ""
		}
		// 简单的十六进制验证
		for _, c := range part {
			if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
				return ""
			}
		}
	}

	return mac
}

// 加密密钥（必须与 internal/license/crypto.go 中的密钥相同）
var encryptionKey = []byte("ops-web-license-encryption-key-2024-v1")

// encryptLicenseData 加密许可数据
func encryptLicenseData(plaintext []byte) ([]byte, error) {
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

// pkcs7Padding PKCS7填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

