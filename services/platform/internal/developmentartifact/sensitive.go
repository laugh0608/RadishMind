package developmentartifact

import (
	"regexp"
	"strings"
)

// This is a bounded deny detector, not a proof that arbitrary text is public.
// Source/patch are never rewritten. Logs omit entire affected lines.
var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(authorization\s*[:=]|bearer\s+|-----BEGIN [A-Z ]*PRIVATE KEY-----)`),
	regexp.MustCompile(`(?i)(api[_-]?key|access[_-]?token|password|secret|dsn)\s*["']?\s*[:=]\s*\S+`),
	regexp.MustCompile(`\b(sk-[A-Za-z0-9_-]{8,}|AKIA[A-Z0-9]{16})\b`),
	regexp.MustCompile(`(?i)(https?|postgres(ql)?|mysql)://[^\s/]+:[^\s/]+@`),
	regexp.MustCompile(`(?i)(/(Users|home|private|tmp|var|etc|root|opt)/|\b[A-Z]:\\|file://)`),
}

func sensitive(s string) bool {
	for _, p := range sensitivePatterns {
		if p.MatchString(s) {
			return true
		}
	}
	return false
}
func redactLog(s string) string {
	parts := strings.SplitAfter(s, "\n")
	for i, line := range parts {
		if sensitive(line) {
			parts[i] = "[redacted]"
			if strings.HasSuffix(line, "\n") {
				parts[i] += "\n"
			}
		}
	}
	return strings.Join(parts, "")
}
