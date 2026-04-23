package logging

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
)

var redactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`sk-[A-Za-z0-9_-]+`),
	regexp.MustCompile(`(?i)authorization:\s*bearer\s+[^\s]+`),
	regexp.MustCompile(`(?i)cookie:\s*[^\n\r]+`),
	regexp.MustCompile(`(?i)(access_token|refresh_token|openai_api_key)\s*[:=]\s*[^\s"']+`),
	regexp.MustCompile(`[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]{10,}`),
}

type Logger struct {
	verbose bool
	out     io.Writer
	err     io.Writer
	mu      sync.Mutex
}

func New(verbose bool) *Logger {
	return &Logger{verbose: verbose, out: os.Stdout, err: os.Stderr}
}

func RedactSecrets(input string) string {
	masked := input
	for _, pattern := range redactPatterns {
		masked = pattern.ReplaceAllStringFunc(masked, func(value string) string {
			parts := strings.SplitN(value, ":", 2)
			if len(parts) == 2 && strings.Contains(value, ":") {
				return parts[0] + ": [REDACTED]"
			}
			return "[REDACTED]"
		})
	}
	return masked
}

type RedactingWriter struct {
	target io.Writer
	buffer strings.Builder
	mu     sync.Mutex
}

func NewRedactingWriter(target io.Writer) *RedactingWriter {
	return &RedactingWriter{target: target}
}

func (w *RedactingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Write(p)
	return len(p), w.flushCompleteLines()
}

func (w *RedactingWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffer.Len() == 0 {
		return nil
	}
	_, err := io.WriteString(w.target, RedactSecrets(w.buffer.String()))
	w.buffer.Reset()
	return err
}

func (w *RedactingWriter) flushCompleteLines() error {
	buffer := w.buffer.String()
	lastNewline := strings.LastIndexByte(buffer, '\n')
	if lastNewline == -1 {
		return nil
	}
	complete := buffer[:lastNewline+1]
	remaining := buffer[lastNewline+1:]
	if _, err := io.WriteString(w.target, RedactSecrets(complete)); err != nil {
		return err
	}
	w.buffer.Reset()
	w.buffer.WriteString(remaining)
	return nil
}

func (l *Logger) Infof(format string, args ...any) {
	l.write(l.out, format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.write(l.err, format, args...)
}

func (l *Logger) Verbosef(format string, args ...any) {
	if !l.verbose {
		return
	}
	l.write(l.out, format, args...)
}

func (l *Logger) write(w io.Writer, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = fmt.Fprintln(w, RedactSecrets(fmt.Sprintf(format, args...)))
}
