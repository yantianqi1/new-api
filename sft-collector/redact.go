package main

import (
	"regexp"
)

var redactPatterns = []struct {
	re   *regexp.Regexp
	repl string
}{
	{regexp.MustCompile(`sk-[A-Za-z0-9_-]{8,}`), "[REDACTED_KEY]"},
	{regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._\-+=/]{8,}`), "Bearer [REDACTED_TOKEN]"},
	{regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`), "[REDACTED_EMAIL]"},
	{regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]+?-----END [A-Z ]*PRIVATE KEY-----`), "[REDACTED_PRIVATE_KEY]"},
}

func redactText(s string) string {
	if s == "" {
		return s
	}
	for _, p := range redactPatterns {
		s = p.re.ReplaceAllString(s, p.repl)
	}
	return s
}

func redactRecord(rec ShareGPTRecord) ShareGPTRecord {
	rec.System = redactText(rec.System)
	rec.Tools = redactText(rec.Tools)
	if rec.Conversations != nil {
		out := make([]ShareGPTTurn, len(rec.Conversations))
		for i, turn := range rec.Conversations {
			out[i] = ShareGPTTurn{From: turn.From, Value: redactText(turn.Value)}
		}
		rec.Conversations = out
	}
	return rec
}
