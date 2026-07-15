package agent

import (
	"os/exec"
	"runtime"
	"strings"
)

// ShellExecutor handles executing commands in the system shell.
type ShellExecutor struct{}

// Execute runs a command in the system shell and returns the combined stdout and stderr.
func (s *ShellExecutor) Execute(command string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Even if there's an error, we want to return the output (which contains stderr)
		return strings.TrimSpace(string(output)), err
	}

	return strings.TrimSpace(string(output)), nil
}
