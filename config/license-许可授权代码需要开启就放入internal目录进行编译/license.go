package license

import (
	"encoding/json"
	"fmt"
	"os"
	"ops-web/internal/logger"
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
	// LicenseFilePath 许可文件路径
	LicenseFilePath = "config/license.json"
	// DateFormat 日期格式
	DateFormat = "2006-01-02"
)

// ValidateLicense 验证许可是否有效
// 返回 (isValid, errorMessage)
func ValidateLicense() (bool, string) {
	// 1. 读取许可文件
	licenseInfo, err := GetLicenseInfo()
	if err != nil {
		logger.Errorf("许可验证-读取许可文件失败: %v", err)
		return false, "系统错误，请联系管理员"
	}

	// 2. 检查授权日期是否过期
	if IsLicenseExpired() {
		logger.Errorf("许可验证-授权已过期: 到期日期=%s", licenseInfo.ExpireDate)
		return false, "系统错误，请联系管理员"
	}

	// 3. 检查MAC地址是否匹配
	match, err := IsMacAddressMatch()
	if err != nil {
		logger.Errorf("许可验证-MAC地址检查失败: %v", err)
		return false, "系统错误，请联系管理员"
	}
	if !match {
		logger.Errorf("许可验证-MAC地址不匹配: 许可MAC=%s, 服务器MAC=%s", licenseInfo.MacAddress, getServerMac())
		return false, "系统错误，请联系管理员"
	}

	return true, ""
}

// GetLicenseInfo 获取许可信息
func GetLicenseInfo() (*LicenseInfo, error) {
	// 检查文件是否存在
	if _, err := os.Stat(LicenseFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("许可文件不存在: %s", LicenseFilePath)
	}

	// 读取文件内容
	encryptedData, err := os.ReadFile(LicenseFilePath)
	if err != nil {
		return nil, fmt.Errorf("读取许可文件失败: %v", err)
	}

	// 解密数据
	decryptedData, err := DecryptLicenseData(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("解密许可文件失败: %v", err)
	}

	// 解析JSON
	var licenseInfo LicenseInfo
	if err := json.Unmarshal(decryptedData, &licenseInfo); err != nil {
		return nil, fmt.Errorf("解析许可文件失败: %v", err)
	}

	// 标准化MAC地址格式
	licenseInfo.MacAddress = NormalizeMacAddress(licenseInfo.MacAddress)

	return &licenseInfo, nil
}

// IsLicenseExpired 检查许可是否过期
func IsLicenseExpired() bool {
	licenseInfo, err := GetLicenseInfo()
	if err != nil {
		return true // 如果无法读取许可，视为过期
	}

	// 解析到期日期
	expireDate, err := time.Parse(DateFormat, licenseInfo.ExpireDate)
	if err != nil {
		logger.Errorf("许可验证-解析到期日期失败: %v", err)
		return true
	}

	// 获取当前日期（只比较日期部分，不考虑时间）
	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	expireDateOnly := time.Date(expireDate.Year(), expireDate.Month(), expireDate.Day(), 0, 0, 0, 0, expireDate.Location())

	// 如果当前日期大于到期日期，则已过期
	return currentDate.After(expireDateOnly)
}

// IsMacAddressMatch 检查MAC地址是否匹配
func IsMacAddressMatch() (bool, error) {
	licenseInfo, err := GetLicenseInfo()
	if err != nil {
		return false, err
	}

	// 获取服务器MAC地址
	serverMac, err := GetServerMacAddress()
	if err != nil {
		return false, err
	}

	// 标准化两个MAC地址
	licenseMac := NormalizeMacAddress(licenseInfo.MacAddress)
	serverMac = NormalizeMacAddress(serverMac)

	// 比较MAC地址（不区分大小写）
	return strings.EqualFold(licenseMac, serverMac), nil
}

// getServerMac 获取服务器MAC地址（内部辅助函数，用于日志）
func getServerMac() string {
	mac, err := GetServerMacAddress()
	if err != nil {
		return "获取失败"
	}
	return mac
}

