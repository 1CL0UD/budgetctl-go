package routes

import "math"

// PaginationInput is a reusable pagination input struct that can be embedded in any handler
type PaginationInput struct {
	Page    int `query:"page" default:"1" minimum:"1" doc:"Page number"`
	PerPage int `query:"per_page" default:"20" maximum:"100" minimum:"1" doc:"Items per page"`
}

// ToLimitOffset converts page/per_page to SQL limit/offset
func (p PaginationInput) ToLimitOffset() (int32, int32) {
	limit := int32(p.PerPage)
	offset := int32((p.Page - 1) * p.PerPage)
	return limit, offset
}

// PaginatedResponse is a generic wrapper for paginated responses
type PaginatedResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

// Meta contains pagination metadata
type Meta struct {
	CurrentPage int   `json:"current_page"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
	PerPage     int   `json:"per_page"`
}

// NewPaginatedResponse creates a new paginated response with the given data and metadata
func NewPaginatedResponse[T any](data []T, total int64, page, perPage int) *PaginatedResponse[T] {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	
	// Handle empty slice instead of null in JSON
	if data == nil {
		data = []T{}
	}

	return &PaginatedResponse[T]{
		Data: data,
		Meta: Meta{
			CurrentPage: page,
			TotalItems:  total,
			TotalPages:  totalPages,
			PerPage:     perPage,
		},
	}
}