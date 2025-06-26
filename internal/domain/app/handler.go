package app

import (
	"log/slog"
	"net/http"
)

func Setup(s *service, l *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	return mux
}
