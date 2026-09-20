package handler

import (
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/dto"
	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// DisposalRequestHandler 临期处置申请接口。
type DisposalRequestHandler struct {
	svc *service.DisposalRequestService
	log *slog.Logger
}

// NewDisposalRequestHandler 构造处置申请接口。
func NewDisposalRequestHandler(svc *service.DisposalRequestService, log *slog.Logger) *DisposalRequestHandler {
	return &DisposalRequestHandler{svc: svc, log: log}
}

// Create 提交处置申请。
func (h *DisposalRequestHandler) Create(c *gin.Context) {
	var req dto.CreateDisposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("临期处置申请（DisposalRequest）参数不合法", err))
		return
	}
	result, err := h.svc.Create(c.Request.Context(), userID(c), service.CreateDisposalInput{
		FoodItemID: req.FoodItemID,
		Quantity:   req.Quantity,
		Method:     req.Method,
		Reason:     req.Reason,
	})
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, result)
}

// List 处置申请列表（待处理与全部申请状态）。
func (h *DisposalRequestHandler) List(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	foodItemID := parseUint(c.Query("food_item_id"))
	status := c.Query("status")
	page := parseQueryInt(c.Query("page"), dto.DefaultPage)
	pageSize := parseQueryInt(c.Query("page_size"), dto.DefaultPageSize)
	list, total, err := h.svc.List(c.Request.Context(), userID(c), familyID, foodItemID, status, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: list, Total: total, Page: page, Size: pageSize})
}

// Detail 处置申请详情。
func (h *DisposalRequestHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	result, err := h.svc.GetByID(c.Request.Context(), userID(c), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Approve 批准处置申请（仅家庭管理员）。
func (h *DisposalRequestHandler) Approve(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.ReviewDisposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("处置申请审批（DisposalRequest）参数不合法", err))
		return
	}
	result, err := h.svc.Approve(c.Request.Context(), userID(c), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Reject 驳回处置申请（仅家庭管理员；只关闭申请，不改库存）。
func (h *DisposalRequestHandler) Reject(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.ReviewDisposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("处置申请驳回（DisposalRequest）参数不合法", err))
		return
	}
	result, err := h.svc.Reject(c.Request.Context(), userID(c), id, req.Remark)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}
