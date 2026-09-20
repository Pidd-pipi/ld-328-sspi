package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/gin-gonic/gin"
)

// bucket 简易令牌桶。
type bucket struct {
	tokens float64
	last   time.Time
}

// RateLimiter 基于用户维度的内存限流（令牌桶）。
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64 // 每秒补充令牌数
	burst   float64 // 桶容量
}

// NewRateLimiter 构造限流器，perMin 为每分钟最大请求数。
func NewRateLimiter(perMin int) *RateLimiter {
	return &RateLimiter{
		buckets: map[string]*bucket{},
		rate:    float64(perMin) / 60.0,
		burst:   float64(perMin),
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.burst, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Limit 返回限流中间件（key 取用户 ID 或客户端 IP）。
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if v, ok := c.Get(UserIDKey); ok {
			key = "u" + toString(v)
		}
		if !rl.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": constants.CodeRateLimited, "message": constants.MsgRateLimited, "data": nil,
			})
			return
		}
		c.Next()
	}
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if i, ok := v.(uint); ok {
		return string(rune(i))
	}
	return ""
}
