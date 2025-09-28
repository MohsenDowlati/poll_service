package controller

import (
	"strconv"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

func extractPagination(c *gin.Context) domain.PaginationQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize)))

	query := domain.PaginationQuery{Page: page, PageSize: pageSize}
	query.Normalize(defaultPageSize, maxPageSize)

	return query
}
