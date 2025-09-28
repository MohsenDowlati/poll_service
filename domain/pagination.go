package domain

// PaginationQuery captures paging preferences supplied via query params.
type PaginationQuery struct {
	Page     int
	PageSize int
}

// Normalize ensures sane defaults and caps the requested page size.
func (p *PaginationQuery) Normalize(defaultPageSize, maxPageSize int) {
	if p.Page < 1 {
		p.Page = 1
	}

	if p.PageSize <= 0 {
		p.PageSize = defaultPageSize
	}

	if maxPageSize > 0 && p.PageSize > maxPageSize {
		p.PageSize = maxPageSize
	}
}

// Skip returns the number of documents to skip for the given page.
func (p PaginationQuery) Skip() int64 {
	if p.Page < 1 {
		return 0
	}
	return int64((p.Page - 1) * p.PageSize)
}

// Limit returns the page size as an int64 for Mongo options.
func (p PaginationQuery) Limit() int64 {
	if p.PageSize < 1 {
		return 0
	}
	return int64(p.PageSize)
}

// PaginationResult describes the paging information returned to clients.
type PaginationResult struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginationResult computes pagination metadata for the given request and total.
func NewPaginationResult(query PaginationQuery, total int64) PaginationResult {
	totalPages := 0
	if query.PageSize > 0 {
		totalPages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}

	return PaginationResult{
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}
