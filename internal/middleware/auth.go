package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"yourapp/internal/shared"
)

func AuthRequired(auth shared.Authstore, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ck, err := r.Cookie(shared.SessionToken)
			if err != nil || ck == nil || ck.Value == "" {
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			ss, err := auth.GetBySessionToken(r.Context(), ck.Value)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
					return
				}
				log.Error("failed to get session by token", slog.String("error", err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if time.Now().After(ss.ExpiresAt) {
				shared.DeleteCookie(w, r, ck.Name)
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			newExpiry := ss.ExpiresAt.Add(time.Hour)
			if err := auth.UpdateExpiry(r.Context(), ss.Token, newExpiry); err != nil {
				log.Error("failed to update session expiry", slog.String("error", err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			user, err := auth.GetUserBySessionID(r.Context(), ss.Token)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
					return
				}
				log.Error("failed to get user by session ID", slog.String("error", err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			// Add user and session to context
			ctx := context.WithValue(r.Context(), shared.UserKey, user)
			ctx = context.WithValue(ctx, shared.SessionKey, ss)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
