package sentry

import (
	"regexp"
	"strings"
)

var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(authorization|cookie|set-cookie|x-api-key)(["' :=]+)[^\r\n,;}]+`),
	regexp.MustCompile(`(?i)(token|password|secret)(["' :=]+)[^\s,;}]+`),
	regexp.MustCompile(`(?i)bearer\s+[a-z0-9._~+/=-]+`),
}

func sanitizeText(value string) string {
	sanitized := value
	for _, pattern := range sensitivePatterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, func(match string) string {
			if strings.HasPrefix(strings.ToLower(match), "bearer ") {
				return "Bearer [REDACTED]"
			}
			separator := strings.IndexAny(match, " :=\"'")
			if separator < 0 {
				return "[REDACTED]"
			}
			return match[:separator] + "=[REDACTED]"
		})
	}
	return sanitized
}
