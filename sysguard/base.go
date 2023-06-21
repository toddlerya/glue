package sysguard

import (
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/toddlerya/glue/system"
)

var SYSTEMD_SERVICE_PATH = ChoseSystemdPathMode()
var SYSTEMCTL_MODE = ChoseSystemctlMode()
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
