package http

import "testing"

func TestValidateEmail(t *testing.T) {
	valid := []string{
		"user@example.com",
		"first.last@example.co.uk",
		"a@b.io",
		"user+tag@domain.com",
	}
	for _, e := range valid {
		if err := ValidateEmail(e); err != nil {
			t.Errorf("expected valid %q, got error: %v", e, err)
		}
	}

	invalid := []string{
		"",
		"no-at-sign",
		"@no-user.com",
		"no-domain@",
		"spaces in@email.com",
		"two@@at.com",
	}
	for _, e := range invalid {
		if err := ValidateEmail(e); err == nil {
			t.Errorf("expected invalid %q, got nil", e)
		}
	}
}

func TestValidateRequired(t *testing.T) {
	if err := ValidateRequired("name", "John"); err != nil {
		t.Errorf("non-empty should be valid: %v", err)
	}
	if err := ValidateRequired("name", ""); err == nil {
		t.Error("empty should be invalid")
	}
	if err := ValidateRequired("name", "   "); err == nil {
		t.Error("whitespace-only should be invalid")
	}
}

func TestValidateUUID(t *testing.T) {
	if err := ValidateUUID("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Errorf("valid UUID rejected: %v", err)
	}
	if err := ValidateUUID("not-a-uuid"); err == nil {
		t.Error("invalid UUID accepted")
	}
	if err := ValidateUUID(""); err == nil {
		t.Error("empty accepted")
	}
}

func TestValidateEmail_RejectsDisplayName(t *testing.T) {
	if err := ValidateEmail("John Doe <john@example.com>"); err == nil {
		t.Error("expected display name to be rejected")
	}
}
