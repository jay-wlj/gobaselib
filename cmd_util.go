package gobaselib

import (
	"errors"
	// "fmt"
	"os/exec"
	"time"
)

var ErrTimeout = errors.New("timeout_cmd")

func CombinedOutput(cmd *exec.Cmd, timeout time.Duration) ([]byte, error) {
	var timer *time.Timer
	timer = time.AfterFunc(timeout, func() {
		// timer.Stop()
		cmd.Process.Kill()
	})
	defer timer.Stop()
	output, err := cmd.CombinedOutput()

	if err != nil && err.Error() == "signal: killed" {
		err = ErrTimeout
	}
	return output, err
}
