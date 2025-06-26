package auth

import (
	"log/slog"
	"net/http"
	"yourapp/internal/models"
	"yourapp/internal/shared"
	"yourapp/web"

	"golang.org/x/crypto/bcrypt"
)

// Setup initializes the authentication routes and handlers.
// It receives a service and a logger to propagate down to handlers.
func Setup(s *service, l *slog.Logger) http.Handler {
	oauth := s.oauth()
	mux := http.NewServeMux()

	// Handlers that don't need direct logging (rendering static pages)
	mux.HandleFunc("GET /register", getRegisterHandler)
	mux.HandleFunc("GET /login", getLoginHandler)

	// Handlers requiring logging context
	mux.HandleFunc("POST /register", postRegisterHandler(s, l))
	mux.HandleFunc("POST /login", postLoginHandler(s, l))
	mux.HandleFunc("GET /google/login", handleGoogleLogin(oauth))
	mux.HandleFunc("GET /google/callback", s.handleGoogleCallback(s, oauth))
	return mux
}

// getLoginHandler renders the login page.
// No specific logging needed here unless there's an error rendering.
func getLoginHandler(w http.ResponseWriter, r *http.Request) {
	web.RenderPage(w, r, "login.html", nil)
}

// getRegisterHandler renders the signup page.
// No specific logging needed here.
func getRegisterHandler(w http.ResponseWriter, r *http.Request) {
	web.RenderPage(w, r, "signup.html", nil)
}

// postLoginHandler handles user login requests.
func postLoginHandler(s *service, l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := l.With(
			slog.String("handler", "postLogin"),
			slog.String("request_id", shared.GetRequestID(r.Context())), // Placeholder for getting request ID
		)

		email := r.FormValue("email")
		user := models.User{
			Email:    email,
			Password: r.FormValue("password"),
		}

		log.Info("Attempting user login", slog.String("user_email", user.Email)) // Log attempt

		token, err := s.login(r.Context(), r, user)
		if err == nil {
			shared.SetSecureCookie(w, r, shared.SessionToken, token, shared.TwoDays, true)
			http.Redirect(w, r, "/app", http.StatusSeeOther)
			log.Info("User logged in successfully", slog.String("user_email", user.Email))
			return
		}

		switch err {
		case shared.ErrInvalidInputs:
			w.WriteHeader(http.StatusUnprocessableEntity)
			log.Warn("Login failed: Invalid inputs",
				slog.String("user_email", user.Email),
				slog.String("error_type", "ErrInvalidInputs"),
				slog.Any("error", err),
			)
		case bcrypt.ErrMismatchedHashAndPassword:
			w.WriteHeader(http.StatusUnprocessableEntity)
			log.Warn("Login failed: Invalid credentials (password mismatch)",
				slog.String("user_email", user.Email),
				slog.String("error_type", "ErrMismatchedHashAndPassword"),
				slog.Any("error", err), // Log the bcrypt error
			)
		case shared.ErrUserAlreadyExist: // This case typically doesn't apply to login, but kept for consistency
			w.WriteHeader(http.StatusConflict)
			log.Error("Login failed: User already exists (unexpected for login flow)",
				slog.String("user_email", user.Email),
				slog.String("error_type", "ErrUserAlreadyExist"),
				slog.Any("error", err),
			)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Login failed: Internal server error",
				slog.String("user_email", user.Email),
				slog.String("error_type", "InternalError"),
				slog.Any("error", err), // Log the actual error
			)
		}
	}
}

// postRegisterHandler handles new user registration requests.
func postRegisterHandler(ser *service, appLogger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := appLogger.With(
			slog.String("handler", "postRegister"),
			slog.String("request_id", shared.GetRequestID(r.Context())), // Placeholder
		)

		email := r.FormValue("email")
		user := models.NewUser{
			Email:    email,
			Password: r.FormValue("password"), // Password never logged
			Repeat:   r.FormValue("confirm"),
		}

		log.Info("Attempting user registration", slog.String("user_email", user.Email))

		token, err := ser.registerUser(r.Context(), r, user)
		if err == nil {
			shared.SetSecureCookie(w, r, shared.SessionToken, token, shared.TwoDays, true)
			http.Redirect(w, r, "/app", http.StatusSeeOther)
			log.Info("User registered successfully", slog.String("user_email", user.Email))
			return
		}

		switch err {
		case shared.ErrInvalidInputs:
			w.WriteHeader(http.StatusUnprocessableEntity)
			log.Warn("Registration failed: Invalid inputs",
				slog.String("user_email", user.Email),
				slog.String("error_type", "ErrInvalidInputs"),
				slog.Any("error", err),
			)
		case shared.ErrUserAlreadyExist:
			w.WriteHeader(http.StatusConflict)
			log.Warn("Registration failed: User already exists",
				slog.String("user_email", user.Email),
				slog.String("error_type", "ErrUserAlreadyExist"),
				slog.Any("error", err),
			)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Registration failed: Internal server error",
				slog.String("user_email", user.Email),
				slog.String("error_type", "InternalError"),
				slog.Any("error", err),
			)
		}
	}
}
