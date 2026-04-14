package ports

import "time"

// LoginAttemptRepository tracks failed login attempts per account to support lockout.
type LoginAttemptRepository interface {
	// IsLocked returns (true, remainingTTL, nil) when the key is locked.
	// Returns (false, 0, nil) when the key is not locked or does not exist.
	IsLocked(key string) (locked bool, retryAfter time.Duration, err error)
	// Record increments the failed attempt counter for key.
	Record(key string) error
	// Reset clears the failed attempt counter after a successful login.
	Reset(key string) error
}
