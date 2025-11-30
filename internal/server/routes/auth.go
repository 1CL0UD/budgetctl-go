package routes

import (
	"budgetctl-go/internal/database"
	"budgetctl-go/internal/database/gensql"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"budgetctl-go/internal/auth"

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

	c.SetCookie(&http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // TODO: Set true in Prod (HTTPS)
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})

	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173/")
}

func RegisterAuthRoutes(e *echo.Echo, db database.Service) {
	// Entry points for starting OAuth (plan calls for /auth/login/google; keep /auth/:provider for compatibility)
	e.GET("/auth/login/:provider", beginAuth)
	e.GET("/auth/:provider", beginAuth)
	e.GET("/auth/:provider/callback", func(c echo.Context) error {
		return completeAuth(c, db)
	})
	e.POST("/auth/logout", logout)
	e.GET("/auth/me", getCurrentUser(db))
}

func logout(c echo.Context) error {
	// Clear auth cookie
	c.SetCookie(&http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // TODO: Set true in Prod (HTTPS)
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})

	return c.NoContent(http.StatusNoContent)
}

func getCurrentUser(db database.Service) echo.HandlerFunc {
	type response struct {
		ID          int64           `json:"id"`
		Name        *string         `json:"name"`
		Email       string          `json:"email"`
		AvatarURL   *string         `json:"avatarUrl"`
		Preferences json.RawMessage `json:"preferences"`
	}

	return func(c echo.Context) error {
		cookie, err := c.Cookie("auth_token")
		if err != nil || cookie.Value == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error":   "unauthorized",
				"message": "Not authenticated",
			})
		}

		userID, err := auth.ParseToken(cookie.Value)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error":   "invalid_token",
				"message": "Invalid or expired session token",
			})
		}

		ctx := c.Request().Context()
		user, err := db.GetQueries().GetUserByID(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "User not found",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "database_error",
				"message": "Failed to load user",
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
