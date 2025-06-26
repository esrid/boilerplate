package middleware

import (
	"log"
	"net/http"
	"yourapp/internal/models"
	"yourapp/internal/shared"
)

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost ||
			r.Method == http.MethodPut ||
			r.Method == http.MethodDelete ||
			r.Method == http.MethodPatch { // PATCH is also often considered unsafe

			ses, found := r.Context().Value(shared.SessionKey).(*models.Session)
			if !found {
				log.Println("unable to find user session")
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}

			// Get the CSRF token from the request header
			headerCsrfToken := r.Header.Get("X-CSRF-TOKEN")

			// Compare the header CSRF token with the session's CsrfCode
			if headerCsrfToken == "" {
				log.Println("unable to find headerCsrfToken")
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}
			if headerCsrfToken != ses.CsrfCode {
				log.Println("header csrf token and session csrf token do not math")
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
