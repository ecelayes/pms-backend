package domain

import (
	"errors"
	"github.com/google/uuid"
	"regexp"
	"time"
)

var (
	ErrInvalidEmail = errors.New("invalid email")
	emailRegex      = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
)

type User struct {
	id        string
	email     string
	password  string
	salt      string
	role      UserRole
	firstName string
	lastName  string
	phone     string
	createdAt time.Time
}

func NewUser(email, hashedPassword, salt string, role UserRole, firstName, lastName, phone string) (*User, error) {
	if email == "" || !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	if firstName == "" || lastName == "" {
		return nil, errors.New("invalid name: first and last name are required")
	}
	return &User{
		id:        uuid.New().String(),
		email:     email,
		password:  hashedPassword,
		salt:      salt,
		role:      role,
		firstName: firstName,
		lastName:  lastName,
		phone:     phone,
		createdAt: time.Now(),
	}, nil
}
func ReconstituteUser(
	id, email, password, salt, role, firstName, lastName, phone string,
	createdAt time.Time,
) *User {
	return &User{
		id:        id,
		email:     email,
		password:  password,
		salt:      salt,
		role:      UserRole(role),
		firstName: firstName,
		lastName:  lastName,
		phone:     phone,
		createdAt: createdAt,
	}
}
func (u *User) ID() string        { return u.id }
func (u *User) Email() string     { return u.email }
func (u *User) Password() string  { return u.password }
func (u *User) Salt() string      { return u.salt }
func (u *User) Role() UserRole    { return u.role }
func (u *User) FirstName() string { return u.firstName }
func (u *User) LastName() string  { return u.lastName }
func (u *User) Phone() string     { return u.phone }
func (u *User) Update(role UserRole, firstName, lastName, phone string) {
	if role != "" {
		u.role = role
	}
	if firstName != "" {
		u.firstName = firstName
	}
	if lastName != "" {
		u.lastName = lastName
	}
	if phone != "" {
		u.phone = phone
	}
}
func (u *User) ChangePassword(hashedPassword, salt string) {
	u.password = hashedPassword
	u.salt = salt
}
