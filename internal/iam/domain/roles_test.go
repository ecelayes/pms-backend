package domain

import (
	"testing"
)

func TestUserRole_NewUserRole(t *testing.T) {
	tests := []struct {
		input   string
		want    UserRole
		wantErr bool
	}{
		{string(RoleUser), RoleUser, false},
		{string(RoleSuperAdmin), RoleSuperAdmin, false},
		{"admin", "", true},
		{"owner", "", true},
		{"", "", true},
		{"USER", "", true}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := NewUserRole(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUserRole(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("NewUserRole(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestUserRole_String(t *testing.T) {
	if RoleUser.String() != "user" {
		t.Errorf("RoleUser.String() = %q, want user", RoleUser.String())
	}
	if RoleSuperAdmin.String() != "super_admin" {
		t.Errorf("RoleSuperAdmin.String() = %q, want super_admin", RoleSuperAdmin.String())
	}
}

func TestUserRole_IsSuperAdmin(t *testing.T) {
	if !RoleSuperAdmin.IsSuperAdmin() {
		t.Error("RoleSuperAdmin.IsSuperAdmin() = false, want true")
	}
	if RoleUser.IsSuperAdmin() {
		t.Error("RoleUser.IsSuperAdmin() = true, want false")
	}
}

func TestUserRole_IsUser(t *testing.T) {
	if !RoleUser.IsUser() {
		t.Error("RoleUser.IsUser() = false, want true")
	}
	if RoleSuperAdmin.IsUser() {
		t.Error("RoleSuperAdmin.IsUser() = true, want false")
	}
}
