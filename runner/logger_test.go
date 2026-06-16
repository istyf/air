package runner

import (
	"strings"
	"testing"

	"github.com/air-verse/air/runner/output"
)

func TestLogFuncWritesToStderr(t *testing.T) {
	t.Parallel()

	const LogMessage string = "test message from air"

	_, errorOutput := output.CaptureStderr(func() {
		logFn := newLogFunc(output.Raw, cfgLog{})
		logFn(LogMessage)
	})

	if !strings.Contains(errorOutput, LogMessage) {
		t.Errorf("expected log output on stderr, got: %s", errorOutput)
	}
}
