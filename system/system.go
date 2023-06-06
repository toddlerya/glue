package system

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/host"
	"github.com/sirupsen/logrus"
	"github.com/toddlerya/glue/command"
	"github.com/toddlerya/glue/kit"
	"github.com/vishvananda/netlink"
)

type CurrentUserInfo struct {
	Username string
	Name     string
	HomeDir  string
	Uid      string
	Gid      string
}

type NetInterfaceInfo struct {
	Index        int    // positive integer that starts at one, zero is never used
	MTU          int    // maximum transmission unit
	Name         string // e.g., "en0", "lo0", "eth0.100"
	HardwareAddr string // IEEE MAC-48, EUI-48 and EUI-64 form
	Flags        string // e.g., FlagUp, FlagLoopback, FlagMulticast
	IPV4Addr     string
	IPV6Addr     string
}

func GetHomeDir() string {
	user, err := user.Current()
	if err != nil {
		fmt.Printf("获取用户目录失败: %s\n", err.Error())
		os.Exit(-1)
	}
	return user.HomeDir
}

// 获取系统信息
func GetHostInfo() (host.InfoStat, error) {
	hostInfo, err := host.Info()
	return *hostInfo, err
}

// 获取本机首选IP地址
func GetOutboundIP() (string, error) {
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 5*time.Second)
	if err != nil {
		// 解决无法进行UDP拨号导致的panic错误
		// dial udp 8.8.8.8:80: connect: network is unreachable
		// panic: runtime error: invalid memory address or nil pointer dereference
		// [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x10c1909]
		return "127.0.0.1", err
	}

	defer func() {
		// 当 net.Dial() 返回错误并且 conn 对象为 nil 时，调用 conn.Close() 方法会导致空指针异常。
		// 为了避免这种情况，可以在 err 不为 nil 时，直接返回错误，并不执行 conn.Close() 方法
		if conn != nil {
			conn.Close()
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP.String()
	return ip, err
}

// 获取本机首选IP地址的另一种备选方案
func GetOutboundIPByInterfaceAndRoute() (string, error) {
	// 获取所有路由
	ipAddress := "127.0.0.1"
	defaultGateWay := ""
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return ipAddress, err
	}

	// 获取默认路由
	for _, route := range routes {
		if route.Gw != nil {
			defaultGateWay = route.Gw.To4().String()
		}
	}
	logrus.Debugf("defaultGateWay: %s", defaultGateWay)

	netInterfaceInfoSlice, err := GetNetInterfacesInfo()
	if err != nil {
		return ipAddress, err
	} else {
		for _, netInter := range netInterfaceInfoSlice {
			// 如果获取到默认网关，使用默认网关的前两位匹配IP地址
			if len(defaultGateWay) > 3 {
				// 10.1.2.254的网关前缀 10.1
				gateWayPrefix := strings.Join(strings.Split(defaultGateWay, ".")[0:2], ".")
				if strings.HasPrefix(netInter.IPV4Addr, gateWayPrefix) {
					ipAddress = netInter.IPV4Addr
					break
				}
			} else {
				// 网卡命名规范
				// https://blog.51cto.com/u_15127507/3941816
				// https://developer.aliyun.com/article/609587
				// 查找物理网卡
				if netInter.IPV4Addr != "" {
					netNamePatten := regexp.MustCompile(`(en\w+|wl\w+|ww\w+)`)
					if netNamePatten.MatchString(netInter.Name) {
						ipAddress = netInter.IPV4Addr
						break
					}
				}
			}
		}
	}
	return ipAddress, err
}

// 获取本机所有网卡对应的IP地址信息
func GetNetInterfacesInfo() ([]NetInterfaceInfo, error) {
	netInterfaceInfoSlice := []NetInterfaceInfo{}
	interfaces, err := net.Interfaces()
	if err != nil {
		return netInterfaceInfoSlice, err
	}
	for _, inter := range interfaces {
		netInterfaceInfo := NetInterfaceInfo{
			Index:        inter.Index,
			Flags:        inter.Flags.String(),
			HardwareAddr: inter.HardwareAddr.String(),
			MTU:          inter.MTU,
			Name:         inter.Name,
		}
		addrs, err := inter.Addrs()
		if err != nil {
			return netInterfaceInfoSlice, err
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ipNet.IP.To4() != nil {
					netInterfaceInfo.IPV4Addr = ipNet.IP.String()
				} else if ipNet.IP.To16() != nil {
					netInterfaceInfo.IPV6Addr = ipNet.IP.String()
				}
			}
		}

		netInterfaceInfoSlice = append(netInterfaceInfoSlice, netInterfaceInfo)
	}
	return netInterfaceInfoSlice, err
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

// 获取当前用户信息
func GetCurrentUserInfo() (CurrentUserInfo, error) {
	currentUserInfo := CurrentUserInfo{}
	user, err := user.Current()
	if err != nil {
		return currentUserInfo, err
	}
	currentUserInfo.Gid = user.Gid
	currentUserInfo.Uid = user.Uid
	currentUserInfo.HomeDir = user.HomeDir
	currentUserInfo.Name = user.Name
	currentUserInfo.Username = user.Username
	return currentUserInfo, err
}

// 获取进程当前的运行目录
func GetExecDir() (string, error) {
	// 当前运行目录
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取当前运行目录失败! ERROR: %s", err.Error())
	}
	execDir := filepath.Dir(execPath)
	return execDir, err
}
