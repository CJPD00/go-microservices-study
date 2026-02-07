package domain

import (
	"regexp"
	"time"
)

// User represents the user domain entity
type User struct {
	ID        uint
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// EmailRegex is the pattern for validating emails
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validate validates the user entity
func (u *User) Validate() error {
	if u.Name == "" {
		return ErrNameRequired
	}
	if len(u.Name) < 2 || len(u.Name) > 100 {
		return ErrNameLength
	}
	if u.Email == "" {
		return ErrEmailRequired
	}
	if !EmailRegex.MatchString(u.Email) {
		return ErrEmailInvalid
	}
	return nil
}

// NewUser creates a new user with validation
func NewUser(name, email string) (*User, error) {
	user := &User{
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}
