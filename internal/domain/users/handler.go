package users

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"yourapp/internal/models"
)

func Setup(s *service, l *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /me", deleteHandler(s))
	mux.HandleFunc("PUT /me", updateHandler(s))
	return mux
}

func deleteHandler(s *service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err := s.r.getUserByID(r.Context(), id)
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err := s.r.delete(r.Context(), id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func updateHandler(s *service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := models.User{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		_, err := s.update(r.Context(), user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
