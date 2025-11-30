package routes

import (
	"context"
	"net/http"
	"time"

	"budgetctl-go/internal/database"
	"budgetctl-go/internal/database/gensql"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// Transaction Input/Output Types for Huma

type ListTransactionsRequest struct {
	PaginationInput
	Search     string   `query:"search" doc:"Search in description, tags, notes"`
	DateFrom   string   `query:"date_from" doc:"Filter by date from (YYYY-MM-DD)"`
	DateTo     string   `query:"date_to" doc:"Filter by date to (YYYY-MM-DD)"`
	Categories []string `query:"category" doc:"Filter by categories"`
	Type       string   `query:"type" doc:"Filter by type (income|expense|all)"`
	MinAmount  float64  `query:"min_amount" doc:"Minimum amount filter"`
	MaxAmount  float64  `query:"max_amount" doc:"Maximum amount filter"`
	Tags       []string `query:"tag" doc:"Filter by tags"`
}

type ListTransactionsResponse struct {
	Body *PaginatedResponse[gensql.Transaction]
}

type GetTransactionRequest struct {
	ID int64 `path:"id" doc:"Transaction ID"`
}

type GetTransactionResponse struct {
	Body *gensql.Transaction
}

type CreateTransactionRequest struct {
	Body gensql.CreateTransactionParams
}

type CreateTransactionResponse struct {
	Body *gensql.Transaction
}

type UpdateTransactionRequest struct {
	ID   int64                           `path:"id" doc:"Transaction ID"`
	Body gensql.UpdateTransactionParams
}

type UpdateTransactionResponse struct {
	Body *gensql.Transaction
}

type DeleteTransactionRequest struct {
	ID int64 `path:"id" doc:"Transaction ID"`
}

type GetCategoriesResponse struct {
	Body []gensql.GetCategoriesRow
}

type GetTagsResponse struct {
	Body []gensql.GetTagsRow
}

// Huma Handlers

func RegisterTransactionRoutes(api huma.API, db database.Service) {
	// List Transactions
	huma.Register(api, huma.Operation{
		OperationID: "list-transactions",
		Method:      http.MethodGet,
		Path:        "/transactions",
		Summary:     "List Transactions",
		Tags:        []string{"Transactions"},
	}, func(ctx context.Context, input *ListTransactionsRequest) (*ListTransactionsResponse, error) {
		user, err := getUserFromContext(ctx)
		if err != nil {
			return nil, err
		}

		queries := db.GetQueries()
		limit, offset := input.ToLimitOffset()

		// Build filter parameters
		params := gensql.ListTransactionsWithFiltersParams{
			UserID:  user.ID,
			Column2: input.Search,
			Limit:   limit,
			Offset:  offset,
		}

		// Parse date filters
		if input.DateFrom != "" {
			if d, err := time.Parse("2006-01-02", input.DateFrom); err == nil {
				params.Column3 = pgtype.Date{Time: d, Valid: true}
			}
		}
		if input.DateTo != "" {
			if d, err := time.Parse("2006-01-02", input.DateTo); err == nil {
				params.Column4 = pgtype.Date{Time: d, Valid: true}
			}
		}

		// Parse array filters
		if len(input.Categories) > 0 {
			params.Column5 = input.Categories
		}
		if input.Type != "" && input.Type != "all" {
			params.Column6 = input.Type
		}
		if len(input.Tags) > 0 {
			params.Column9 = input.Tags
		}

		// Parse amount filters
		if input.MinAmount > 0 {
			params.Column7 = pgtype.Numeric{Valid: true}
			params.Column7.Scan(input.MinAmount)
		}
		if input.MaxAmount > 0 {
			params.Column8 = pgtype.Numeric{Valid: true}
			params.Column8.Scan(input.MaxAmount)
		}

		// Fetch data
		transactions, err := queries.ListTransactionsWithFilters(ctx, params)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to fetch transactions", err)
		}

		// Fetch count for pagination
		countParams := gensql.CountTransactionsParams{
			UserID:  user.ID,
			Column2: params.Column2,
			Column3: params.Column3,
			Column4: params.Column4,
			Column5: params.Column5,
			Column6: params.Column6,
			Column7: params.Column7,
			Column8: params.Column8,
			Column9: params.Column9,
		}

		total, err := queries.CountTransactions(ctx, countParams)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to count transactions", err)
		}

		return &ListTransactionsResponse{
			Body: NewPaginatedResponse(transactions, total, input.Page, input.PerPage),
		}, nil
	})

	// Get Transaction
	huma.Register(api, huma.Operation{
		OperationID: "get-transaction",
		Method:      http.MethodGet,
		Path:        "/transactions/{id}",
		Summary:     "Get Transaction",
		Tags:        []string{"Transactions"},
	}, func(ctx context.Context, input *GetTransactionRequest) (*GetTransactionResponse, error) {
		user, err := getUserFromContext(ctx)
		if err != nil {
			return nil, err
		}

		queries := db.GetQueries()
		transaction, err := queries.GetTransactionByID(ctx, gensql.GetTransactionByIDParams{
			ID:     input.ID,
			UserID: user.ID,
		})
		if err != nil {
			return nil, huma.Error404NotFound("Transaction not found", err)
		}

		return &GetTransactionResponse{Body: &transaction}, nil
	})

	// Create Transaction
	huma.Register(api, huma.Operation{
		OperationID: "create-transaction",
		Method:      http.MethodPost,
		Path:        "/transactions",
		Summary:     "Create Transaction",
		Tags:        []string{"Transactions"},
	}, func(ctx context.Context, input *CreateTransactionRequest) (*CreateTransactionResponse, error) {
		user, err := getUserFromContext(ctx)
		if err != nil {
			return nil, err
		}

		// Set required fields and defaults
		params := input.Body
		params.UserID = user.ID
		if params.Currency == "" {
			params.Currency = "USD"
		}
		if params.Status == "" {
			params.Status = "pending"
		}
		if params.Account == "" {
			params.Account = ""
		}
		if !params.Date.Valid {
			params.Date = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		}

		queries := db.GetQueries()
		transaction, err := queries.CreateTransaction(ctx, params)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to create transaction", err)
		}

		return &CreateTransactionResponse{Body: &transaction}, nil
	})

	// Update Transaction
	huma.Register(api, huma.Operation{
		OperationID: "update-transaction",
		Method:      http.MethodPatch,
		Path:        "/transactions/{id}",
		Summary:     "Update Transaction",
		Tags:        []string{"Transactions"},
	}, func(ctx context.Context, input *UpdateTransactionRequest) (*UpdateTransactionResponse, error) {
		user, err := getUserFromContext(ctx)
		if err != nil {
			return nil, err
		}

		// Set required fields
		params := input.Body
		params.ID = input.ID
		params.UserID = user.ID

		queries := db.GetQueries()
		transaction, err := queries.UpdateTransaction(ctx, params)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to update transaction", err)
		}

		return &UpdateTransactionResponse{Body: &transaction}, nil
	})

	// Delete Transaction
	huma.Register(api, huma.Operation{
		OperationID: "delete-transaction",
		Method:      http.MethodDelete,
		Path:        "/transactions/{id}",
		Summary:     "Delete Transaction",
		Tags:        []string{"Transactions"},
	}, func(ctx context.Context, input *DeleteTransactionRequest) (*struct{}, error) {
		user, err := getUserFromContext(ctx)
		if err != nil {
			return nil, err
		}

		queries := db.GetQueries()
		err = queries.DeleteTransaction(ctx, gensql.DeleteTransactionParams{
			ID:     input.ID,
			UserID: user.ID,
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to delete transaction", err)
		}

		return nil, nil
	})

	// Get Categories
	huma.Register(api, huma.Operation{
		OperationID: "get-categories",
		Method:      http.MethodGet,
		Path:        "/categories",
		Summary:     "Get Categories",
		Tags:        []string{"Transactions"},
	}, func(ctx context.Context, input *struct{}) (*GetCategoriesResponse, error) {
		user, err := getUserFromContext(ctx)
		if err != nil {
			return nil, err
		}

		queries := db.GetQueries()
		categories, err := queries.GetCategories(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to fetch categories", err)
		}

		return &GetCategoriesResponse{Body: categories}, nil
	})

	// Get Tags
	huma.Register(api, huma.Operation{
		OperationID: "get-tags",
		Method:      http.MethodGet,
		Path:        "/tags",
		Summary:     "Get Tags",
		Tags:        []string{"Transactions"},
	}, func(ctx context.Context, input *struct{}) (*GetTagsResponse, error) {
		user, err := getUserFromContext(ctx)
		if err != nil {
			return nil, err
		}

		queries := db.GetQueries()
		tags, err := queries.GetTags(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to fetch tags", err)
		}

		return &GetTagsResponse{Body: tags}, nil
	})
}

// Helper functions

func getUserFromContext(ctx context.Context) (*gensql.User, error) {
	// Since we're using Huma with Echo adapter, we need to extract the user
	// from the Echo context that was set by the auth middleware
	// For now, let's implement a simple approach - in a real implementation,
	// we'd need to properly bridge Echo and Huma contexts
	
	// This is a temporary solution - we'll need to properly integrate
	// the Echo auth middleware with Huma
	return &gensql.User{
		ID:    1, // TODO: Get real user ID from auth context
		Email: "user@example.com", // TODO: Get real user from auth context
	}, nil
}
