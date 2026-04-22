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
