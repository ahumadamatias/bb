// Package iostreams provides a minimal abstraction over stdin/stdout/stderr
// with TTY and color detection, similar in spirit to cli/cli's iostreams.
package iostreams

import (
	"io"
	"os"

	"golang.org/x/term"
)

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer

	stdinFd  uintptr
	stdoutFd uintptr

	isStdinTTY  bool
	isStdoutTTY bool

	colorEnabled bool
}

func System() *IOStreams {
	stdout := os.Stdout
	stdin := os.Stdin

	io := &IOStreams{
		In:     stdin,
		Out:    stdout,
		ErrOut: os.Stderr,

		stdinFd:  stdin.Fd(),
		stdoutFd: stdout.Fd(),
	}

	io.isStdinTTY = term.IsTerminal(int(io.stdinFd))
	io.isStdoutTTY = term.IsTerminal(int(io.stdoutFd))
	io.colorEnabled = io.isStdoutTTY && os.Getenv("NO_COLOR") == ""

	return io
}

func (s *IOStreams) IsStdoutTTY() bool {
	return s.isStdoutTTY
}

func (s *IOStreams) IsStdinTTY() bool {
	return s.isStdinTTY
}

func (s *IOStreams) ColorEnabled() bool {
	return s.colorEnabled
}

// ReadPassword reads a line from stdin without echoing it, when stdin is a TTY.
// Falls back to a plain line read otherwise (e.g. piped input in scripts/tests).
func (s *IOStreams) ReadPassword() (string, error) {
	if s.isStdinTTY {
		b, err := term.ReadPassword(int(s.stdinFd))
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	var line string
	_, err := fscanLine(s.In, &line)
	return line, err
}

func fscanLine(r io.Reader, out *string) (int, error) {
	buf := make([]byte, 0, 128)
	b := make([]byte, 1)
	for {
		n, err := r.Read(b)
		if n > 0 {
			if b[0] == '\n' {
				break
			}
			buf = append(buf, b[0])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return len(buf), err
		}
	}
	*out = string(buf)
	return len(buf), nil
}

const (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorCyan   = "\x1b[36m"
	colorBold   = "\x1b[1m"
)

func (s *IOStreams) colorize(code, text string) string {
	if !s.colorEnabled {
		return text
	}
	return code + text + colorReset
}

func (s *IOStreams) Red(text string) string    { return s.colorize(colorRed, text) }
func (s *IOStreams) Green(text string) string  { return s.colorize(colorGreen, text) }
func (s *IOStreams) Yellow(text string) string { return s.colorize(colorYellow, text) }
func (s *IOStreams) Cyan(text string) string   { return s.colorize(colorCyan, text) }
func (s *IOStreams) Bold(text string) string   { return s.colorize(colorBold, text) }
