package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/fatih/color"
)

type Color int

const (
	NoColor Color = iota
	Raw
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

type stream struct {
	writeMu       sync.Mutex
	captureMu     sync.Mutex
	writer        io.Writer
	disableColors bool
}

var stderr = stream{
	writer:        os.Stderr,
	disableColors: color.NoColor,
}

var colorNameMap = map[string]Color{
	"none":    NoColor,
	"raw":     Raw,
	"red":     Red,
	"green":   Green,
	"yellow":  Yellow,
	"blue":    Blue,
	"magenta": Magenta,
	"cyan":    Cyan,
	"white":   White,
}

var colorMap = map[Color]*color.Color{
	NoColor: nil, Raw: nil,
	Red:     color.New(color.FgRed),
	Green:   color.New(color.FgGreen),
	Yellow:  color.New(color.FgYellow),
	Blue:    color.New(color.FgBlue),
	Magenta: color.New(color.FgMagenta),
	Cyan:    color.New(color.FgCyan),
	White:   color.New(color.FgWhite),
}

func ColorFromName(name string) Color {
	if clr, ok := colorNameMap[name]; ok {
		return clr
	}

	return White
}

func CaptureStderr(fn func()) string {
	return stderr.capture(fn)
}

func DisableColors(disable bool) {
	stderr.writeMu.Lock()
	defer stderr.writeMu.Unlock()

	stderr.disableColors = disable
}

func ColorsDisabled() bool {
	stderr.writeMu.Lock()
	defer stderr.writeMu.Unlock()

	return stderr.disableColors
}

func Stderrf(format string, args ...any) {
	stderr.write(func(w io.Writer) {
		_, _ = fmt.Fprintf(w, format, args...)
	})
}

func StderrColorf(c Color, format string, args ...any) {
	if ColorsDisabled() {
		Stderrf(format, args...)
		return
	}

	outputFunc := fmt.Fprintf

	stderr.write(func(w io.Writer) {
		if clr, ok := colorMap[c]; ok && clr != nil {
			outputFunc = clr.Fprintf
		}

		outputFunc(w, format, args...)
	})
}

func Stderrln(args ...any) {
	stderr.write(func(w io.Writer) {
		_, _ = fmt.Fprintln(w, args...)
	})
}

func (s *stream) capture(fn func()) string {
	s.captureMu.Lock()
	defer s.captureMu.Unlock()

	var buf bytes.Buffer

	s.writeMu.Lock()
	old := s.writer
	s.writer = &buf
	s.writeMu.Unlock()

	defer func() {
		s.writeMu.Lock()
		s.writer = old
		s.writeMu.Unlock()
	}()

	fn()

	return buf.String()
}

func (s *stream) write(fn func(io.Writer)) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	fn(s.writer)
}
