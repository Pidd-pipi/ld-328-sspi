package handler

import (
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/dto"
	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// ConsumptionRecordHandler 消耗记录接口。
type ConsumptionRecordHandler struct {
	svc *service.ConsumptionRecordService
	log *slog.Logger
}

// NewConsumptionRecordHandler 构造消耗记录接口。
func NewConsumptionRecordHandler(svc *service.ConsumptionRecordService, log *slog.Logger) *ConsumptionRecordHandler {
	return &ConsumptionRecordHandler{svc: svc, log: log}
}

// List 消耗记录列表。
func (h *ConsumptionRecordHandler) List(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	page := parseQueryInt(c.Query("page"), dto.DefaultPage)
	pageSize := parseQueryInt(c.Query("page_size"), dto.DefaultPageSize)
	records, total, err := h.svc.List(c.Request.Context(), userID(c), familyID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: records, Total: total, Page: page, Size: pageSize})
}

// Analysis 消耗频率分析。
func (h *ConsumptionRecordHandler) Analysis(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	month := c.Query("month")
	result, err := h.svc.Analysis(c.Request.Context(), userID(c), familyID, month)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}
