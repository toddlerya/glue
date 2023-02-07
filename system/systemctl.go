package system

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/toddlerya/glue/command"
	"github.com/toddlerya/glue/files"
)

// systemd service config template
var serviceTemplate = `
[Unit]
Description={{ .Description }}
ConditionPathExists={{ .WorkingDirectory }}
After=network.target

[Service]
Type=simple
WorkingDirectory={{ .WorkingDirectory }}
ExecStart={{ .ExecStart }}
Restart=on-failure
RestartSec=5s
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier={{ .Name }}

[Install]
WantedBy=multi-user.target
`

type SystemdServiceConfig struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	WorkingDirectory string `json:"working_directory"`
	ExecStart        string `json:"exec_start"`
}

var SYSTEMD_SERVICE_PATH = filepath.Join(GetHomeDir(), ".config", "systemd", "user")

// 生成systemd配置文件
func GenSystemdServiceConfigFile(systemdServiceConfig SystemdServiceConfig) error {
	tmpl, err := template.New(systemdServiceConfig.Name).Parse(serviceTemplate)
	if err != nil {
		return err
	}
	serviceConfigFilePath := filepath.Join(SYSTEMD_SERVICE_PATH, systemdServiceConfig.Name+".service")
	f, err := os.Create(serviceConfigFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = tmpl.Execute(f, systemdServiceConfig)
	return err
}

func PreSetupSystemdService() error {

	err := files.CreateDirIfNotExist(SYSTEMD_SERVICE_PATH, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("创建%s目录失败: %s", SYSTEMD_SERVICE_PATH, err.Error())
	}
	return err
}

/*
部署exporter
1. 渲染生成service文件
2. systemctl --user daemon-reload  -- 加载配置
3. systemctl --user enable appName -- 设为开机启动
4. systemctl --user start appName  -- 启动服务
5. systemctl --user status appName -- 检查是否启动成功
*/
func SetupSystemdService(systemdServiceConfig SystemdServiceConfig) error {
	// 渲染生成service文件
	err := GenSystemdServiceConfigFile(systemdServiceConfig)
	if err != nil {
		return err
	}
	// 加载配置
	reloadStdout, reloadStderr, err := command.RunByBash("systemd daemon-reload", "systemctl --user daemon-reload")
	if err != nil {
		return err
	}
	if reloadStdout != "" || reloadStderr != "" {
		return fmt.Errorf("加载%s systemd配置失败! stdout: %s stderr: %s", systemdServiceConfig.Name, reloadStdout, reloadStderr)
	}
	// 设为开机启动
	enableStdout, enableStderr, err := command.RunByBash("systemd enable", "systemctl --user enable "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if enableStdout != "" || !strings.HasPrefix(enableStderr, "Created symlink") {
		return fmt.Errorf("%s设为开机启动失败! stdout: %s stderr: %s", systemdServiceConfig.Name, enableStdout, enableStderr)
	}
	// 启动服务
	startStdout, startStderr, err := command.RunByBash("systemd start", "systemctl --user start "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if startStdout != "" || startStderr != "" {
		return fmt.Errorf("%s启动失败! stdout: %s stderr: %s", systemdServiceConfig.Name, startStdout, startStderr)
	}
	// 检查服务启动状态
	statusStdout, statusStderr, err := command.RunByBash("systemd status", "systemctl --user status "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if !strings.Contains(statusStdout, "Active: active (running)") || statusStderr != "" {
		return fmt.Errorf("检查%s运行状态失败! stdout: %s stderr: %s", systemdServiceConfig.Name, statusStdout, statusStderr)
	}
	return nil
}

/*
卸载清理
1. systemctl --user stop appName -- 停止服务
2. systemctl --user disable appName -- 禁用服务
3. rm ~/.config/systemd/user/appName.service -- 删除服务的systemd配置
4. rm -r ~/.Flame/exporters/appName -- 删除服务文件和配置 【非必须】
*/
func UnSetupSystemService(systemdServiceConfig SystemdServiceConfig, deleteExporterWorkingDirectory bool) error {
	// 停止服务
	stopStdout, stopStderr, err := command.RunByBash("systemd start", "systemctl --user stop "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if stopStdout != "" || stopStderr != "" {
		return fmt.Errorf("%s停止失败! stdout: %s stderr: %s", systemdServiceConfig.Name, stopStdout, stopStderr)
	}
	// 检测服务停止状态, 需要确认进程是否停止成功
	statusStdout, statusStderr, err := command.RunByBash("systemd status", "systemctl --user status "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if !strings.Contains(statusStdout, "Active: inactive (dead)") || statusStderr != "" {
		return fmt.Errorf("检查%s运行状态失败! stdout: %s stderr: %s", systemdServiceConfig.Name, statusStdout, statusStderr)
	}
	// 禁用服务
	disableStdout, disableStderr, err := command.RunByBash("systemd enable", "systemctl --user disable "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if disableStdout != "" || (disableStderr != "" && !strings.HasPrefix(disableStderr, "Removed")) {
		return fmt.Errorf("%s禁用失败! stdout: %s stderr: %s", systemdServiceConfig.Name, disableStdout, disableStderr)
	}
	// 删除服务的systemd配置
	if err = files.Delete(filepath.Join(SYSTEMD_SERVICE_PATH, systemdServiceConfig.Name+".service")); err != nil {
		return fmt.Errorf("%s 守护配置清理失败! ERROR: %s", systemdServiceConfig.Name, err.Error())
	}
	// 删除服务文件和配置
	if deleteExporterWorkingDirectory {
		if err = files.Delete(systemdServiceConfig.WorkingDirectory); err != nil {
			return fmt.Errorf("%s工作目录清理失败! ERROR: %s", systemdServiceConfig.Name, err.Error())
		}
	}
	return err
}
