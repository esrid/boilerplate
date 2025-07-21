// Package auth provides authentication-related functionalities.
package auth

import (
	"boilerplate/internal/models"
	"boilerplate/pkg/utils"
	"boilerplate/views/components"
	"boilerplate/views/pages"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewHandler(s Service) *chi.Mux {
	r := chi.NewRouter()
	h := &handler{service: s}
	r.Get("/login", h.loginPage)
	r.Post("/login", h.login)
	r.Get("/register", h.registerPage)
	r.Post("/register", h.register)
	return r
}

type handler struct {
	service Service
}

func (h *handler) loginPage(w http.ResponseWriter, r *http.Request) {
	components.Layout(components.PageHead{Title: "Login"}, pages.LoginPage("")).Render(r.Context(), w)
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	payload := new(models.LoginPayload)
	if err := r.ParseForm(); err != nil {
		pages.LoginPage("invalid request").Render(r.Context(), w)
		return
	}
	payload.Email = r.FormValue("email")
	payload.Password = r.FormValue("password")

	if err := utils.Validate(payload); err != nil {
		pages.LoginPage(err.Error()).Render(r.Context(), w)
		return
	}

	session, err := h.service.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		pages.LoginPage(err.Error()).Render(r.Context(), w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
	})
	fmt.Println("cooool")

	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func (h *handler) registerPage(w http.ResponseWriter, r *http.Request) {
	components.Layout(components.PageHead{Title: "Register"}, pages.RegisterPage("")).Render(r.Context(), w)
}

func (h *handler) register(w http.ResponseWriter, r *http.Request) {
	payload := new(models.NewUser)
	if err := r.ParseForm(); err != nil {
		pages.RegisterPage("invalid request").Render(r.Context(), w)
		return
	}

	payload.Email = r.FormValue("email")
	payload.Password = r.FormValue("password")
	payload.Repeat = r.FormValue("repeat")

	if err := utils.Validate(payload); err != nil {
		slog.Error("unable to regiser user", slog.String("error", err.Error()))
		pages.RegisterPage(err.Error()).Render(r.Context(), w)
		return
	}

	session, err := h.service.Register(r.Context(), payload.Email, payload.Password)
	if err != nil {
		slog.Error("unable to regiser user", slog.String("error", err.Error()))
		pages.RegisterPage(err.Error()).Render(r.Context(), w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}
