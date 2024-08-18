package notifier

import (
	"context"
	"fmt"

	emailprovider "github.com/ramisoul84/kfc-notification/internal/provider/email_provider"
	"github.com/ramisoul84/kfc-notification/pkg/logger"
)

// Notifier renders templates and dispatches notifications.
type Notifier struct {
	email    emailprovider.EmailProvider
	renderer *renderer
	log      *logger.Logger
}

// New creates a Notifier, parsing all embedded templates at startup.
func New(email emailprovider.EmailProvider, log *logger.Logger) (*Notifier, error) {
	r, err := newRenderer()
	if err != nil {
		return nil, err
	}
	return &Notifier{
		email:    email,
		renderer: r,
		log:      log,
	}, nil
}

// SendWelcomeEmail renders and sends the welcome email.
func (n *Notifier) SendWelcomeEmail(ctx context.Context, to, role, password string) error {
	if to == "" {
		return fmt.Errorf("welcome email: recipient is required")
	}

	body, err := n.renderer.render("welcome", map[string]string{
		"RoleName": role,
		"Password": password,
	})
	if err != nil {
		return err
	}

	if err := n.email.Send(ctx, to, "Welcome to KFC CRM", body); err != nil {
		n.log.Error("welcome email failed", "to", to, "error", err)
		return fmt.Errorf("welcome email: %w", err)
	}

	n.log.Info("welcome email sent", "to", to)
	return nil
}
