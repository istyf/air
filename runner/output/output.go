package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

var stderr = stream{
	writer: os.Stderr,
}

type stream struct {
	writeMu   sync.Mutex
	captureMu sync.Mutex
	writer    io.Writer
}

func CaptureStderr(fn func()) string {
	return stderr.capture(fn)
}

func Stderrf(format string, args ...any) {
	stderr.write(func(w io.Writer) {
		_, _ = fmt.Fprintf(w, format, args...)
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
