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

type ColorMode int

const (
	ColorAuto ColorMode = iota
	ColorAlways
	ColorNever
)

type stream struct {
	writeMu   sync.Mutex
	captureMu sync.Mutex
	writer    io.Writer
	colorMode ColorMode
}

var stderr = stream{
	writer:    os.Stderr,
	colorMode: ColorAuto,
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

var colorAttrs = map[Color]color.Attribute{
	Red:     color.FgRed,
	Green:   color.FgGreen,
	Yellow:  color.FgYellow,
	Blue:    color.FgBlue,
	Magenta: color.FgMagenta,
	Cyan:    color.FgCyan,
	White:   color.FgWhite,
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

func GetColorMode() ColorMode {
	stderr.writeMu.Lock()
	defer stderr.writeMu.Unlock()

	return stderr.colorMode
}

func SetColorMode(mode ColorMode) {
	stderr.writeMu.Lock()
	defer stderr.writeMu.Unlock()

	stderr.colorMode = mode
}

func SetColorModeFromString(mode string) error {
	theMode, ok := map[string]ColorMode{
		"":       ColorAuto,
		"always": ColorAlways,
		"auto":   ColorAuto,
		"never":  ColorNever,
	}[mode]

	if !ok {
		return fmt.Errorf("unsupported color mode: %s. Expected always, auto, or never", mode)
	}

	SetColorMode(theMode)

	return nil
}

func Stderrf(format string, args ...any) {
	stderr.write(func(w io.Writer) {
		_, _ = fmt.Fprintf(w, format, args...)
	})
}

func StderrColorf(c Color, format string, args ...any) {
	outputFunc := fmt.Fprintf

	var theColor *color.Color

	stderr.write(func(w io.Writer) {
		if attribute, ok := colorAttrs[c]; ok && stderr.colorMode != ColorNever {
			theColor = color.New(attribute)
			if stderr.colorMode == ColorAlways {
				theColor.EnableColor()
			}
			outputFunc = theColor.Fprintf
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
