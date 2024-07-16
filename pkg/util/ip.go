package util

import "net"

func LocalMulIPv4() []string {
	netInters, _ := net.Interfaces()
	var ips []string
	for i := 0; i < len(netInters); i++ {
		if (netInters[i].Flags & net.FlagUp) != 0 {
			addrList, _ := netInters[i].Addrs()

			for _, address := range addrList {
				if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						ips = append(ips, ipNet.IP.String())
					}
				}
			}
		}

	}
	return ips
}
