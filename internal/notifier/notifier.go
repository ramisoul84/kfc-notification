package notifier

import (
	"context"
	"fmt"
	"time"

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

type DevicePairingEmailInput struct {
	To             string
	RestaurantName string
	ExpiresAt      time.Time
	Devices        []DevicePairingItem
}

type DevicePairingItem struct {
	SerialNumber string
	DeviceType   string
	Code         string
}

func (n *Notifier) SendDevicePairingEmail(ctx context.Context, in *DevicePairingEmailInput) error {
	if in == nil || in.To == "" {
		return fmt.Errorf("device pairing email: recipient is required")
	}

	devices := make([]map[string]string, 0, len(in.Devices))
	for _, d := range in.Devices {
		devices = append(devices, map[string]string{
			"SerialNumber": d.SerialNumber,
			"DeviceType":   d.DeviceType,
			"Code":         d.Code,
		})
	}

	body, err := n.renderer.render("device_pairing", map[string]any{
		"RestaurantName": in.RestaurantName,
		"ExpiresAt":      in.ExpiresAt.Format("15:04 MST"),
		"Devices":        devices,
	})
	if err != nil {
		return err
	}

	subject := "Device pairing codes"
	if in.RestaurantName != "" {
		subject = "Device pairing codes — " + in.RestaurantName
	}

	if err := n.email.Send(ctx, in.To, subject, body); err != nil {
		n.log.Error("device pairing email failed", "to", in.To, "error", err)
		return fmt.Errorf("device pairing email: %w", err)
	}

	n.log.Info("device pairing email sent",
		"to", in.To,
		"device_count", len(in.Devices),
	)
	return nil
}
