package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func GetPagination(c *gin.Context) Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	return Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p *Pagination) Paginate() (offset int, limit int) {
	offset = (p.Page - 1) * p.PageSize
	limit = p.PageSize
	return
}
