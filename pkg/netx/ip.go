package netx

import "net"

// GetOutboundIP 获得对外通信的 IP 地址
// 通常拿到的是自己网络的局域网 IP，不过一般注册都只考虑服务端和客户端都在一个局域网内
// 如果客户端不在，那通过注册中心拿到这个注册的 IP 也没法访问到服务端
// 此时可以考虑，通过读取配置、从环境变量里面注入公网 IP 的方法，注册该服务端的公网 IP
func GetOutboundIP() string {
	// 发送到 dns 服务器 8.8.8.8:80
	conn, err := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	if err != nil {
		return ""
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()

}
