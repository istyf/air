package runner

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

func (e *Engine) killCmd(cmd *exec.Cmd) (pid int, err error) {
	pid = cmd.Process.Pid

	var killDelay time.Duration

	waitResult := make(chan error)
	go func() {
		defer close(waitResult)
		_, _ = cmd.Process.Wait()
	}()

	if e.config.Build.SendInterrupt {
		e.mainDebug("sending interrupt to process %d", pid)
		// Sending a signal to make it clear to the process that it is time to turn off
		if err = cmd.Process.Signal(os.Interrupt); err != nil {
			return
		}
		// the kill delay is 0 by default unless the user has configured send_interrupt=true
		// in which case it is fetched from the kill_delay setting in the .air.toml
		killDelay = e.config.killDelay()
		e.mainDebug("setting a kill timer for %s", killDelay.String())
	}

	// prepare a cancel context that can stop the killing if it is not needed
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	killResult := make(chan error)
	// Spawn a goroutine that will kill the process after kill delay if we have not
	// received a wait result before that.
	go func() {
		select {
		case <-time.After(killDelay):
			e.mainDebug("kill timer expired")
			killResult <- cmd.Process.Kill()
		case <-ctx.Done():
			e.mainDebug("kill timer canceled")
			return
		}
	}()

	results := make([]error, 0, 2)

	for {
		// collect the responses from the kill and wait goroutines
		select {
		case err = <-killResult:
			results = append(results, err)
		case err = <-waitResult:
			results = append(results, err)
			// if we have a kill delay, we ignore the kill result
			if killDelay > 0 && len(results) == 1 {
				results = append(results, nil)
			}
		}

		if len(results) == 2 {
			err = errors.Join(results...)
			return
		}
	}
}

func (e *Engine) startCmd(cmd string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	c := exec.Command("/bin/sh", "-c", cmd)
	f, err := pty.Start(c)
	return c, f, f, err
}
