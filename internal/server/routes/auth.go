package routes

import (
	"budgetctl-go/internal/database"
	"budgetctl-go/internal/database/gensql"
	"database/sql"
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
	q := c.Request().URL.Query()
	q.Add("provider", "google")
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

	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173/dashboard")
}

func RegisterAuthRoutes(e *echo.Echo, db database.Service) {
	e.GET("/auth/:provider", beginAuth)
	e.GET("/auth/:provider/callback", func(c echo.Context) error {
		return completeAuth(c, db)
	})
}
