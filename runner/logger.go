package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/air-verse/air/runner/output"
)

type logFunc func(string, ...interface{})

type logger struct {
	config  *Config
	loggers map[string]logFunc
}

func newLogger(cfg *Config) *logger {
	if cfg == nil {
		return nil
	}

	colors := cfg.colorInfo()
	loggers := make(map[string]logFunc, len(colors))
	for name, nameColor := range colors {
		loggers[name] = newLogFunc(output.ColorFromName(nameColor), cfg.Log)
	}

	loggers["default"] = defaultLogger()

	return &logger{
		config:  cfg,
		loggers: loggers,
	}
}

func newLogFunc(c output.Color, cfg cfgLog) logFunc {
	return func(msg string, v ...interface{}) {
		// There are some escape sequences to format color in terminal, so cannot
		// just trim new line from right.
		if cfg.Silent {
			return
		}
		msg = strings.ReplaceAll(msg, "\n", "")
		msg = strings.TrimSpace(msg)
		if len(msg) == 0 {
			return
		}
		// TODO: filter msg by regex
		msg = msg + "\n"
		if cfg.AddTime {
			t := time.Now().Format("15:04:05")
			msg = fmt.Sprintf("[%s] %s", t, msg)
		}

		output.StderrColorf(c, msg, v...)
	}
}

func (l *logger) main() logFunc {
	return l.getLogger("main")
}

func (l *logger) build() logFunc {
	return l.getLogger("build")
}

func (l *logger) runner() logFunc {
	return l.getLogger("runner")
}

func (l *logger) watcher() logFunc {
	return l.getLogger("watcher")
}

func rawLogger() logFunc {
	return newLogFunc(output.Raw, defaultConfig().Log)
}

func defaultLogger() logFunc {
	return newLogFunc(output.White, defaultConfig().Log)
}

func (l *logger) getLogger(name string) logFunc {
	v, ok := l.loggers[name]
	if !ok {
		return rawLogger()
	}
	return v
}
