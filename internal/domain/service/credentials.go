package service

import (
	"errors"

	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

// dummyHash equalizes bcrypt timing for missing credentials, preventing
// email/username enumeration via response-time analysis.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-timing-equalization"), 12)

// credentialFinder looks up the stored password hash for a caller.
// It must return ports.ErrNotFound when no such credential exists.
type credentialFinder func() (passwordHash string, err error)

// verifyCredentials implements the lockout + constant-time credential check.
// On success it returns nil. Callers must consume ErrInvalidCredentials and
// AccountLockedError appropriately.
func verifyCredentials(
	attempts ports.LoginAttemptRepository,
	key, password string,
	find credentialFinder,
) error {
	if locked, retryAfter, err := attempts.IsLocked(key); err != nil {
		return err
	} else if locked {
		return &ports.AccountLockedError{RetryAfter: retryAfter}
	}

	hash, err := find()
	if errors.Is(err, ports.ErrNotFound) {
		// Always run bcrypt to equalize timing and avoid enumeration.
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		if recErr := attempts.Record(key); recErr != nil {
			return recErr
		}
		return ports.ErrInvalidCredentials
	}
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if recErr := attempts.Record(key); recErr != nil {
			return recErr
		}
		return ports.ErrInvalidCredentials
	}

	return attempts.Reset(key)
}
