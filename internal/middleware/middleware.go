package middleware

import (
	"log/slog"
	"yourapp/internal/shared"
)

type M struct {
	Log  *slog.Logger
	Auth shared.Authstore
}
