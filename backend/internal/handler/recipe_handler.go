package handler

import (
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// RecipeHandler 食谱接口。
type RecipeHandler struct {
	svc *service.RecipeService
	log *slog.Logger
}

// NewRecipeHandler 构造食谱接口。
func NewRecipeHandler(svc *service.RecipeService, log *slog.Logger) *RecipeHandler {
	return &RecipeHandler{svc: svc, log: log}
}

// Recommendations 智能推荐。
func (h *RecipeHandler) Recommendations(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	result, err := h.svc.Recommend(c.Request.Context(), userID(c), familyID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// List 全部食谱。
func (h *RecipeHandler) List(c *gin.Context) {
	recipes, err := h.svc.ListAll(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, recipes)
}

// Detail 食谱详情。
func (h *RecipeHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	recipe, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, recipe)
}
