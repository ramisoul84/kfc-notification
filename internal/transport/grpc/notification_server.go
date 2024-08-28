package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	notificationv1 "github.com/ramisoul84/kfc-notification/gen/notification/v1"
	"github.com/ramisoul84/kfc-notification/internal/notifier"
	"github.com/ramisoul84/kfc-notification/pkg/logger"
)

// NotificationServer implements notificationv1.NotificationServiceServer.
type NotificationServer struct {
	notificationv1.UnimplementedNotificationServiceServer

	notifier *notifier.Notifier
	logger   *logger.Logger
}

func NewNotificationServer(n *notifier.Notifier, log *logger.Logger) *NotificationServer {
	return &NotificationServer{notifier: n, logger: log}
}

func (s *NotificationServer) SendWelcomeEmail(
	ctx context.Context,
	req *notificationv1.SendWelcomeEmailRequest,
) (*notificationv1.SendWelcomeEmailResponse, error) {
	to := firstNonEmpty(req.GetTo(), req.GetEmail())
	if to == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient email required")
	}

	if err := s.notifier.SendWelcomeEmail(
		ctx,
		to,
		req.GetRoleName(),
		req.GetInitialPassword(),
	); err != nil {
		s.logger.Error("welcome email failed", "to", to, "error", err)
		return nil, status.Error(codes.Internal, "failed to send welcome email")
	}

	return &notificationv1.SendWelcomeEmailResponse{
		Success:        true,
		Message:        "welcome email sent",
		NotificationId: uuid.NewString(),
		SentAt:         timestamppb.Now(),
	}, nil
}

func (s *NotificationServer) SendDevicePairingEmail(
	ctx context.Context,
	req *notificationv1.SendDevicePairingEmailRequest,
) (*notificationv1.SendDevicePairingEmailResponse, error) {
	if req.GetTo() == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient email required")
	}
	if len(req.GetDevices()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one device is required")
	}

	if err := s.notifier.SendDevicePairingEmail(ctx, &notifier.DevicePairingEmailInput{
		To:             req.GetTo(),
		RestaurantName: req.GetRestaurantName(),
		ExpiresAt:      time.Unix(req.GetExpiresAt(), 0).UTC(),
		Devices:        toNotifierDevices(req.GetDevices()),
	}); err != nil {
		s.logger.Error("device pairing email failed",
			"to", req.GetTo(),
			"error", err,
		)
		return nil, status.Error(codes.Internal, "failed to send device pairing email")
	}

	return &notificationv1.SendDevicePairingEmailResponse{
		Success: true,
		Message: "device pairing email sent",
		SentAt:  timestamppb.Now(),
	}, nil
}

func toNotifierDevices(in []*notificationv1.PairingItem) []notifier.DevicePairingItem {
	out := make([]notifier.DevicePairingItem, 0, len(in))
	for _, d := range in {
		out = append(out, notifier.DevicePairingItem{
			SerialNumber: d.GetSerialNumber(),
			DeviceType:   d.GetDeviceType(),
			Code:         d.GetCode(),
		})
	}
	return out
}

// firstNonEmpty returns the first non-empty string.
// Kept because the proto carries both "to" and "email" for compatibility.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
