package system

import (
	"bytes"
	"fmt"
	"os/exec"
	"unicode/utf8"

	"github.com/duke-git/lancet/v2/validator"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// ref: https://github.com/duke-git/lancet/blob/main/system/os.go

type (
	Option func(*exec.Cmd)
)

// ExecCommand execute command, return the stdout and stderr string and exitCode of command, and error if error occur
// param `command` is a complete command string, like, ls -a (linux), dir(windows), ping 127.0.0.1
// in linux,  use /bin/bash -c to execute command
// in windows, use powershell.exe to execute command
// Play: https://go.dev/play/p/n-2fLyZef-4
func ExecCommand(command string, opts ...Option) (stdout, stderr string, exitCode int, err error) {
	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer

	cmd := exec.Command("/bin/bash", "-c", command)
	if IsWindows() {
		cmd = exec.Command("powershell.exe", command)
	}

	logrus.Debugf("exec.cmd: %s\n", cmd.String())

	for _, opt := range opts {
		if opt != nil {
			opt(cmd)
		}
	}
	cmd.Stdout = &stdOutBuf
	cmd.Stderr = &stdErrBuf
	fmt.Println("cmd.String()==>", cmd.String())

	err = cmd.Run()
	exitCode = cmd.ProcessState.ExitCode()

	if err != nil {
		if utf8.Valid(stdErrBuf.Bytes()) {
			stderr = byteToString(stdErrBuf.Bytes(), "UTF8")
		} else if validator.IsGBK(stdErrBuf.Bytes()) {
			stderr = byteToString(stdErrBuf.Bytes(), "GBK")
		}
		return
	}

	data := stdOutBuf.Bytes()
	if utf8.Valid(data) {
		stdout = byteToString(data, "UTF8")
	} else if validator.IsGBK(data) {
		stdout = byteToString(data, "GBK")
	}

	return
}

func byteToString(data []byte, charset string) string {
	var result string

	switch charset {
	case "GBK":
		decodeBytes, _ := simplifiedchinese.GBK.NewDecoder().Bytes(data)
		result = string(decodeBytes)
	case "GB18030":
		decodeBytes, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(data)
		result = string(decodeBytes)
	case "UTF8":
		fallthrough
	default:
		result = string(data)
	}

	return result
}
