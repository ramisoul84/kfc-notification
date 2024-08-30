package smsprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/ramisoul84/kfc-notification/internal/config"
	"github.com/ramisoul84/kfc-notification/pkg/logger"
)

// SmsProvider delivers a plain-text SMS.
//
// Only the log provider ships today. A real gateway (SMS.ru, SMSC,
// Twilio, …) is one more implementation of this interface plus a case in
// NewSmsProvider; callers never change.
type SmsProvider interface {
	Send(ctx context.Context, to, text string) error
}

func NewSmsProvider(cfg config.SMSConfig, log *logger.Logger) (SmsProvider, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "log", "":
		return &logProvider{log: log}, nil
	default:
		return nil, fmt.Errorf("unknown sms provider %q", cfg.Provider)
	}
}

// logProvider writes the message to the service log instead of sending it.
// Development only: the message contains the login code.
type logProvider struct {
	log *logger.Logger
}

func (p *logProvider) Send(_ context.Context, to, text string) error {
	p.log.Info("📱 sms (log provider, not sent)", "to", to, "text", text)
	return nil
}
