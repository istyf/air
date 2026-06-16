package output

import "io"

var stdoutWriter = streamWriter{stream: &stdout}
var stderrWriter = streamWriter{stream: &stderr}

type streamWriter struct {
	stream *stream
}

func StdoutWriter() io.Writer {
	return stdoutWriter
}

func StderrWriter() io.Writer {
	return stderrWriter
}

func (w streamWriter) Write(p []byte) (int, error) {
	var err error

	w.stream.write(func(dst io.Writer) {
		_, err = dst.Write(p)
	})

	if err != nil {
		return 0, err
	}
	return len(p), nil
}
