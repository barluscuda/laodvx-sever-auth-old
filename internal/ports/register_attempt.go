package ports

import "time"

// RegistrationAttemptRepository tracks failed OTP-verification attempts per
// email to support a registration-verify lockout.
type RegistrationAttemptRepository interface {
	// IsLocked returns (true, remainingTTL, nil) when the key is locked.
	// Returns (false, 0, nil) when the key is not locked or does not exist.
	IsLocked(key string) (locked bool, retryAfter time.Duration, err error)
	// Record increments the failed attempt counter for key.
	Record(key string) error
	// Reset clears the failed attempt counter after a successful verification.
	Reset(key string) error
}
