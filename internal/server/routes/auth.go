package routes

import (
	"budgetctl-go/internal/auth"
	"budgetctl-go/internal/database"
	"budgetctl-go/internal/database/gensql"
	"budgetctl-go/internal/server/middleware"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
)

func beginAuth(c echo.Context) error {
	provider := c.Param("provider")
	if provider == "" {
		provider = "google"
	}
	q := c.Request().URL.Query()
	q.Set("provider", provider)
	c.Request().URL.RawQuery = q.Encode()

	// Start the redirect to Google
	gothic.BeginAuthHandler(c.Response(), c.Request())
	return nil
}

func completeAuth(c echo.Context, db database.Service) error {
	// 1. Exchange the code for a User Profile
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if user.Email == "" {
		return c.String(http.StatusBadRequest, "email not provided by OAuth provider")
	}

	fmt.Printf("User Logged In: %s (%s)\n", user.Name, user.Email)

	ctx := c.Request().Context()
	queries := db.GetQueries()

	dbUser, err := queries.GetUserByEmail(ctx, user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			params := gensql.CreateUserParams{
				Email:        user.Email,
				PasswordHash: "google_oauth_user",
				Name:         optionalString(user.Name),
				AvatarUrl:    optionalString(user.AvatarURL),
			}

			dbUser, err = queries.CreateUser(ctx, params)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Failed to create user: "+err.Error())
			}
			fmt.Printf("Created new user: %s\n", dbUser.Email)
		} else {
			return c.String(http.StatusInternalServerError, "Database error: "+err.Error())
		}
	} else {
		fmt.Printf("Logged in existing user: %s\n", dbUser.Email)
	}

	token, err := auth.GenerateToken(dbUser.ID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate token")
	}

	cookieConfig := cookieSecurityConfig()

	c.SetCookie(&http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   cookieConfig.secure,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: cookieConfig.sameSite,
	})

	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173/")
}

func RegisterAuthRoutes(e *echo.Echo, db database.Service) {
	userStore := db.GetQueries()
	authMiddleware := middleware.AuthMiddleware(userStore)

	// Entry points for starting OAuth (plan calls for /auth/login/google; keep /auth/:provider for compatibility)
	e.GET("/auth/login/:provider", beginAuth)
	e.GET("/auth/:provider", beginAuth)
	e.GET("/auth/:provider/callback", func(c echo.Context) error {
		return completeAuth(c, db)
	})
	e.POST("/auth/logout", logout)
	e.GET("/auth/me", getCurrentUser(), authMiddleware)
}

func logout(c echo.Context) error {
	cookieConfig := cookieSecurityConfig()
	// Clear auth cookie
	c.SetCookie(&http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cookieConfig.secure,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: cookieConfig.sameSite,
	})

	return c.NoContent(http.StatusNoContent)
}

func getCurrentUser() echo.HandlerFunc {
	type response struct {
		ID          int64           `json:"id"`
		Name        *string         `json:"name"`
		Email       string          `json:"email"`
		AvatarURL   *string         `json:"avatarUrl"`
		Preferences json.RawMessage `json:"preferences"`
	}

	return func(c echo.Context) error {
		user, ok := middleware.UserFromContext(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error":   "unauthorized",
				"message": "Not authenticated",
			})
		}

		prefs := user.Preferences
		if len(prefs) == 0 {
			prefs = []byte("{}")
		}

		return c.JSON(http.StatusOK, response{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			AvatarURL:   user.AvatarUrl,
			Preferences: json.RawMessage(prefs),
		})
	}
}

// optionalString returns a pointer to the string if non-empty, otherwise nil.
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type cookieConfig struct {
	secure   bool
	sameSite http.SameSite
}

func cookieSecurityConfig() cookieConfig {
	isProd := os.Getenv("APP_ENV") == "production" || os.Getenv("ENV") == "production"

	cfg := cookieConfig{
		secure:   isProd,
		sameSite: http.SameSiteLaxMode,
	}

	if isProd {
		// For cross-site redirects (OAuth) choose None+Secure, otherwise stay Lax for local dev.
		cfg.sameSite = http.SameSiteNoneMode
	}

	return cfg
}
