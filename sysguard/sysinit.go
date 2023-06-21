package sysguard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/toddlerya/glue/command"
	"github.com/toddlerya/glue/files"
)

var sysVinitTemplate = `#!/bin/bash
#
# Description: {{ .Description }}
#
# chkconfig: 345 99 01
#
# Source function library.
# . /etc/init.d/functions

# Some things that run always
BASEDIR={{ .WorkingDirectory }}
SERVICE_NAME="{{ .Name }}"
SERVICE_CMD='{{ .ExecStart }}'

kill_process() {
  local PPROCESS_CMD=$1
  ps -ef | grep "${PPROCESS_CMD}" | grep -v 'grep' | awk -F ' ' '{print $2}' | xargs kill -9
}

function check_process_exists() {
    # 获取动态参数 process_name
    local PPROCESS_CMD=$1
    # 使用 ps 命令查询包含该参数的进程状态, 并使用 grep 命令过滤出匹配的进程数量
    local process_count=$(ps -ef | grep "$PPROCESS_CMD" | grep -v grep | wc -l)
    # 判断进程数量是否大于0, 如果大于0则表示进程存在, 返回true, 否则返回false
    if [ $process_count -gt 0 ]; then
        return 0 # true
    else
        return 1 # false
    fi
}

function start() {
  cd $BASEDIR
  $SERVICE_CMD >/dev/null 2>&1 &
  if check_process_exists $SERVICE_CMD; then
    echo "start $SERVICE_NAME ok"
    exit 0
  else
    echo "start $SERVICE_NAME failed"
    exit 1
  fi
}

function status() {
  if check_process_exists $SERVICE_CMD; then
    echo "$SERVICE_NAME is running"
  else
    echo "$SERVICE_NAME is not running"
  fi
}

function stop() {
  kill_process $SERVICE_CMD
    if check_process_exists $SERVICE_CMD; then
      # echo "$SERVICE_NAME is running"
      echo "stop $SERVICE_NAME failed"
      exit 1
    else
      # echo "$SERVICE_NAME is not running"
      exit 0
    fi
}

# Carry out specific functions when asked to by the system
case "$1" in
  start)
    echo "Starting $SERVICE_NAME"
    start
    ;;
  status)
    echo "Status of $SERVICE_NAME"
    status
    ;;
  stop)
    echo "Stopping $SERVICE_NAME"
    stop
    ;;
  restart)
    echo "Restarting $SERVICE_NAME"
    stop
    start
    ;;
  *)
    echo "Usage: /etc/init.d/$SERVICE_NAME {start|status|stop|restart}"
    exit 1
    ;;
esac

exit 0
`

