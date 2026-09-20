package handler

import (
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/dto"
	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// DisposalApplicationHandler 临期处置申请接口。
type DisposalApplicationHandler struct {
	svc *service.DisposalApplicationService
	log *slog.Logger
}

// NewDisposalApplicationHandler 构造处置申请接口。
func NewDisposalApplicationHandler(svc *service.DisposalApplicationService, log *slog.Logger) *DisposalApplicationHandler {
	return &DisposalApplicationHandler{svc: svc, log: log}
}

// Create 提交临期处置申请。
func (h *DisposalApplicationHandler) Create(c *gin.Context) {
	var req dto.CreateDisposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("处置申请（DisposalApplication）参数不合法", err))
		return
	}
	app, err := h.svc.Create(c.Request.Context(), userID(c), service.CreateDisposalInput{
		FoodItemID: req.FoodItemID, Quantity: req.Quantity, Method: req.Method,
	})
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, app)
}

// List 处置申请列表（待处理与全部申请状态，刷新后可回读）。
func (h *DisposalApplicationHandler) List(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	foodItemID := parseUint(c.Query("food_item_id"))
	page := parseQueryInt(c.Query("page"), dto.DefaultPage)
	pageSize := parseQueryInt(c.Query("page_size"), dto.DefaultPageSize)
	apps, total, err := h.svc.List(c.Request.Context(), userID(c), familyID, c.Query("status"), foodItemID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: apps, Total: total, Page: page, Size: pageSize})
}

// Detail 处置申请详情。
func (h *DisposalApplicationHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	app, err := h.svc.GetByID(c.Request.Context(), userID(c), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, app)
}

// Approve 批准：扣减余量、生成消耗记录并结案。
func (h *DisposalApplicationHandler) Approve(c *gin.Context) {
	id := parseUint(c.Param("id"))
	req := dto.ReviewDisposalRequest{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(util.BadRequest("处置审批（DisposalApplication）参数不合法", err))
			return
		}
	}
	app, err := h.svc.Approve(c.Request.Context(), userID(c), id, req.Note)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, app)
}

// Reject 驳回：仅关闭申请，不改库存。
func (h *DisposalApplicationHandler) Reject(c *gin.Context) {
	id := parseUint(c.Param("id"))
	req := dto.ReviewDisposalRequest{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(util.BadRequest("处置审批（DisposalApplication）参数不合法", err))
			return
		}
	}
	app, err := h.svc.Reject(c.Request.Context(), userID(c), id, req.Note)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, app)
}
