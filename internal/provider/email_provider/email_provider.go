package emailprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/ramisoul84/kfc-notification/internal/config"
	"github.com/ramisoul84/kfc-notification/pkg/logger"
)

// EmailProvider sends an already-rendered email.
type EmailProvider interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}

func NewEmailProvider(cfg config.EmailConfig, log *logger.Logger) (EmailProvider, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "smtp":
		if cfg.SMTPHost == "" {
			return nil, fmt.Errorf("smtp: host is required")
		}
		if cfg.SMTPPort == 0 {
			return nil, fmt.Errorf("smtp: port is required")
		}
		if cfg.From == "" {
			return nil, fmt.Errorf("smtp: from address is required")
		}
		return newSMTP(cfg, log), nil

	case "resend":
		if cfg.ResendAPIKey == "" {
			return nil, fmt.Errorf("resend: api key is required")
		}
		if cfg.From == "" {
			return nil, fmt.Errorf("resend: from address is required")
		}
		return newResend(cfg, log), nil

	case "log", "":
		return newLog(log), nil

	default:
		return nil, fmt.Errorf("unknown email provider %q", cfg.Provider)
	}
}
