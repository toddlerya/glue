package system

import (
	"net"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/host"
	"github.com/toddlerya/glue/kit"
)

// 获取系统信息
func GetHostInfo() (host.InfoStat, error) {
	hostInfo, err := host.Info()

	return *hostInfo, err
}

// 获取本机首选IP地址
func GetOutBoundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		defer conn.Close()
	}

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP.String()
	return ip, err
}

// 校验端口是否被使用，端口未使用则返回true
func VerifyPortIsUnused(port uint16) (bool, error) {
	var result bool = false
	// 改为使用本地回环地址来检测端口号是否被占用
	// listen, err := net.Listen("tcp", net.JoinHostPort(GetOutBoundIP(), strconv.Itoa(port)))
	listen, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port))))
	// 有错误说明端口已被占用
	if err != nil {
		if !strings.HasSuffix(err.Error(), "address already in use") {
			return result, err
		}
	} else {
		result = true
		defer listen.Close()
	}
	return result, err
}

// 随机选取可用的端口号
func RandomPort(blackList []uint16) uint16 {
	var port uint16
	for p := uint16(40000); p <= 50000; p++ {
		if kit.Contains(blackList, p) {
			continue
		}
		ok, _ := VerifyPortIsUnused(p)
		if ok {
			port = uint16(p)
			break
		}
	}
	return port
}
