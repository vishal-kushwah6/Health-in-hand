package common

const (
	DefaultPage  = 1
	DefaultLimit = 20
)

// PaginationQuery is embedded in list-endpoint query structs.
// Call Normalize() after validation to fill defaults.
type PaginationQuery struct {
	Page  int `json:"page"  form:"page"  validate:"omitempty,gte=1"`
	Limit int `json:"limit" form:"limit" validate:"omitempty,gte=1,lte=50"`
}

func (p *PaginationQuery) Normalize() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.Limit < 1 {
		p.Limit = DefaultLimit
	}
}

// Skip is the Mongo skip value for the current page.
func (p PaginationQuery) Skip() int64 { return int64((p.Page - 1) * p.Limit) }

// PaginatedResponse is the standard list response.
type PaginatedResponse[T any] struct {
	Items []T   `json:"items"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}
