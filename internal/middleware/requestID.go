package middleware

import (
	"context"
	"net/http"
	"time"
	"yourapp/internal/shared"
)

func RequesID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := "req-" + time.Now().Format("20060102150405.000000") // Simple placeholder
		ctx := context.WithValue(r.Context(), shared.RequestKEY, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
