package integration

import (
	"net/http"
	"testing"
)

// ── verify-email ──────────────────────────────────────────────────────────────

func TestVerifyEmail_Success(t *testing.T) {
	email := uniqueEmail()
	register(t, email, "Valid@1234")

	token, ok := testEmailSender.LastToken(email)
	if !ok {
		t.Fatal("expected a verification token to be captured")
	}

	resp := do(t, http.MethodPost, tenantPath("/auth/verify-email"), map[string]string{
		"email": email,
		"otp":   token,
	})
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status: got %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestVerifyEmail_InvalidOTP(t *testing.T) {
	resp := do(t, http.MethodPost, tenantPath("/auth/verify-email"), map[string]string{
		"email": "nobody@ci.test",
		"otp":   "000000",
	})
	assertError(t, resp, http.StatusUnprocessableEntity, "err_verification_token_invalid")
}

func TestVerifyEmail_MissingBody(t *testing.T) {
	resp := do(t, http.MethodPost, tenantPath("/auth/verify-email"), nil)
	assertError(t, resp, http.StatusBadRequest, "err_invalid_request")
}

// ── login blocked until verified ─────────────────────────────────────────────

func TestLogin_EmailNotVerified(t *testing.T) {
	email := uniqueEmail()
	register(t, email, "Valid@1234")

	// User is not in the database until OTP is verified, so login returns
	// invalid credentials rather than a "not verified" error.
	resp := do(t, http.MethodPost, tenantPath("/auth/login"), map[string]string{
		"email":    email,
		"password": "Valid@1234",
	})
	assertError(t, resp, http.StatusUnauthorized, "err_invalid_credentials")
}

func TestLogin_AllowedAfterVerification(t *testing.T) {
	email := uniqueEmail()
	register(t, email, "Valid@1234")
	verifyEmail(t, email)

	resp := do(t, http.MethodPost, tenantPath("/auth/login"), map[string]string{
		"email":    email,
		"password": "Valid@1234",
	})
	assertTokens(t, resp, http.StatusOK)
}

// ── resend verification ───────────────────────────────────────────────────────

func TestResendVerification_Success(t *testing.T) {
	email := uniqueEmail()
	register(t, email, "Valid@1234")

	// Resend should issue a new token.
	resp := do(t, http.MethodPost, tenantPath("/user/resend-verification"), map[string]string{
		"email": email,
	})
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status: got %d, want 202", resp.StatusCode)
	}
	resp.Body.Close()

	// The new token should still allow verification.
	verifyEmail(t, email)

	resp = do(t, http.MethodPost, tenantPath("/auth/login"), map[string]string{
		"email":    email,
		"password": "Valid@1234",
	})
	assertTokens(t, resp, http.StatusOK)
}

func TestResendVerification_UnknownEmail(t *testing.T) {
	// Must always return 202 — never reveal whether the email is registered.
	resp := do(t, http.MethodPost, tenantPath("/user/resend-verification"), map[string]string{
		"email": "nobody-" + uniqueEmail(),
	})
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status: got %d, want 202", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestResendVerification_AlreadyVerified(t *testing.T) {
	email := uniqueEmail()
	registerAndLogin(t, email, "Valid@1234") // registers + verifies

	// Resend on an already-verified account is a no-op but still 202.
	resp := do(t, http.MethodPost, tenantPath("/user/resend-verification"), map[string]string{
		"email": email,
	})
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status: got %d, want 202", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestResendVerification_MissingBody(t *testing.T) {
	resp := do(t, http.MethodPost, tenantPath("/user/resend-verification"), nil)
	assertError(t, resp, http.StatusBadRequest, "err_invalid_request")
}
