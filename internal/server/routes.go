package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", s.HelloWorldHandler)
	e.GET("/health", s.healthHandler)
	api := e.Group("/api")
	api.GET("/helloworld", s.HelloWorldHandler)

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Hello World"})
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
