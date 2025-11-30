package routes

import (
	"budgetctl-go/internal/auth"
	"budgetctl-go/internal/database/gensql"
	authmw "budgetctl-go/internal/server/middleware"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

const testPasetoKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

type mockUserStore struct {
	user gensql.User
	err  error
}

func (m *mockUserStore) GetUserByID(ctx context.Context, id int64) (gensql.User, error) {
	if m.err != nil {
		return gensql.User{}, m.err
	}
	return m.user, nil
}

func setupAuthTestServer(store authmw.UserStore) *echo.Echo {
	e := echo.New()
	e.GET("/auth/me", getCurrentUser(), authmw.AuthMiddleware(store))
	return e
}

func TestAuthMeSuccess(t *testing.T) {
	os.Setenv("PASETO_KEY", testPasetoKey)

	name := "Test User"
	avatar := "http://example.com/avatar.png"
	store := &mockUserStore{
		user: gensql.User{
			ID:          42,
			Email:       "test@example.com",
			Name:        &name,
			AvatarUrl:   &avatar,
			Preferences: []byte(`{"theme":"dark"}`),
		},
	}

	token, err := auth.GenerateToken(store.user.ID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	e := setupAuthTestServer(store)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID          int64           `json:"id"`
		Name        *string         `json:"name"`
		Email       string          `json:"email"`
		AvatarURL   *string         `json:"avatarUrl"`
		Preferences json.RawMessage `json:"preferences"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.ID != store.user.ID || resp.Email != store.user.Email {
		t.Fatalf("unexpected user in response: %+v", resp)
	}
	if resp.Name == nil || *resp.Name != name {
		t.Fatalf("expected name %q, got %#v", name, resp.Name)
	}
	if resp.AvatarURL == nil || *resp.AvatarURL != avatar {
		t.Fatalf("expected avatar %q, got %#v", avatar, resp.AvatarURL)
	}
	if string(resp.Preferences) != `{"theme":"dark"}` {
		t.Fatalf("unexpected preferences: %s", string(resp.Preferences))
	}
}

func TestAuthMeMissingCookie(t *testing.T) {
	os.Setenv("PASETO_KEY", testPasetoKey)

	e := setupAuthTestServer(&mockUserStore{
		err: pgx.ErrNoRows,
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "unauthorized" {
		t.Fatalf("expected error \"unauthorized\", got %q", resp["error"])
	}
}

func TestAuthMeInvalidToken(t *testing.T) {
	os.Setenv("PASETO_KEY", testPasetoKey)

	e := setupAuthTestServer(&mockUserStore{
		err: pgx.ErrNoRows,
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "not-a-valid-token"})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "unauthorized" {
		t.Fatalf("expected error \"unauthorized\", got %q", resp["error"])
	}
}
