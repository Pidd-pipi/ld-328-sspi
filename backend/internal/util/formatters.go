package util

import (
	"fmt"
	"strings"
	"time"
)

// formatters.go 同时包含日期格式化、剩余天数文本、新鲜度状态文本、类别文本、存放位置文本、角色文本。
// 多个 handler 与 service 直接引用本文件（屎山耦合点）。

const timeLayout = "2006-01-02 15:04:05"

// FormatTime 时间格式化为字符串。
func FormatTime(t time.Time) string { return t.Format(timeLayout) }

// FormatDate 日期格式化为字符串。
func FormatDate(t time.Time) string { return t.Format("2006-01-02") }

// FormatMoney 金额保留两位小数。
func FormatMoney(v float64) string { return fmt.Sprintf("%.2f", v) }

// RemainingDaysText 根据剩余天数生成文案。
func RemainingDaysText(days int) string {
	switch {
	case days < 0:
		return fmt.Sprintf("已过期 %d 天", -days)
	case days == 0:
		return "今天到期"
	case days <= 3:
		return fmt.Sprintf("剩余 %d 天（临期）", days)
	default:
		return fmt.Sprintf("剩余 %d 天", days)
	}
}

// FreshnessStatusText 新鲜度状态中文文案。
func FreshnessStatusText(status string) string {
	switch status {
	case "fresh":
		return "充裕"
	case "expiring":
		return "临期"
	case "expired":
		return "已过期"
	case "consumed":
		return "已消耗"
	default:
		return "未知"
	}
}

// CategoryText 食品类别中文文案。
func CategoryText(category string) string {
	switch category {
	case "fresh":
		return "生鲜"
	case "dairy":
		return "乳制品"
	case "cooked":
		return "熟食"
	case "bakery":
		return "烘焙"
	case "frozen":
		return "冷冻"
	case "other":
		return "其他"
	default:
		return "其他"
	}
}

// StorageLocationText 存放位置中文文案。
func StorageLocationText(location string) string {
	switch location {
	case "fridge":
		return "冰箱"
	case "pantry":
		return "储藏室"
	case "freezer":
		return "冷冻室"
	case "counter":
		return "台面"
	case "other":
		return "其他"
	default:
		return "其他"
	}
}

// RoleText 角色中文文案。
func RoleText(role string) string {
	switch role {
	case "admin":
		return "管理员"
	case "member":
		return "成员"
	default:
		return "成员"
	}
}

// JoinNames 拼接名称列表。
func JoinNames(names []string) string { return strings.Join(names, "、") }
