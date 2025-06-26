package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"yourapp/internal/models"
	"yourapp/internal/shared"
	"yourapp/pkg"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (s *service) oauth() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.k.GoogleSecret,
		ClientSecret: s.k.GoogleSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  s.k.GoogleCallBack,
		Scopes:       []string{"email", "profile"},
	}
}

const (
	oauthStateCookie         = "oauth_state"
	oauthStateMaxAge         = 10 * time.Minute
	sessionMaxAge            = 24 * time.Hour
	defaultCallback          = "http://localhost:80/auth/google/callback"
	googleUserInfoURL string = "https://www.googleapis.com/oauth2/v2/userinfo"
	redirectPath      string = "/app"
)

func handleGoogleLogin(oauth *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := pkg.GenerateToken()

		http.SetCookie(w, &http.Cookie{
			Name:     oauthStateCookie,
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Secure:   shared.IsSecureConnection(r),
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(oauthStateMaxAge.Seconds()),
		})

		authURL := oauth.AuthCodeURL(state)
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}

func (s *service) handleGoogleCallback(service *service, oauth *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := validateOAuthState(r); err != nil {
			return
		}

		shared.DeleteCookie(w, r, oauthStateCookie)

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			return
		}

		token, err := oauth.Exchange(ctx, code)
		if err != nil {
			return
		}

		userInfo, err := fetchGoogleUserInfo(ctx, oauth, token)
		if err != nil {
			return
		}

		user, err := findOrCreateUser(ctx, service, userInfo)
		if err != nil {
			return
		}

		s.tokenAndSession(r.Context(), user.ID, r)
		gtoken := pkg.GenerateToken()

		newtoken, err := service.r.createSession(ctx, models.Session{
			UserID:    user.ID,
			Token:     gtoken,
			ExpiresAt: time.Now().Add(shared.TwoDays),
			IPAddress: pkg.GetIPAddressBytes(r),
			UserAgent: r.UserAgent(),
		})
		if err != nil {
			return
		}

		shared.SetSecureCookie(w, r, shared.SessionToken, newtoken, shared.TwoDays, true)

		http.Redirect(w, r, redirectPath, http.StatusSeeOther)
	}
}

func validateOAuthState(r *http.Request) error {
	stateCookie, err := r.Cookie(oauthStateCookie)
	if err != nil {
		return fmt.Errorf("missing state cookie: %w", err)
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam == "" || stateParam != stateCookie.Value {
		return fmt.Errorf("state mismatch")
	}

	return nil
}

func fetchGoogleUserInfo(ctx context.Context, oauth *oauth2.Config, token *oauth2.Token) (*models.GoogleUser, error) {
	client := oauth.Client(ctx, token)
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google API returned status %d", resp.StatusCode)
	}

	var userInfo models.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

func findOrCreateUser(ctx context.Context, service *service, userInfo *models.GoogleUser) (*models.User, error) {
	user, err := service.r.GetUserByEmail(ctx, userInfo.Email)
	switch err {
	case nil:
		if user.Oauth {
			return user, nil
		} else {
			err = service.r.setOauthToTrue(ctx, user.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to link Google account: %w", err)
			}

			// Refresh user data to get updated OAuth status
			user, err = service.r.GetUserByEmail(ctx, userInfo.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to refresh user data after linking: %w", err)
			}

			return user, nil
		}

	case sql.ErrNoRows:
		newUser := &models.User{
			Email: userInfo.Email,
			Oauth: true,
		}

		user, err = service.r.createUserWithGoogle(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create Google user: %w", err)
		}

		return user, nil

	default:
		return nil, fmt.Errorf("models error when looking up user: %w", err)
	}
}
