package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_NewUser(t *testing.T) {
	t.Run("valid user creation", func(t *testing.T) {
		user, err := NewUser(
			"test@example.com",
			"hashedpassword",
			"salt123",
			RoleUser,
			"John",
			"Doe",
			"+1234567890",
		)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email())
		assert.Equal(t, "John", user.FirstName())
		assert.Equal(t, "Doe", user.LastName())
		assert.Equal(t, RoleUser, user.Role())
		assert.NotEmpty(t, user.ID())
	})

	t.Run("rejects invalid email", func(t *testing.T) {
		user, err := NewUser(
			"invalid-email",
			"hashedpassword",
			"salt123",
			RoleUser,
			"John",
			"Doe",
			"+1234567890",
		)

		assert.ErrorIs(t, err, ErrInvalidEmail)
		assert.Nil(t, user)
	})

	t.Run("rejects empty first name", func(t *testing.T) {
		user, err := NewUser(
			"test@example.com",
			"hashedpassword",
			"salt123",
			RoleUser,
			"",
			"Doe",
			"+1234567890",
		)

		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("rejects empty last name", func(t *testing.T) {
		user, err := NewUser(
			"test@example.com",
			"hashedpassword",
			"salt123",
			RoleUser,
			"John",
			"",
			"+1234567890",
		)

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUser_ReconstituteUser(t *testing.T) {
	user := ReconstituteUser(
		"user-123",
		"test@example.com",
		"hashedpassword",
		"salt123",
		"super_admin",
		"Jane",
		"Smith",
		"+0987654321",
		time.Now(),
	)

	assert.Equal(t, "user-123", user.ID())
	assert.Equal(t, "test@example.com", user.Email())
	assert.Equal(t, RoleSuperAdmin, user.Role())
	assert.Equal(t, "Jane", user.FirstName())
	assert.Equal(t, "Smith", user.LastName())
}

func TestUser_Update(t *testing.T) {
	user, _ := NewUser(
		"test@example.com",
		"hashedpassword",
		"salt123",
		RoleUser,
		"John",
		"Doe",
		"+1234567890",
	)

	t.Run("update role only", func(t *testing.T) {
		user.Update(RoleSuperAdmin, "", "", "")
		assert.Equal(t, RoleSuperAdmin, user.Role())
		assert.Equal(t, "John", user.FirstName()) // unchanged
	})

	t.Run("update name only", func(t *testing.T) {
		user.Update("", "Jane", "", "")
		assert.Equal(t, "Jane", user.FirstName())
		assert.Equal(t, "Doe", user.LastName())
	})

	t.Run("update phone only", func(t *testing.T) {
		user.Update("", "", "", "+9999999999")
		assert.Equal(t, "+9999999999", user.Phone())
	})
}

func TestUser_ChangePassword(t *testing.T) {
	user, _ := NewUser(
		"test@example.com",
		"oldpassword",
		"oldsalt",
		RoleUser,
		"John",
		"Doe",
		"+1234567890",
	)

	user.ChangePassword("newpassword", "newsalt")

	assert.Equal(t, "newpassword", user.Password())
	assert.Equal(t, "newsalt", user.Salt())
}
