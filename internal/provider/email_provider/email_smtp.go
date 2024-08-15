package emailprovider

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/ramisoul84/kfc-notification/internal/config"
	"github.com/ramisoul84/kfc-notification/pkg/logger"
)

// smtpProvider sends email via SMTP.
//
// Supports:
//   - implicit TLS on port 465
//   - STARTTLS on port 587 (and any other port when SMTPUseTLS is true)
//   - unauthenticated relays (when SMTPUsername is empty)
//
// Context: the ctx is honored for the TCP dial and the TLS handshake.
// net/smtp itself has no context support, so per-command cancellation
// is bounded by the dial timeout instead.
type smtpProvider struct {
	cfg config.EmailConfig
	log *logger.Logger
}

func newSMTP(cfg config.EmailConfig, log *logger.Logger) *smtpProvider {
	return &smtpProvider{cfg: cfg, log: log}
}

func (p *smtpProvider) Send(ctx context.Context, to, subject, htmlBody string) error {
	addr := net.JoinHostPort(p.cfg.SMTPHost, fmt.Sprintf("%d", p.cfg.SMTPPort))

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial %s: %w", addr, err)
	}
	defer conn.Close()

	// Implicit TLS (port 465): wrap from the start.
	if p.cfg.SMTPUseTLS && p.cfg.SMTPPort == 465 {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: p.cfg.SMTPHost})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("smtp tls handshake: %w", err)
		}
		conn = tlsConn
	}

	client, err := smtp.NewClient(conn, p.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() {
		// QUIT failures are non-fatal; log for visibility.
		if err := client.Quit(); err != nil {
			p.log.Warn("smtp QUIT failed", "error", err)
		}
	}()

	// Identify ourselves. Some servers reject clients that skip EHLO.
	if err := client.Hello(p.localName()); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}

	// STARTTLS for 587-style connections.
	if p.cfg.SMTPUseTLS && p.cfg.SMTPPort != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: p.cfg.SMTPHost}); err != nil {
				return fmt.Errorf("smtp starttls: %w", err)
			}
		}
	}

	// Optional authentication.
	if p.cfg.SMTPUsername != "" {
		auth := smtp.PlainAuth("", p.cfg.SMTPUsername, p.cfg.SMTPPassword, p.cfg.SMTPHost)
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}
	}

	if err := client.Mail(p.cfg.From); err != nil {
		return fmt.Errorf("smtp MAIL FROM <%s>: %w", p.cfg.From, err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO <%s>: %w", to, err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}

	msg := p.buildMessage(to, subject, htmlBody)
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close body: %w", err)
	}

	return nil
}

// localName returns the HELO/EHLO name.
// RFC 5321 suggests the FQDN; the sender domain is a good proxy.
func (p *smtpProvider) localName() string {
	if i := strings.IndexByte(p.cfg.From, '@'); i >= 0 && i < len(p.cfg.From)-1 {
		return p.cfg.From[i+1:]
	}
	return "localhost"
}

// buildMessage constructs a minimal RFC 5322 HTML email.
//
// All WriteString calls use literals only; no string concatenation
// occurs at the call site, so each WriteString appends directly to
// the builder's internal buffer without an intermediate allocation.
func (p *smtpProvider) buildMessage(to, subject, htmlBody string) []byte {
	var b strings.Builder
	b.Grow(len(htmlBody) + 256)

	b.WriteString("From: ")
	if p.cfg.FromName != "" {
		b.WriteString(mime.QEncoding.Encode("utf-8", p.cfg.FromName))
		b.WriteString(" <")
		b.WriteString(p.cfg.From)
		b.WriteString(">")
	} else {
		b.WriteString(p.cfg.From)
	}
	b.WriteString("\r\n")

	b.WriteString("To: ")
	b.WriteString(to)
	b.WriteString("\r\n")

	b.WriteString("Subject: ")
	b.WriteString(mime.QEncoding.Encode("utf-8", subject))
	b.WriteString("\r\n")

	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("Date: ")
	b.WriteString(time.Now().UTC().Format(time.RFC1123Z))
	b.WriteString("\r\n")
	b.WriteString("\r\n")
	b.WriteString(htmlBody)

	return []byte(b.String())
}
