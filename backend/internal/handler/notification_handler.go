package handler

import (
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/dto"
	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知接口。
type NotificationHandler struct {
	svc *service.NotificationService
	log *slog.Logger
}

// NewNotificationHandler 构造通知接口。
func NewNotificationHandler(svc *service.NotificationService, log *slog.Logger) *NotificationHandler {
	return &NotificationHandler{svc: svc, log: log}
}

// List 通知列表。
func (h *NotificationHandler) List(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	page := parseQueryInt(c.Query("page"), dto.DefaultPage)
	pageSize := parseQueryInt(c.Query("page_size"), dto.DefaultPageSize)
	unreadOnly := c.Query("unread_only") == "true"
	items, total, err := h.svc.List(c.Request.Context(), userID(c), familyID, unreadOnly, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}

// MarkRead 标记单条已读。
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := h.svc.MarkRead(c.Request.Context(), userID(c), id); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"id": id, "read": true})
}

// MarkAllRead 标记全部已读。
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	if err := h.svc.MarkAllRead(c.Request.Context(), userID(c), familyID); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"read": true})
}

// UnreadCount 未读数量。
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	familyID := parseUint(c.Query("family_id"))
	count, err := h.svc.UnreadCount(c.Request.Context(), userID(c), familyID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"unread": count})
}
