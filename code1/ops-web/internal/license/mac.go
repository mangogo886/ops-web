package license

import (
	"fmt"
	"net"
	"strings"
)

// GetServerMacAddress 获取服务器MAC地址
// 返回格式：XX:XX:XX:XX:XX:XX（大写，冒号分隔）
func GetServerMacAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("获取网络接口失败: %v", err)
	}

	// 遍历所有网络接口，找到第一个非环回、有效的接口
	for _, iface := range interfaces {
		// 跳过环回接口
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 跳过无效的接口
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 获取MAC地址
		mac := iface.HardwareAddr
		if len(mac) > 0 {
			return NormalizeMacAddress(mac.String()), nil
		}
	}

	return "", fmt.Errorf("未找到有效的网络接口")
}

// NormalizeMacAddress 标准化MAC地址格式
// 统一转换为大写，使用冒号分隔
// 输入格式可能是：xx:xx:xx:xx:xx:xx 或 xx-xx-xx-xx-xx-xx
// 输出格式：XX:XX:XX:XX:XX:XX
func NormalizeMacAddress(mac string) string {
	if mac == "" {
		return ""
	}

	// 先将所有字符转为大写
	mac = strings.ToUpper(mac)

	// 将连字符替换为冒号
	mac = strings.ReplaceAll(mac, "-", ":")

	// 移除多余的空格
	mac = strings.ReplaceAll(mac, " ", "")

	return mac
}

