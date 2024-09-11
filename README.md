# kfc-notification

Outbound notification service for the KFC Russia platform. Sends
transactional emails triggered by CRM events.

## What it does

- Sends templated HTML emails via a pluggable provider (`smtp`, `resend`,
  or `log` for local development).
- Exposes a small gRPC API so other services can request notifications
  without embedding email logic.
- Currently handles two events:
  - **Welcome email** — sent when a CRM user is created, contains the
    initial password.
  - **Device pairing email** — sent when a restaurant manager generates
    pairing codes for a batch of devices, contains the code table.

## Architecture
CRM ──gRPC──▶┌─────────────────────────┐
│ │
│ kfc-notification │
│ │
└──────┬──────────────────┘
│
┌──────────┴───────────┐
│ │
┌─────▼─────┐ ┌───────▼───────┐
│ Notifier │ │ Provider │
│ (templates│ │ (smtp/resend/ │
│ + logic) │ │ log) │
└───────────┘ └───────────────┘


Two layers:

- **Notifier** — owns the business logic. Renders templates, picks the
  subject, logs the outcome. Knows nothing about SMTP or Resend.
- **Provider** — owns delivery. One interface (`EmailProvider.Send`), three
  implementations. Swap by config; callers never change.

Templates are embedded via `//go:embed` — no runtime file paths, no
Dockerfile `COPY`.

## Key flows

### Welcome email

1. CRM creates a user, generates an initial password.
2. CRM calls `SendWelcomeEmail` over gRPC with the recipient, role, and
   password.
3. `NotificationServer.SendWelcomeEmail` validates the input and calls
   `Notifier.SendWelcomeEmail`.
4. Notifier renders `welcome.html` with `{RoleName, Password}` and calls
   the configured provider.
5. Provider sends via SMTP or Resend (or logs it in dev).

### Device pairing email

1. CRM generates pairing codes for every device in a restaurant.
2. CRM calls `SendDevicePairingEmail` over gRPC with the recipient,
   restaurant name, expiry, and a list of `{serial, type, code}`.
3. Notifier renders `device_pairing.html` — a table of codes.
4. Provider delivers it.

## gRPC API

```proto
service NotificationService {
  rpc SendWelcomeEmail(SendWelcomeEmailRequest) returns (SendWelcomeEmailResponse);
  rpc SendDevicePairingEmail(SendDevicePairingEmailRequest) returns (SendDevicePairingEmailResponse);
}