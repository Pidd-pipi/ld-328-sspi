package handler

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/blueship581/cyfreshfood/internal/dto"
	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// FoodItemHandler 食品接口。
type FoodItemHandler struct {
	svc    *service.FoodItemService
	consumeSvc *service.ConsumptionRecordService
	log    *slog.Logger
}

// NewFoodItemHandler 构造食品接口。
func NewFoodItemHandler(svc *service.FoodItemService, consumeSvc *service.ConsumptionRecordService, log *slog.Logger) *FoodItemHandler {
	return &FoodItemHandler{svc: svc, consumeSvc: consumeSvc, log: log}
}

// Create 录入食品。
func (h *FoodItemHandler) Create(c *gin.Context) {
	var req dto.CreateFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("食品（FoodItem）参数不合法", err))
		return
	}
	item, err := h.svc.Create(c.Request.Context(), userID(c), service.CreateFoodInput{
		FamilyID: req.FamilyID, Name: req.Name, Category: req.Category,
		ProductionDate: req.ProductionDate, ShelfLifeDays: req.ShelfLifeDays,
		Quantity: req.Quantity, Unit: req.Unit, StorageLocation: req.StorageLocation,
		OpenedAt: req.OpenedAt, ImageURL: req.ImageURL,
	})
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, item)
}

// List 食品列表。
func (h *FoodItemHandler) List(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	page := parseQueryInt(c.Query("page"), dto.DefaultPage)
	pageSize := parseQueryInt(c.Query("page_size"), dto.DefaultPageSize)
	items, total, err := h.svc.List(c.Request.Context(), userID(c), familyID,
		c.Query("category"), c.Query("status"), c.Query("storage_location"), c.Query("keyword"), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}

// Detail 食品详情（含消耗历史）。
func (h *FoodItemHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	item, err := h.svc.GetByID(c.Request.Context(), userID(c), id)
	if err != nil {
		c.Error(err)
		return
	}
	records, err := h.consumeSvc.ListByFood(c.Request.Context(), userID(c), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"item": item, "consumption_records": records})
}

// Update 编辑食品。
func (h *FoodItemHandler) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.CreateFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("食品（FoodItem）参数不合法", err))
		return
	}
	item, err := h.svc.Update(c.Request.Context(), userID(c), id, service.CreateFoodInput{
		FamilyID: req.FamilyID, Name: req.Name, Category: req.Category,
		ProductionDate: req.ProductionDate, ShelfLifeDays: req.ShelfLifeDays,
		Quantity: req.Quantity, Unit: req.Unit, StorageLocation: req.StorageLocation,
		OpenedAt: req.OpenedAt, ImageURL: req.ImageURL,
	})
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, item)
}

// Delete 删除食品。
func (h *FoodItemHandler) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := h.svc.Delete(c.Request.Context(), userID(c), id); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"deleted": id})
}

// Consume 消耗食品。
func (h *FoodItemHandler) Consume(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.ConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("消耗参数（ConsumptionRecord）不合法", err))
		return
	}
	consumedAt := req.ConsumedAt
	if consumedAt.IsZero() {
		consumedAt = time.Now()
	}
	record, err := h.svc.Consume(c.Request.Context(), userID(c), id, req.Quantity, consumedAt)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, record)
}

// CSVImport CSV 批量导入。
func (h *FoodItemHandler) CSVImport(c *gin.Context) {
	var req dto.CSVImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("CSV 导入（FoodItem）参数不合法", err))
		return
	}
	count, items, err := h.svc.ImportCSV(c.Request.Context(), userID(c), req.FamilyID, req.CSVText)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, gin.H{"imported": count, "items": items})
}

func parseQueryInt(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil && v > 0 {
		return v
	}
	return def
}
