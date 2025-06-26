package server

import (
	"log/slog"
	"net/http"
	"yourapp/config"
	"yourapp/internal/domain/app"
	"yourapp/internal/domain/auth"
	"yourapp/internal/domain/users"
	"yourapp/internal/middleware"
	"yourapp/internal/shared"
	"yourapp/web"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

// AppRouter handles HTTP routing
type AppRouter struct {
	db  *pgxpool.Pool
	log *slog.Logger
	cfg *config.AppConfig
}

// NewRouter creates a new router instance
func NewRouter(db *pgxpool.Pool, config *config.AppConfig, log *slog.Logger) *AppRouter {
	return &AppRouter{
		db:  db,
		cfg: config,
		log: log,
	}
}

// Handler returns the configured HTTP handler
func (r *AppRouter) Handler() http.Handler {
	return r.applyGlobalMiddleware(r.setupRoutes())
}

func (r *AppRouter) setupRoutes() *http.ServeMux {
	// AUTH MODULE
	authR := auth.NewRepository(r.db)
	authS := auth.NewService(authR, r.cfg.APIKEY)

	// USER MODULE
	userR := users.NewRepository(r.db)
	userS := users.NewService(userR)

	// // APP MODULE
	appR := app.NewRepository(r.db)
	appS := app.NewService(appR)

	mux := http.NewServeMux()
	web.Static(mux)

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, req *http.Request) {
		web.RenderPage(w, req, "home.html", nil)
	})

	appHandler := applyMiddlewares(app.Setup(appS, r.log), r.secureRoute(authR)...)
	userHandler := applyMiddlewares(users.Setup(userS, r.log), r.secureRoute(authR)...)

	mux.Handle("/app/", http.StripPrefix("/app", appHandler))
	mux.Handle("/me/", http.StripPrefix("/me", userHandler))

	mux.Handle("/auth/", http.StripPrefix("/auth", auth.Setup(authS, r.log)))

	return mux
}

func (r *AppRouter) applyGlobalMiddleware(handler http.Handler) http.Handler {
	middlewares := r.getGlobalMiddlewares()
	return applyMiddlewares(handler, middlewares...)
}

func (r *AppRouter) getGlobalMiddlewares() []Middleware {
	rateLimiter := middleware.NewRateLimiter(
		rate.Limit(r.cfg.HTTP.RateLimit),
		r.cfg.HTTP.RateBurst,
		r.cfg.HTTP.RateCleanupInt,
	)

	return []Middleware{
		middleware.RequesID,
		// middleware.Logging(r.log),
		rateLimiter.Middleware,
		middleware.Recovery(r.log), // Pass nil for logger as it's removed
	}
}

func applyMiddlewares(final http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final)
	}
	return final
}

func (r *AppRouter) secureRoute(auth shared.Authstore) []Middleware {
	return []Middleware{
		middleware.AuthRequired(auth, r.log),
		middleware.CSRF,
	}
}
