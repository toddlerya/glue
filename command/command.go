package command

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os/exec"
	"syscall"
	"time"
)

// 执行一个shell命令，通过shell参数控制使用/bin/sh还是/bin/bash，获取其标准输出和标准错误
func runShell(shell string, cmd string) ([]byte, []byte, error) {
	var stdoutMsg, stderrMsg []byte
	cmdStuct := exec.Command(shell, "-c", cmd)
	stdout, err := cmdStuct.StdoutPipe()
	if err != nil {
		return stdoutMsg, stderrMsg, fmt.Errorf("创建stdout命令管道失败: CMD: %s ERROR: %s", cmd, err.Error())
	}
	defer stdout.Close()

	stderr, err := cmdStuct.StderrPipe()
	if err != nil {
		return stdoutMsg, stderrMsg, fmt.Errorf("创建stderr命令管道失败: CMD: %s ERROR: %s", cmd, err.Error())
	}
	defer stderr.Close()

	if err = cmdStuct.Start(); err != nil {
		return stdoutMsg, stderrMsg, fmt.Errorf("启动命令失败: CMD: %s ERROR: %s", cmd, err.Error())
	}

	stdoutMsg, err = ioutil.ReadAll(stdout)
	if err != nil {
		return stdoutMsg, stderrMsg, fmt.Errorf("获取命令执行标准输出失败: CMD: %s ERROR: %s", cmd, err.Error())
	}
	stderrMsg, err = ioutil.ReadAll(stderr)
	if err != nil {
		return stdoutMsg, stderrMsg, fmt.Errorf("获取命令标准错误失败: CMD: %s ERROR: %s", cmd, err.Error())
	}
	return stdoutMsg, stderrMsg, err
}

// 封装shell执行命令方法，提供友好的输出
func Run(tag, shell, cmd string) (string, string, error) {
	stdout, stderr, err := runShell(shell, cmd)
	// 不知道之前为什么写个了bash %s，加上这个会导致java命令无法运行...报错为: 无法执行二进制文件
	// stdout, stderr, err := runShell(shell, fmt.Sprintf("bash %s", cmd))
	return string(stdout), string(stderr), err
}

// 通过/bin/sh shell执行命令
func RunBySh(tag, cmd string) (string, string, error) {
	return Run(tag, "/bin/sh", cmd)
}

// 通过/bin/bash shell执行命令
func RunByBash(tag, cmd string) (string, string, error) {
	return Run(tag, "/bin/bash", cmd)
}

// 异步执行命令, 实时获取命令输出
func RunCmdStream(tag, shell, cmd string, stdoutChan, stderrChan, shutdownChan chan string) error {
	cmdStuct := exec.Command(shell, "-c", cmd)

	stdout, err := cmdStuct.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout命令管道失败: CMD: %s ERROR: %s", cmd, err.Error())
	}
	defer stdout.Close()

	stderr, err := cmdStuct.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr命令管道失败: CMD: %s ERROR: %s", cmd, err.Error())
	}
	defer stderr.Close()

	if err = cmdStuct.Start(); err != nil {
		return fmt.Errorf("启动命令失败: CMD: %s ERROR: %s", cmd, err.Error())
	}

	// 程序退出时kill子进程 Ref: https://colobu.com/2020/12/27/go-with-os-exec/
	cmdStuct.SysProcAttr = &syscall.SysProcAttr{Setpgid: false}

	// 异步启动命令
	err = cmdStuct.Start()
	if err != nil && err.Error() != "exec: already started" {
		return fmt.Errorf("异步启动命令失败: %s", err.Error())
	}

	// 获取实时标准输出
	go func() {
		stdoutScanner := bufio.NewScanner(stdout)
		// 实时循环读取流中的一行内容
		for stdoutScanner.Scan() {
			stdoutLine := stdoutScanner.Text()
			// fmt.Println(stdoutLine)
			stdoutChan <- stdoutLine
		}
		// 获取实时标准错误
		stderrScanner := bufio.NewScanner(stderr)
		for stderrScanner.Scan() {
			stderrLine := stderrScanner.Text()
			stderrChan <- stderrLine
		}
	}()

	// 等待接受退出信号
	go func() {
		for {
			shutdownSignal, ok := <-shutdownChan
			if !ok {
				break
			}
			if shutdownSignal == "exit" {
				// kill子进程
				syscall.Kill(cmdStuct.Process.Pid, syscall.SIGKILL)
			} else {
				time.Sleep(3000 * time.Millisecond)
			}
		}
	}()

	// 阻塞等待命令结束
	err = cmdStuct.Wait()
	if err != nil && err.Error() != "signal: killed" {
		return fmt.Errorf("阻塞异步命令任务失败: %s", err.Error())

	}
	return err
}
