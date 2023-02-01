package request

import (
	"fmt"
	"net/url"
)

// 处理URL，获取移出query内容后的基础url
// http://172.16.45.106:8113/api/btnAuthInfo?searchId=1670398510299 --> http://172.16.45.106:8113/api/btnAuthInfo
func GetBaseUrlWithoutQueryString(rawUrl string) (string, error) {
	var baseUrl string
	u, err := url.Parse(rawUrl)
	if err != nil {
		return baseUrl, fmt.Errorf("url.Parse error: %s", err.Error())
	}
	baseUrl += u.Scheme
	baseUrl += "://"
	baseUrl += u.Host
	baseUrl += u.Path
	return baseUrl, err
}
