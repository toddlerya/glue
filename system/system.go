package system

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/host"
	"github.com/toddlerya/glue/command"
	"github.com/toddlerya/glue/kit"
)

// 获取系统信息
func GetHostInfo() (host.InfoStat, error) {
	hostInfo, err := host.Info()
	return *hostInfo, err
}

// 获取本机首选IP地址
func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// log.WithFields(log.Fields{"net": "net.Dial()"}).Error(err)
		// 解决无法进行UDP拨号导致的panic错误
		// dial udp 8.8.8.8:80: connect: network is unreachable
		// panic: runtime error: invalid memory address or nil pointer dereference
		// [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x10c1909]
		return "127.0.0.1", err
	}
	// 只有conn对象不为nil才能执行Close，直接defer Close是有空指针风险的
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP.String()
	return ip, err
}

// 获取本机所有的IP, 效果等同于 ip addr | grep inet | awk -F " " '{ print $2}'
func GetAllIP() (map[string][]string, error) {
	allIPInfo := map[string][]string{}
	ipArray, err := net.InterfaceAddrs()
	if err == nil {
		for _, ip := range ipArray {
			if strings.Contains(ip.String(), ".") {
				allIPInfo["IPV4"] = append(allIPInfo["IPV4"], ip.String())
			} else if strings.Contains(ip.String(), ":") {
				allIPInfo["IPV6"] = append(allIPInfo["IPV6"], ip.String())
			}
		}
	}
	return allIPInfo, err
}

// 获取本机所有IPv4地址
func GetAllIPV4Slice() ([]string, error) {
	var ipv4Slice []string
	hostIPMap, err := GetAllIP()
	if err != nil {
		return ipv4Slice, fmt.Errorf("获取宿主机IPv4信息错误: %s", err.Error())

	}
	rawIpv4Slice := hostIPMap["IPV4"]
	//IPV4Slice: [127.0.0.1/8 10.1.2.132/24 172.19.0.1/16 192.182.0.1/16 192.100.0.1/16]
	for _, v := range rawIpv4Slice {
		ip := strings.Split(v, "/")[0]
		ipv4Slice = append(ipv4Slice, ip)
	}
	return ipv4Slice, err
}

// 随机生成一个IP网段
func RandomSubnetIP() string {
	rand.Seed(time.Now().UnixNano())
	ip := fmt.Sprintf("%d.%d.%d.1", 100+rand.Intn(100), 100+rand.Intn(140), 100+rand.Intn(140))
	return ip
}

// 带黑名单的生成随机IP网段
func RandomSubnetIPWithBlackList(blackList []string) string {
	ip := RandomSubnetIP()
	// 将具体ip改为网段ip，比如10.1.2.132改为10.1.0.1
	for index, ele := range blackList {
		spEle := strings.Split(ele, ".")
		spEle[2] = "0"
		spEle[3] = "1"
		newEle := strings.Join(spEle, ".")
		blackList[index] = newEle
	}
	if kit.Contains(blackList, ip) {
		return RandomSubnetIPWithBlackList(blackList)
	} else {
		return ip
	}
}

// 校验端口是否被使用，端口未使用则返回true
func VerifyPortIsUnused(port uint16) (bool, error) {
	var result bool
	ip, _ := GetOutboundIP()
	listen, err := net.Listen("tcp", net.JoinHostPort(ip, strconv.Itoa(int(port))))
	// 有错误说明端口已被占用
	if err != nil {
		if strings.HasSuffix(err.Error(), "bind: address already in use") {
			result = false
			err = nil
		}
	} else {
		result = true
		defer listen.Close()
	}
	return result, err
}

/*
检查运行环境类型, Linux主机、K8S的Pod、Docker容器
- host: Linux主机
- pod: K8S的Pod
- container: Docker容器
- unknown: 其他未知的情况
*/
func VerifyRuntimeEnv() (string, error) {
	stdout, stderr, err := command.RunByBash("cat cgroup", "cat /proc/1/cpuset")
	if err != nil {
		return "", fmt.Errorf("检查运行环境类型失败! STDERR: %s ERROR: %s", stderr, err.Error())
	}
	stdout = strings.TrimSpace(stdout)
	if stdout == "/" {
		return "host", err
	} else if strings.HasPrefix(stdout, "/kubepods/") {
		return "pod", err
	} else if strings.HasPrefix(stdout, "/docker/") {
		return "container", err
	} else {
		return "unknown", err
	}
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
