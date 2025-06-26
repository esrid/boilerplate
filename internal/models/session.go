package models

import (
	"net"
	"time"
)

type Session struct {
	UserID    string
	Token     string
	CsrfCode  string
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress net.IP
	UserAgent string
}

type Otp struct {
	Code      int
	UserId    string
	CreatedAt time.Time
	Used      bool
}
