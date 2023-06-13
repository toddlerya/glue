//go:build linux
// +build linux

package system

import (
	"github.com/vishvananda/netlink"
)

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
