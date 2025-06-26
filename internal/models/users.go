package models

import "time"

type User struct {
	ID        string
	Email     string `validate:"required,email,max=256"`
	Password  string `validate:"required"`
	GoogleID  string
	Oauth     bool
	Verify    bool
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NewUser struct {
	Email    string `validate:"required,email,max=256"`
	Password string `validate:"required"`
	Repeat   string
}

type GoogleUser struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}
