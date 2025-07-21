// Package server provides the server for the application.
package server

import (
	"boilerplate/internal/domain/auth"
	"boilerplate/views"
	"boilerplate/views/components"
	"boilerplate/views/pages"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/jackc/pgx/v5"
)

func NewRouter(conn *pgx.Conn) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Handle("/static/*", http.FileServer(http.FS(views.StaticFiles)))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		components.Layout(components.PageHead{Title: "About"}, pages.About()).Render(r.Context(), w)
	})

	authHandler := auth.NewHandler(auth.NewService(auth.NewRepository(conn)))

	router.Mount("/auth", authHandler)

	return router
}
