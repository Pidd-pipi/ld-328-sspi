package constants

// NotificationType 通知类型枚举。
const (
	NotificationExpiring = "expiring" // 临期提醒
	NotificationExpired  = "expired"  // 过期提醒
)

// NotificationTypeList 全部通知类型。
var NotificationTypeList = []string{NotificationExpiring, NotificationExpired}

// ReminderScanLockKey Redis 临期扫描分布式锁键。
const ReminderScanLockKey = "cyfreshfood:reminder:scan_lock"

// ReminderScanTickerName 定时任务名称（日志模板引用）。
const ReminderScanTickerName = "reminder_scan"
