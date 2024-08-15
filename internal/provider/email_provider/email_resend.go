package emailprovider

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v3"

	"github.com/ramisoul84/kfc-notification/internal/config"
	"github.com/ramisoul84/kfc-notification/pkg/logger"
)

// resendProvider sends email via the Resend HTTP API using the official SDK.
type resendProvider struct {
	client *resend.Client
	cfg    config.EmailConfig
	log    *logger.Logger
}

func newResend(cfg config.EmailConfig, log *logger.Logger) *resendProvider {
	return &resendProvider{
		client: resend.NewClient(cfg.ResendAPIKey),
		cfg:    cfg,
		log:    log,
	}
}

func (p *resendProvider) Send(ctx context.Context, to, subject, htmlBody string) error {
	from := p.cfg.From
	if p.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", p.cfg.FromName, p.cfg.From)
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := p.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend: send: %w", err)
	}

	p.log.Debug("resend email sent",
		"id", sent.Id,
		"to", to,
	)

	return nil
}
