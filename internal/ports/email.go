package ports

// EmailSender is the contract for sending transactional emails.
type EmailSender interface {
	SendVerificationEmail(toEmail, token string) error
}