/*
生成SysVinit的启动脚本
1. 写入/etc/init.d/xxxxx
2. chmod 755 /etc/init.d/xxxxx
*/
func GenSysVinitServiceScript(systemdServiceConfig SystemdServiceConfig) error {
	tmpl, err := template.New(systemdServiceConfig.Name).Parse(sysVinitTemplate)
	if err != nil {
		return err
	}
	serviceScriptFilePath := filepath.Join(ROOT_MODE_SYSVINIT_SCRIPT_PATH, systemdServiceConfig.Name)
	f, err := os.Create(serviceScriptFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = tmpl.Execute(f, systemdServiceConfig)
	if err != nil {
		return err
	}
	// 对输出的文件增加可执行权限
	err = os.Chmod(serviceScriptFilePath, 0755)
	return err
}

/*
SysVinit模式部署
1. 渲染生成SysVinit脚本文件
2. chkconfig --add myservice           --- 注册服务
3. chkconfig --level 345 myservice on  --- 设置开机启动
4. service myservice start             --- 启动服务
5. service myservice status            --- 检查服务启动状态
*/
func SetupSysVinitService(systemdServiceConfig SystemdServiceConfig) error {
	// 渲染生成service文件
	err := GenSysVinitServiceScript(systemdServiceConfig)
	if err != nil {
		return err
	}
	// 将服务注册到SysVinit的自动启动列表
	addStdout, addStderr, err := command.RunByBash("chkconfig add", "chkconfig --add "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if addStdout != "" || addStderr != "" {
		return fmt.Errorf("%s注册到SysVinit的启动列表! stdout: %s stderr: %s", systemdServiceConfig.Name, addStdout, addStderr)
	}

	// 将服务注册到SysVinit的自动启动列表
	onStdout, onStderr, err := command.RunByBash("chkconfig on", "chkconfig --level 345 "+systemdServiceConfig.Name+" on")
	if err != nil {
		return err
	}
	if onStdout != "" || onStderr != "" {
		return fmt.Errorf("%s设置为345运行级别开机自启动失败! stdout: %s stderr: %s", systemdServiceConfig.Name, onStdout, onStderr)
	}

	// 启动服务
	startStdout, startStderr, err := command.RunByBash("service start", "service "+systemdServiceConfig.Name+" start")
	if err != nil {
		return err
	}
	if !strings.HasSuffix(strings.TrimSpace(startStdout), "ok") || startStderr != "" {
		return fmt.Errorf("%s启动失败! stdout: %s stderr: %s", systemdServiceConfig.Name, startStdout, startStderr)
	}

	// 检查启动状态
	statusStdout, statusStderr, err := command.RunByBash("service status", "service "+systemdServiceConfig.Name+" status")
	if err != nil {
		return err
	}
	if !strings.HasSuffix(strings.TrimSpace(statusStdout), "running") || statusStderr != "" {
		return fmt.Errorf("检查%s运行状态失败! stdout: %s stderr: %s", systemdServiceConfig.Name, statusStdout, statusStderr)
	}
	return nil
}

/*
SysVinit模式卸载
1. service myservice stop                --- 停止服务
2. service myservice status              --- 检查停止是否成功
3. chkconfig --level 345 myservice off   --- 取消开机自启
4. chkconfig --del myservice             --- 注销服务
5. rm -rf /etc/init.d/myservice          --- 删除服务脚本
*/
func UnSetupSysVinitService(systemdServiceConfig SystemdServiceConfig, deleteExporterWorkingDirectory bool) error {
	// 停止服务
	stopStdout, stopStderr, err := command.RunByBash("service stop", "service "+systemdServiceConfig.Name+" stop")
	// Unit xxxxx.service could not be found，这种情况 exit status 4
	if err != nil && !(strings.HasPrefix(err.Error(), "等待命令执行结束失败") && strings.HasSuffix(err.Error(), "exit status 4")) {
		return err
	}
	if strings.TrimSpace(stopStdout) != "Stopping "+systemdServiceConfig.Name || stopStderr != "" {
		return fmt.Errorf("%s停止失败! stdout: %s stderr: %s", systemdServiceConfig.Name, stopStdout, stopStderr)
	}
	// 检查停止状态
	statusStdout, statusStderr, err := command.RunByBash("service status", "service "+systemdServiceConfig.Name+" status")
	if err != nil && !(strings.HasPrefix(err.Error(), "等待命令执行结束失败") && strings.HasSuffix(err.Error(), "exit status 3")) {
		return err
	}
	if !strings.HasSuffix(strings.TrimSpace(statusStdout), "not running") || statusStderr != "" {
		return fmt.Errorf("检查%s运行状态失败! stdout: %s stderr: %s", systemdServiceConfig.Name, statusStdout, statusStderr)
	}
	// 取消开机自启
	offStdout, offStderr, err := command.RunByBash("chkconfig off", "chkconfig --level 345 "+systemdServiceConfig.Name+" off")
	if err != nil {
		return err
	}
	if offStdout != "" || offStderr != "" {
		return fmt.Errorf("%s关闭开机自启动失败! stdout: %s stderr: %s", systemdServiceConfig.Name, offStdout, offStderr)
	}
	// 将服务从SysVinit的自动启动列表删除
	delStdout, delStderr, err := command.RunByBash("chkconfig del", "chkconfig --del "+systemdServiceConfig.Name)
	if err != nil {
		return err
	}
	if delStdout != "" || delStderr != "" {
		return fmt.Errorf("%s注销SysVinit的启动列表! stdout: %s stderr: %s", systemdServiceConfig.Name, delStdout, delStderr)
	}
	// 删除服务的systemd配置
	if err = files.Delete(filepath.Join(ROOT_MODE_SYSVINIT_SCRIPT_PATH, systemdServiceConfig.Name)); err != nil {
		return fmt.Errorf("%s SysVinit脚本清理失败! ERROR: %s", systemdServiceConfig.Name, err.Error())
	}
	// 删除服务文件和配置
	if deleteExporterWorkingDirectory {
		if err = files.Delete(systemdServiceConfig.WorkingDirectory); err != nil {
			return fmt.Errorf("%s工作目录清理失败! ERROR: %s", systemdServiceConfig.Name, err.Error())
		}
	}
	return nil
}
