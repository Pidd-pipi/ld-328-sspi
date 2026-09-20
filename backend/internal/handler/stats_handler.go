package handler

import (
	"bytes"
	"log/slog"
	"strconv"

	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

// StatsHandler 统计接口。
type StatsHandler struct {
	svc *service.StatsService
	log *slog.Logger
}

// NewStatsHandler 构造统计接口。
func NewStatsHandler(svc *service.StatsService, log *slog.Logger) *StatsHandler {
	return &StatsHandler{svc: svc, log: log}
}

// Dashboard 看板数据。
func (h *StatsHandler) Dashboard(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	data, err := h.svc.Dashboard(c.Request.Context(), userID(c), familyID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, data)
}

// Statistics 分类统计报表。
func (h *StatsHandler) Statistics(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	month := c.Query("month")
	data, err := h.svc.Statistics(c.Request.Context(), userID(c), familyID, month)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, data)
}

// ExportPDF 导出统计 PDF。
func (h *StatsHandler) ExportPDF(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	month := c.Query("month")
	data, err := h.svc.Statistics(c.Request.Context(), userID(c), familyID, month)
	if err != nil {
		c.Error(err)
		return
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("CyFreshFood 分类统计报表", false)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "CyFreshFood Category Statistics")
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 8, "Month: "+month)
	pdf.Ln(8)
	pdf.Cell(0, 8, "Waste Amount Estimate: "+util.FormatMoney(data.WasteAmount)+" CNY")
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Category Share:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 11)
	for _, row := range data.CategoryShare {
		pdf.Cell(0, 7, util.CategoryText(row.Category)+"  count="+strconv.FormatInt(row.Count, 10)+"  qty="+util.FormatMoney(row.TotalQuantity))
		pdf.Ln(7)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.Error(util.InternalError("PDF 生成（Statistics）失败", err))
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="statistics.pdf"`)
	c.Data(200, "application/pdf", buf.Bytes())
}
