package routes

import (
	"budgetctl-go/internal/database"
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterHealth(api huma.API, db database.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health Check",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *struct{}) (*struct {
		Body map[string]string `json:"body"`
	}, error) {

		status := db.Health()

		return &struct {
			Body map[string]string `json:"body"`
		}{
			Body: status,
		}, nil
	})
}
