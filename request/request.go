package request

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/toddlerya/glue/files"
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

// HTTP协议下载文件存储到本地，若成功则返回 true, <nil>
func HttpDownload(url string, savePath string) (bool, error) {
	saveDir := filepath.Dir(savePath)
	err := files.CreateDirIfNotExist(saveDir, os.ModePerm)
	if err != nil {
		return false, err
	}
	save, err := os.Create(savePath)
	if err != nil {
		return false, err
	}
	defer save.Close()
	response, err := http.Get(url)
	if err != nil {
		return false, err
	} else {
		if response.StatusCode != 200 {
			return false, errors.New(response.Status)
		}
	}
	defer response.Body.Close()

	_, err = io.Copy(save, response.Body)
	if err != nil {
		return false, err
	}
	return true, err
}
