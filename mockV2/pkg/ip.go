package pkg

import "net"

// GetLocalIP gets local IP address
func GetLocalIP() string {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, ddr := range addr {
		if ipNet, ok := ddr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
