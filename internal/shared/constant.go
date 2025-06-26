package shared

import (
	"context"
	"time"
	"yourapp/internal/models"
)

var TwoDays = 2 * 24 * time.Hour

var (
	SessionToken string      = "session_token"
	UserKey      UserType    = "user_key"
	SessionKey   SessionType = "session_key"
	RequestKEY   RequestID   = "RequestID"

	SessionAttributes = "user_id, token, csrf_code, created_at, expires_at, ip_address, user_agent"
	UserAttribute     = "id, email, password_hash, google_id, oauth, verify, role, created_at, updated_at"
)

type (
	UserType    string
	SessionType string
	RequestID   string
)

type Authstore interface {
	GetUserBySessionID(ctx context.Context, sid string) (*models.User, error)
	GetBySessionToken(ctx context.Context, cookieHash string) (*models.Session, error)
	UpdateExpiry(ctx context.Context, token string, expiresAt time.Time) error
}
