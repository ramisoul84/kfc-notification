package emailprovider

import (
	"context"
	"fmt"
	"html"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ramisoul84/kfc-notification/pkg/logger"
)

const (
	// Directory where dev emails are written.
	logEmailDir = "tmp/emails"
)

// logProvider does not actually send email. Instead it:
//   - logs a one-line summary
//   - writes the rendered HTML to disk for inspection in a browser
//
// Intended for local development only.
type logProvider struct {
	log *logger.Logger
}

func newLog(log *logger.Logger) *logProvider {
	if err := os.MkdirAll(logEmailDir, 0o755); err != nil {
		log.Warn("could not create dev email directory", "dir", logEmailDir, "error", err)
	}
	return &logProvider{log: log}
}

func (p *logProvider) Send(ctx context.Context, to, subject, htmlBody string) error {
	// 1. Summary log line — always emitted.
	p.log.Info("email (log provider)",
		"to", to,
		"subject", subject,
		"body_length", len(htmlBody),
	)

	// 2. Write the HTML to disk so it can be opened in a browser.
	path, err := p.writeToFile(to, subject, htmlBody)
	if err != nil {
		p.log.Warn("could not write dev email file", "error", err)
		return nil // not fatal — the email is still "sent" for dev purposes
	}

	p.log.Info("email saved", "path", path)
	return nil
}

// writeToFile saves the HTML body to a deterministic path.
// Filename: <timestamp>_<safe-to>_<safe-subject>.html
func (p *logProvider) writeToFile(to, subject, htmlBody string) (string, error) {
	if err := os.MkdirAll(logEmailDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", logEmailDir, err)
	}

	ts := time.Now().Format("20060102-150405")
	name := fmt.Sprintf("%s_%s_%s.html",
		ts,
		sanitize(to),
		sanitize(subject),
	)
	path := filepath.Join(logEmailDir, name)

	// Wrap in a minimal preview header so you know who it was for.
	page := template.HTML(fmt.Sprintf(
		`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>%s</title></head>
<body style="margin:0;font-family:sans-serif;">
<div style="background:#eee;padding:8px 12px;font-size:12px;border-bottom:1px solid #ccc;">
  <strong>To:</strong> %s &nbsp;|&nbsp; <strong>Subject:</strong> %s &nbsp;|&nbsp; <strong>Saved:</strong> %s
</div>
%s
</body></html>`,
		htmlEscape(subject),
		htmlEscape(to),
		htmlEscape(subject),
		time.Now().Format(time.RFC1123),
		htmlBody,
	))

	if err := os.WriteFile(path, []byte(page), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return path, nil
}

// sanitize produces a filename-safe version of s.
func sanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// keep letters, digits, @, dot, dash
	re := regexp.MustCompile(`[^a-z0-9@._-]+`)
	s = re.ReplaceAllString(s, "_")
	if len(s) > 60 {
		s = s[:60]
	}
	if s == "" {
		s = "email"
	}
	return s
}

// htmlEscape escapes a string for safe insertion into HTML.
func htmlEscape(s string) string {
	return html.EscapeString(s)
}
