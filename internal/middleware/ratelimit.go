package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"oracle-etl/internal/config"
)

// rateLimiter는 토큰 버킷 기반 Rate Limiter입니다
type rateLimiter struct {
	mu                sync.RWMutex
	buckets           map[string]*tokenBucket
	requestsPerMinute int
	burstSize         int
}

// tokenBucket은 토큰 버킷을 나타냅니다
type tokenBucket struct {
	tokens     float64
	lastUpdate time.Time
}

// newRateLimiter는 새로운 Rate Limiter를 생성합니다
func newRateLimiter(requestsPerMinute, burstSize int) *rateLimiter {
	rl := &rateLimiter{
		buckets:           make(map[string]*tokenBucket),
		requestsPerMinute: requestsPerMinute,
		burstSize:         burstSize,
	}

	// 주기적으로 오래된 버킷 정리
	go rl.cleanup()

	return rl
}

// cleanup은 오래된 버킷을 주기적으로 정리합니다
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			if now.Sub(bucket.lastUpdate) > 10*time.Minute {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// allow는 요청이 허용되는지 확인하고 토큰을 소비합니다
func (rl *rateLimiter) allow(key string) (allowed bool, remaining int, resetIn time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.buckets[key]

	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(rl.burstSize),
			lastUpdate: now,
		}
		rl.buckets[key] = bucket
	}

	// 시간 경과에 따른 토큰 리필
	elapsed := now.Sub(bucket.lastUpdate)
	tokensPerSecond := float64(rl.requestsPerMinute) / 60.0
	bucket.tokens += elapsed.Seconds() * tokensPerSecond

	// 최대 버스트 크기로 제한
	if bucket.tokens > float64(rl.burstSize) {
		bucket.tokens = float64(rl.burstSize)
	}

	bucket.lastUpdate = now

	// 토큰 소비 시도
	if bucket.tokens >= 1 {
		bucket.tokens--
		remaining = int(bucket.tokens)
		return true, remaining, 0
	}

	// 토큰 부족 - 다음 토큰까지 대기 시간 계산
	tokensNeeded := 1 - bucket.tokens
	resetIn = time.Duration(tokensNeeded/tokensPerSecond*1000) * time.Millisecond

	return false, 0, resetIn
}

// NewRateLimitMiddleware는 Rate Limiting 미들웨어를 생성합니다
func NewRateLimitMiddleware(cfg *config.RateLimitConfig) fiber.Handler {
	// 비활성화된 경우 패스스루
	if !cfg.Enabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	limiter := newRateLimiter(cfg.RequestsPerMinute, cfg.BurstSize)

	return func(c *fiber.Ctx) error {
		// 요청 식별자 생성 (API Key 또는 IP)
		key := getRequestKey(c)

		allowed, remaining, resetIn := limiter.allow(key)

		// Rate Limit 헤더 설정
		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerMinute))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			// 제한 초과
			retryAfterSeconds := int(resetIn.Seconds())
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}
			c.Set("Retry-After", strconv.Itoa(retryAfterSeconds))
			c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(resetIn).Unix(), 10))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"code":        "RATE_LIMIT_EXCEEDED",
				"message":     "요청 제한을 초과했습니다. 잠시 후 다시 시도하세요.",
				"retry_after": retryAfterSeconds,
			})
		}

		return c.Next()
	}
}

// getRequestKey는 요청 식별자를 생성합니다
func getRequestKey(c *fiber.Ctx) string {
	// API Key가 있으면 우선 사용
	apiKey := c.Get("X-API-Key")
	if apiKey != "" {
		return "apikey:" + apiKey
	}

	// X-Forwarded-For 헤더 확인 (프록시 환경)
	forwardedFor := c.Get("X-Forwarded-For")
	if forwardedFor != "" {
		return "ip:" + forwardedFor
	}

	// 직접 연결 IP
	return "ip:" + c.IP()
}
