package sysguard

import (
	"fmt"
	"github.com/toddlerya/glue/system"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

var SYSTEMD_SERVICE_PATH = ChoseSystemdPathMode()
var SYSTEMCTL_MODE = ChoseSystemctlMode()

// systemd service meta config
// ref: https://unix.stackexchange.com/questions/224992/where-do-i-put-my-systemd-unit-file
var USER_MODE_SYSTEMD_SERVICE_PATH = filepath.Join(system.GetHomeDir(), ".config", "systemd", "user")
var ROOT_MODE_SYSTEMD_SERVICE_PATH = "/lib/systemd/system"
var ROOT_MODE_SYSVINIT_SCRIPT_PATH = "/etc/init.d"

func ChoseSystemdPathMode() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	if user.Username == "root" {
		fmt.Println("run as root mode")
		return ROOT_MODE_SYSTEMD_SERVICE_PATH
	} else {
		fmt.Println("run as user mode")
		return USER_MODE_SYSTEMD_SERVICE_PATH
	}
}

func ChoseSystemctlMode() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	if user.Username == "root" {
		return ""
	} else {
		return "--user"
	}
}

// 获取当前执行程序文件的所属用户
func CurrentExecutableFileOwnerInfo() string {
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	fileInfo, err := os.Stat(execPath)
	if err != nil {
		panic(err)
	}
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		panic(err)
	}
	user, err := user.LookupId(fmt.Sprintf("%d", stat.Uid))
	if err != nil {
		panic(err)
	}
	return user.Username
}

// 获取当前shell登陆的用户
func CurrentShellUserName() string {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser == "" {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}
		return u.Name
	} else {
		return sudoUser
	}
}
