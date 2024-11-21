package emailprovider

import (
	"context"

	"github.com/ramisoul84/kfc-notification/pkg/logger"
)

type logProvider struct {
	log *logger.Logger
}

func newLog(log *logger.Logger) *logProvider {
	return &logProvider{log: log}
}

func (p *logProvider) Send(ctx context.Context, to, subject, htmlBody string) error {
	p.log.Info("email (log provider)",
		"to", to,
		"subject", subject,
		"body_length", len(htmlBody),
	)
	return nil
}
