package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"budgetctl-go/internal/database"
	"budgetctl-go/internal/server/routes"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type Server struct {
	port int
	db   database.Service
}

func NewServer() *http.Server {
	port := 8080
	NewServer := &Server{
		port: port,
		db:   database.New(),
	}

	store := sessions.NewCookieStore([]byte("secret_key"))
  gothic.Store = store

	goth.UseProviders(
        google.New(
            os.Getenv("GOOGLE_CLIENT_ID"),    
            os.Getenv("GOOGLE_CLIENT_SECRET"),
            "http://localhost:8080/auth/google/callback", 
        ),
    )

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	config := huma.DefaultConfig("BudgetCtl API", "1.0.0")
	api := humaecho.New(e, config)

	routes.RegisterHealth(api, s.db)
	routes.RegisterHello(api)

	routes.RegisterAuthRoutes(e, s.db)
	routes.RegisterTransactionRoutes(api, s.db)

	return e
}
