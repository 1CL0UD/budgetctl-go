package main

import (
	"budgetctl-go/internal/server"
	"fmt"
)

func main() {
	srv := server.NewServer()

	fmt.Println("Server listening on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
