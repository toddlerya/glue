package ps

import (
	"strings"

	"github.com/shirou/gopsutil/process"
)

type ProcessInfo struct {
	Name    string `json:"name"`
	Cwd     string `json:"cwd"`
	Cmdline string `json:"cmdlineArgs"`
	Exc     string `json:"exec"`
}

// 根据进程匹配字符获取进程信息
func GetProcessInfoByName(processCmdlineMatch string) ([]ProcessInfo, error) {
	var matchedProcessInfoArray []ProcessInfo
	allPids, err := process.Pids()
	if err != nil {
		return matchedProcessInfoArray, err
	} else {
		for _, pid := range allPids {
			proocessInfo, err := GetProcessInfoByPid(pid)
			if err != nil {
				return matchedProcessInfoArray, err
			} else {
				Cmdline := proocessInfo.Cmdline
				if strings.Contains(Cmdline, processCmdlineMatch) {
					matchedProcessInfoArray = append(matchedProcessInfoArray, proocessInfo)
				}
			}
		}
	}
	return matchedProcessInfoArray, err
}

// 根据给定的进程号获取进程信息
func GetProcessInfoByPid(pid int32) (ProcessInfo, error) {
	processInfo := ProcessInfo{}
	proc, err := process.NewProcess(pid)
	if err != nil {
		return processInfo, err
	} else {
		// 可以考虑这一串分支判断改为switch
		name, err := proc.Name()
		if err != nil {
			return processInfo, err
		} else {
			processInfo.Name = name
		}
		cwd, err := proc.Cwd()
		if err != nil {
			return processInfo, err
		} else {
			processInfo.Cwd = cwd
		}
		cmdline, err := proc.Cmdline()
		if err != nil {
			return processInfo, err
		} else {
			processInfo.Cmdline = cmdline
		}
		exc, err := proc.Exe()
		if err != nil {
			return processInfo, err
		} else {
			processInfo.Exc = exc
		}
	}
	return processInfo, err
}
