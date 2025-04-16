package util

import "net"

// GetUsableLanIP 获取可用的局域网IP地址
func GetUsableLanIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	// 然后寻找可用的局域网IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				// 检查是否是有效的局域网IP地址
				if isUsableLanIP(ip4) {
					return ip4.String(), nil
				}
			}
		}
	}

	return "", nil
}

// isUsableLanIP 检查IP是否为可用的局域网IP
func isUsableLanIP(ip net.IP) bool {
	// 检查是否是192.168.x.y格式的地址
	if ip[0] == 192 && ip[1] == 168 {
		// 排除子网地址 (192.168.x.0)
		if ip[3] != 0 {
			return true
		}
	}

	// 您可以根据需要添加其他条件
	// 例如检查10.x.y.z或172.16-31.x.y范围的IP

	return false
}
