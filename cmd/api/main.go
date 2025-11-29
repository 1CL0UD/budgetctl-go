package main

import (
	"budgetctl-go/internal/server"
	"fmt"
)

// @title           BudgetCtl API
// @version         1.0
// @description     Personal Finance API built with Go and Echo.
// @host            localhost:8080
// @BasePath        /api
func main() {
	srv := server.NewServer()

	fmt.Println("Server listening on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
