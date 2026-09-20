package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config 集中解析全部环境变量：数据库、Redis、JWT、限流、扫描周期、CORS 与端口。
type Config struct {
	Port            string `env:"PORT" envDefault:"8080"`
	DBHost          string `env:"DB_HOST" envDefault:"localhost"`
	DBPort          string `env:"DB_PORT" envDefault:"5432"`
	DBName          string `env:"DB_NAME" envDefault:"cyfreshfood_db"`
	DBUser          string `env:"DB_USER" envDefault:"cyfreshfood_user"`
	DBPassword      string `env:"DB_PASSWORD" envDefault:"cyfreshfood_pwd"`
	DBSSLMode       string `env:"DB_SSLMODE" envDefault:"disable"`
	RedisHost       string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort       string `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword   string `env:"REDIS_PASSWORD" envDefault:""`
	JWTSecret       string `env:"JWT_SECRET" envDefault:"change_me_to_a_long_random_string"`
	JWTExpireHours  int    `env:"JWT_EXPIRE_HOURS" envDefault:"72"`
	RateLimitPerMin int    `env:"RATE_LIMIT_PER_MIN" envDefault:"120"`
	ScanIntervalMin int    `env:"SCAN_INTERVAL_MIN" envDefault:"5"`
	CORSOrigins     string `env:"APP_CORS_ORIGINS" envDefault:"http://localhost:18628"`
}

// Load 使用 caarlos0/env 解析环境变量。
func Load() (Config, error) {
	return env.ParseAs[Config]()
}

// DSN 构造 PostgreSQL 连接串。
func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

// RedisAddr 构造 Redis 地址。
func (c Config) RedisAddr() string { return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort) }

// ScanInterval 返回临期扫描周期。
func (c Config) ScanInterval() time.Duration {
	return time.Duration(c.ScanIntervalMin) * time.Minute
}

// CORSOriginsList 解析逗号分隔的 CORS 来源白名单，生产默认不允许通配符。
func (c Config) CORSOriginsList() []string {
	raw := strings.TrimSpace(c.CORSOrigins)
	if raw == "" || raw == "*" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}
