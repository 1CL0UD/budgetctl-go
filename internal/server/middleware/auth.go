package middleware

import (
	"budgetctl-go/internal/auth"
	"budgetctl-go/internal/database/gensql"
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

const (
	userContextKey   = "auth_user"
	userIDContextKey = "auth_user_id"
)

// UserStore is the minimal interface needed for auth middleware to load users.
type UserStore interface {
	GetUserByID(ctx context.Context, id int64) (gensql.User, error)
}

// AuthMiddleware validates the auth_token cookie, loads the user, and stores it on the context.
func AuthMiddleware(store UserStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("auth_token")
			if err != nil || cookie.Value == "" {
				return unauthorized(c, "Not authenticated")
			}

			userID, err := auth.ParseToken(cookie.Value)
			if err != nil {
				return unauthorized(c, "Invalid or expired session token")
			}

			ctx := c.Request().Context()
			user, err := store.GetUserByID(ctx, userID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
					return unauthorized(c, "User not found")
				}
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error":   "database_error",
					"message": "Failed to load user",
				})
			}

			c.Set(userContextKey, user)
			c.Set(userIDContextKey, userID)

			return next(c)
		}
	}
}

// UserFromContext retrieves the authenticated user set by AuthMiddleware.
func UserFromContext(c echo.Context) (gensql.User, bool) {
	val := c.Get(userContextKey)
	if val == nil {
		return gensql.User{}, false
	}

	user, ok := val.(gensql.User)
	return user, ok
}

// UserIDFromContext retrieves the authenticated user ID set by AuthMiddleware.
func UserIDFromContext(c echo.Context) (int64, bool) {
	val := c.Get(userIDContextKey)
	if val == nil {
		return 0, false
	}

	userID, ok := val.(int64)
	return userID, ok
}

func unauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{
		"error":   "unauthorized",
		"message": message,
	})
}
