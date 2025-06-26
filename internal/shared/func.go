package shared

import (
	"context"
	"net/http"
	"time"
	"yourapp/internal/models"
)

func DeleteCookie(w http.ResponseWriter, r *http.Request, name string) {
	SetSecureCookie(w, r, name, "", -time.Second, true)
}

func IsSecureConnection(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func SetSecureCookie(w http.ResponseWriter, r *http.Request, name, value string, maxAge time.Duration, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: httpOnly,
		Secure:   IsSecureConnection(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(maxAge.Seconds()),
	})
}

func GetRequestID(ctx context.Context) string {
	// Example: retrieve from context if set by a middleware
	if id, ok := ctx.Value(RequestKEY).(RequestID); ok {
		return string(id)
	}
	return "unknown" // Or generate a new one if not found
}

func GetUser(ctx context.Context) *models.User {
	u, ok := ctx.Value(UserKey).(*models.User)
	if !ok {
		return nil
	}
	return u
}
