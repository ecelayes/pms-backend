package domain

import (
	"strings"
	"testing"
)

func TestNewPassword_RejectsTooShort(t *testing.T) {
	for _, p := range []string{"", "a", "abcdefg"} {
		if _, err := NewPassword(p); err == nil {
			t.Errorf("expected error for %q, got nil", p)
		}
	}
}

func TestNewPassword_RejectsCommonPasswords(t *testing.T) {
	common := []string{
		"password123",
		"Password123",
		"123456789",
		"qwerty123",
		"Admin12345",
		"Welcome1!",
		"Passw0rd",
	}
	for _, p := range common {
		_, err := NewPassword(p)
		if err == nil {
			t.Errorf("expected common-password rejection for %q", p)
			continue
		}
		if !strings.Contains(err.Error(), "common") {
			t.Errorf("error for %q should mention commonness: %v", p, err)
		}
	}
}

func TestNewPassword_RequiresVariety(t *testing.T) {
	// All lowercase - too simple
	if _, err := NewPassword("alllowercasepassword"); err == nil {
		t.Error("all-lowercase long password should be rejected (no variety)")
	}
	// All digits
	if _, err := NewPassword("1234567890"); err == nil {
		t.Error("all-digits password should be rejected")
	}
}

func TestNewPassword_AcceptsStrong(t *testing.T) {
	strong := []string{
		"MyC0mpl3x.Pass",
		"T0pSecret!Hunter2",
		"Xk$3p4mq9!vR",
		"abc123XYz!",
	}
	for _, p := range strong {
		if pw, err := NewPassword(p); err != nil {
			t.Errorf("strong password %q rejected: %v", p, err)
		} else if pw.String() != p {
			t.Errorf("expected round-trip %q, got %q", p, pw.String())
		}
	}
}

func TestNewPassword_StripsLeadingTrailingSpaces(t *testing.T) {
	pw, err := NewPassword("  GoodP4ss.Word  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pw.String() != "GoodP4ss.Word" {
		t.Errorf("expected trimmed value, got %q", pw.String())
	}
}

func TestPassword_String(t *testing.T) {
	pw, err := NewPassword("Good.Pass1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pw.String() != "Good.Pass1" {
		t.Errorf("expected Good.Pass1, got %q", pw.String())
	}
}

func TestPassword_Equals(t *testing.T) {
	a, _ := NewPassword("Same.Pass1")
	b, _ := NewPassword("Same.Pass1")
	c, _ := NewPassword("Other.Pass1")
	if !a.Equals(b) {
		t.Error("equal passwords should be equal")
	}
	if a.Equals(c) {
		t.Error("different passwords should not be equal")
	}
}

func TestNewPasswordFromHash(t *testing.T) {
	hashed := "$2a$10$abcdef1234567890"
	pw := NewPasswordFromHash(hashed)
	if pw.String() != hashed {
		t.Errorf("expected hash round-trip, got %q", pw.String())
	}
}

func TestPassword_IsEmpty(t *testing.T) {
	if !NewPasswordFromHash("").IsEmpty() {
		t.Error("empty hash should be empty")
	}
	pw, _ := NewPassword("GoodP4ss!")
	if pw.IsEmpty() {
		t.Error("real password should not be empty")
	}
}

func TestNewPassword_RejectsTooLong(t *testing.T) {
	long := strings.Repeat("Aa1", 50)
	if _, err := NewPassword(long); err == nil {
		t.Error("expected length error for >128 char password")
	}
}

func TestNewPassword_RejectsLowerDigitOnly(t *testing.T) {
	// Lowercase + digit (no uppercase) should be rejected - needs upper for variety
	if _, err := NewPassword("longpassword9"); err == nil {
		t.Error("lowercase+digits should be rejected (no uppercase)")
	}
}

func TestNewPassword_AcceptsUpperDigitOnly(t *testing.T) {
	// Uppercase + digit (no lowercase) is accepted - has variety
	if _, err := NewPassword("ALLCAPS9"); err != nil {
		t.Errorf("uppercase+digits should be valid variety, got: %v", err)
	}
}

func TestNewPassword_AcceptsUpperLowerOnly(t *testing.T) {
	if _, err := NewPassword("MyGoodPassword"); err != nil {
		t.Errorf("upper+lower should be valid variety, got: %v", err)
	}
}

func TestNewPassword_RejectsUpperOnlyNoLowerNoDigit(t *testing.T) {
	// 8 chars, all uppercase - has upper but no lower or digit
	if _, err := NewPassword("ABCDEFGH"); err == nil {
		t.Error("all-uppercase with no lower/digit should be rejected (no variety)")
	}
}
