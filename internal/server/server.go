package server

import (
	"fmt"
	"net/http"
	"time"

	"budgetctl-go/internal/database"
	"budgetctl-go/internal/server/routes"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// Declare Server config
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

	config := huma.DefaultConfig("BudgetCtl API", "1.0.0")
	api := humaecho.New(e, config)

	routes.RegisterHealth(api, s.db)
	routes.RegisterHello(api)

	return e
}
