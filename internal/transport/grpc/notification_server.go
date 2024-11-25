package grpc

import (
	"context"

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
