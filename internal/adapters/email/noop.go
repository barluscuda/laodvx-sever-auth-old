package email

// NoopSender discards all emails. Used when SMTP is not configured.
type NoopSender struct{}

func (NoopSender) SendVerificationEmail(_, _ string) error { return nil }
