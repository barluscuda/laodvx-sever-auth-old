package email

import "sync"

// CaptureSender is an in-memory EmailSender for use in tests.
// It records every (email, token) pair that would have been sent so that
// tests can retrieve the verification token without a real SMTP server.
type CaptureSender struct {
	mu     sync.Mutex
	tokens map[string]string // email → last token
}

func NewCaptureSender() *CaptureSender {
	return &CaptureSender{tokens: make(map[string]string)}
}

func (c *CaptureSender) SendVerificationEmail(toEmail, token string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[toEmail] = token
	return nil
}

// LastToken returns the most-recently captured token for the given email.
// Returns ("", false) if no token has been captured for that address.
func (c *CaptureSender) LastToken(email string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	t, ok := c.tokens[email]
	return t, ok
}
