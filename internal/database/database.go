package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"

	// Import the generated package
	"budgetctl-go/internal/database/gensql"
)

type Service interface {
	Health() map[string]string
	Close()
	GetQueries() *gensql.Queries
}

type service struct {
	db *pgxpool.Pool
	*gensql.Queries
}

var (
	dbInstance *service
)

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	databaseUrl := os.Getenv("DATABASE_URL")
	db, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance = &service{
		db:      db,
		Queries: gensql.New(db),
	}
	return dbInstance
}

func (s *service) GetQueries() *gensql.Queries {
	return s.Queries
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := s.db.Ping(ctx)
	if err != nil {
		log.Printf("Database connection error: %v", err)
		return map[string]string{
			"status": "down",
			"error":  fmt.Sprintf("%v", err),
		}
	}

	return map[string]string{
		"status":  "up",
		"message": "It's healthy",
	}
}

func (s *service) Close() {
	s.db.Close()
}
