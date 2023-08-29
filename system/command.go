package system

import (
	"bytes"
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

	err = cmd.Run()

	exitCode = cmd.ProcessState.ExitCode()
	stdErrData := stdErrBuf.Bytes()
	if utf8.Valid(stdErrData) {
		stderr = byteToString(stdErrData, "UTF8")
	} else if validator.IsGBK(stdErrData) {
		stderr = byteToString(stdErrData, "GBK")
	}

	stdOutData := stdOutBuf.Bytes()
	if utf8.Valid(stdOutData) {
		stdout = byteToString(stdOutData, "UTF8")
	} else if validator.IsGBK(stdOutData) {
		stdout = byteToString(stdOutData, "GBK")
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
