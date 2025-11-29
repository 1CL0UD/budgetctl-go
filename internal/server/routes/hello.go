package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type HelloInput struct {
	Name string `query:"name" maxLength:"20" default:"World" doc:"Name to greet"`
}

type HelloOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, World!"`
	}
}

func RegisterHello(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-hello",
		Method:      http.MethodGet,
		Path:        "/hello",
		Summary:     "Say Hello",
		Description: "Returns a personalized greeting.",
		Tags:        []string{"Intro"},
	}, helloHandler)
}

func helloHandler(ctx context.Context, input *HelloInput) (*HelloOutput, error) {
	resp := &HelloOutput{}

	resp.Body.Message = fmt.Sprintf("Hello, %s!", input.Name)

	return resp, nil
}
