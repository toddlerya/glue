package sysguard

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/toddlerya/glue/files"
)

// 自动选择使用部署模式
func SetupService(systemdServiceConfig SystemdServiceConfig) error {
	mode, err := AutoChoseSetupMode()
	if err != nil {
		return err
	}
	switch mode {
	case "Systemd":
		err = SetupSystemdService(systemdServiceConfig)
	case "SysVinit":
		err = SetupSysVinitService(systemdServiceConfig)
	}
	return err
}

// 自动选择使用卸载模式
func UnSetupService(systemdServiceConfig SystemdServiceConfig, deleteExporterWorkingDirectory bool) error {
	mode, err := AutoChoseSetupMode()
	if err != nil {
		return err
	}
	switch mode {
	case "Systemd":
		err = UnSetupSystemService(systemdServiceConfig, deleteExporterWorkingDirectory)
	case "SysVinit":
		err = UnSetupSysVinitService(systemdServiceConfig, deleteExporterWorkingDirectory)
	}
	return err
}

// 根据操作系统自动选择使用systemd还是SysVinit
func AutoChoseSetupMode() (string, error) {
	// 判断init模式
	var mode string
	var err error
	// /proc/1/comm is only available in Linux 2.6.33 and later.
	// 在Linux内核版本低于2.6.33的系统中，有/proc/1/cmdline，内容为/sbin/init
	proc1CommFile := `/proc/1/comm`
	if files.PathIsExist(proc1CommFile) {
		byteSlice, err := files.ReadFileAsByteSlice(proc1CommFile)
		if err != nil {
			return mode, fmt.Errorf("read %s failed: %s", proc1CommFile, err.Error())
		}
		if len(byteSlice) != 0 {
			content := string(byteSlice)
			switch strings.TrimSpace(content) {
			case "systemd":
				logrus.Info("Systemd setup")
				mode = "Systemd"
			case "init":
				logrus.Info("SysVinit setup")
				mode = "SysVinit"
			default:
				// 不知道是什么情况，使用SysVinit尝试
				logrus.Warnf("%s is %s unknown init system, try SysVinit setup", proc1CommFile, strings.TrimSpace(content))
				mode = "SysVinit"
			}
		} else {
			// 不知道是什么情况，使用SysVinit尝试
			mode = "SysVinit"
			logrus.Warnf("%s is empty, unknown init system, try SysVinit setup", proc1CommFile)
		}
	} else {
		// 不知道是什么情况，使用SysVinit尝试
		mode = "SysVinit"
		logrus.Warnf("%s not found, unknown init system, try SysVinit setup", proc1CommFile)
	}
	return mode, err
}
