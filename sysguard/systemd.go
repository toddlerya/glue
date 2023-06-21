package sysguard

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

// 定义systemd配置所需信息结构体
type SystemdServiceConfig struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	WorkingDirectory string `json:"working_directory"`
	ExecStart        string `json:"exec_start"`
}

// systemd service config template
var serviceTemplate = `
[Unit]
Description={{ .Description }}
ConditionPathExists={{ .WorkingDirectory }}
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
WorkingDirectory={{ .WorkingDirectory }}
ExecStart={{ .ExecStart }}
Restart=on-failure
RestartSec=5s
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier={{ .Name }}
LimitNOFILE=10000
TimeoutStopSec=20

[Install]
WantedBy=multi-user.target
`

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
systemd模式部署
1. 渲染生成service文件
2. systemctl --user daemon-reload  -- 加载配置
3. systemctl --user enable appName -- 设为开机启动
4. systemctl --user start appName  -- 启动服务
5. systemctl --user status appName -- 检查是否启动成功
*/
func SetupSystemdService(systemdServiceConfig SystemdServiceConfig) error {

	err := PreSetupSystemdService()
	if err != nil {
		return fmt.Errorf("创建%s目录失败: %s", SYSTEMD_SERVICE_PATH, err.Error())
	}

	// 渲染生成service文件
	err = GenSystemdServiceConfigFile(systemdServiceConfig)
	if err != nil {
		return err
	}
	// 加载配置
	reloadStdout, reloadStderr, err := command.RunByBash("systemd daemon-reload", "systemctl "+SYSTEMCTL_MODE+" daemon-reload")
	if err != nil {
		return err
	}
	if reloadStdout != "" || reloadStderr != "" {
		return fmt.Errorf("加载%s systemd配置失败! stdout: %s stderr: %s", systemdServiceConfig.Name, reloadStdout, reloadStderr)
	}
	// 设为开机启动
	enableStdout, enableStderr, err := command.RunByBash("systemd enable", "systemctl "+SYSTEMCTL_MODE+" enable "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if enableStdout != "" || !strings.HasPrefix(enableStderr, "Created symlink") {
		return fmt.Errorf("%s设为开机启动失败! stdout: %s stderr: %s", systemdServiceConfig.Name, enableStdout, enableStderr)
	}
	// 启动服务
	startStdout, startStderr, err := command.RunByBash("systemd start", "systemctl "+SYSTEMCTL_MODE+" start "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if startStdout != "" || startStderr != "" {
		return fmt.Errorf("%s启动失败! stdout: %s stderr: %s", systemdServiceConfig.Name, startStdout, startStderr)
	}
	// 检查服务启动状态
	statusStdout, statusStderr, err := command.RunByBash("systemd status", "systemctl "+SYSTEMCTL_MODE+" status "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if !strings.Contains(statusStdout, "Active: active (running)") || statusStderr != "" {
		return fmt.Errorf("检查%s运行状态失败! stdout: %s stderr: %s", systemdServiceConfig.Name, statusStdout, statusStderr)
	}
	return nil
}

/*
systemd模式卸载清理
1. systemctl --user stop appName -- 停止服务
2. systemctl --user disable appName -- 禁用服务
3. rm ~/.config/systemd/user/appName.service -- 删除服务的systemd配置
4. rm -r WorkingDirectory -- 删除服务文件和配置 【非必须】
*/
func UnSetupSystemService(systemdServiceConfig SystemdServiceConfig, deleteExporterWorkingDirectory bool) error {
	// 停止服务
	stopStdout, stopStderr, err := command.RunByBash("systemd stop", "systemctl "+SYSTEMCTL_MODE+" stop "+systemdServiceConfig.Name)
	// Failed to stop xxxxx.service: Unit xxxx.service not loaded这种情况会exit status 5
	if err != nil && !(strings.HasPrefix(err.Error(), "等待命令执行结束失败") && strings.HasSuffix(err.Error(), "exit status 5")) {
		return err
	}
	if strings.TrimSpace(stopStdout) != "" || strings.TrimSpace(stopStderr) != "" {
		// 如果没有注册过systemd服务, 则直接跳过就好了, 虽然不应该发生这种事情, 但谁知道实际运行时什么鬼情况
		if !strings.Contains(stopStderr, "not loaded.") {
			return fmt.Errorf("%s停止失败! stdout: %s stderr: %s", systemdServiceConfig.Name, stopStdout, stopStderr)
		}
	} else {
		// 检测服务停止状态, 需要确认进程是否停止成功
		statusStdout, statusStderr, err := command.RunByBash("systemd is-active", "systemctl "+SYSTEMCTL_MODE+" is-active "+systemdServiceConfig.Name)
		if err != nil && !(strings.HasPrefix(err.Error(), "等待命令执行结束失败") && strings.HasSuffix(err.Error(), "exit status 3")) {
			return err
		}
		if strings.TrimSpace(statusStdout) != "inactive" || statusStderr != "" {
			return fmt.Errorf("检查%s运行状态失败! stdout: %s stderr: %s", systemdServiceConfig.Name, statusStdout, statusStderr)
		}
		// 禁用服务
		disableStdout, disableStderr, err := command.RunByBash("systemd disable", "systemctl "+SYSTEMCTL_MODE+" disable "+systemdServiceConfig.Name)
		if err != nil {
			return err
		}
		if disableStdout != "" || (disableStderr != "" && !strings.HasPrefix(disableStderr, "Removed")) {
			return fmt.Errorf("%s禁用失败! stdout: %s stderr: %s", systemdServiceConfig.Name, disableStdout, disableStderr)
		}
	}
	// 删除服务的systemd配置
	if err = files.Delete(filepath.Join(SYSTEMD_SERVICE_PATH, systemdServiceConfig.Name+".service")); err != nil {
		return fmt.Errorf("%s守护配置清理失败! ERROR: %s", systemdServiceConfig.Name, err.Error())
	}
	// 删除服务文件和配置
	if deleteExporterWorkingDirectory {
		if err = files.Delete(systemdServiceConfig.WorkingDirectory); err != nil {
			return fmt.Errorf("%s工作目录清理失败! ERROR: %s", systemdServiceConfig.Name, err.Error())
		}
	}
	return err
}
