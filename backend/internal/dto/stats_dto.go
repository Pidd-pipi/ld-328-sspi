package dto

// Pagination 通用分页参数。
type Pagination struct {
	Page     int `form:"page" binding:"gte=1"`
	PageSize int `form:"page_size" binding:"gte=1,lte=200"`
}

// DefaultPage 默认页码。
const DefaultPage = 1

// DefaultPageSize 默认每页数量。
const DefaultPageSize = 20
