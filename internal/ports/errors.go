package ports

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound                  = errors.New("not found")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrInvalidToken              = errors.New("invalid token")
	ErrTokenAlreadyUsed          = errors.New("token already used")
	ErrTokenExpired              = errors.New("token expired")
	ErrDuplicateEmail            = errors.New("email already registered")
	ErrEmailNotVerified          = errors.New("email not verified")
	ErrVerificationTokenInvalid  = errors.New("verification token is invalid or expired")
)

// AccountLockedError is returned when a login is rejected because the account
// has exceeded the failed-attempt limit.  RetryAfter is the remaining lockout
// duration, usable for the Retry-After HTTP header.
type AccountLockedError struct {
	RetryAfter time.Duration
}

func (e *AccountLockedError) Error() string {
	return fmt.Sprintf("account locked for another %s", e.RetryAfter.Round(time.Second))
}
