package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func ExecCommand(command string, arg ...string) (result []byte, err error) {
	if !strings.Contains(command, string(os.PathSeparator)) {
		command, err = exec.LookPath(command)
		if err != nil {
			return
		}
	}
	var cmd = exec.Command(command, arg...)
	var stdout, stderr io.ReadCloser
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		err = errors.New(fmt.Sprintf("ExecCommand:%s %#v %s\n", command, arg, err))
		return
	}
	stderr, _ = cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		return
	}
	result, err = ioutil.ReadAll(stdout)
	if err == nil {
		var buffer []byte
		if buffer, err = ioutil.ReadAll(stderr); err == nil && len(buffer) > 0 {
			err = errors.New(string(buffer))
		}
	}
	if cmd != nil {
		_ = cmd.Wait()
	}
	return
}

var reCommandline = regexp.MustCompile(`"(.+)"\s*(.+)?`)

func ExecCommandString(command string) (result string, err error) {
	var parts = reCommandline.FindStringSubmatch(command)
	var args []string
	if len(parts) == 0 {
		parts = strings.Fields(command)
		command = parts[0]
		args = parts[1:]
	} else {
		command = parts[1]
		args = parts[2:]
	}
	var buffer []byte
	if buffer, err = ExecCommand(command, args...); err == nil {
		result = string(buffer)
	}
	return
}
