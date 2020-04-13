// Package ratelimit - Token Bucket 기반의 Rate Limit 처리를 지원하는 패키지
package ratelimit

import (
	"errors"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/ratelimit/limiter"
)

// ===== [ Constants and Variables ] =====

var (
	// ErrLimited - Rate Limit 제한이 되었을 때의 오류
	ErrLimited = errors.New("ERROR: rate limit exceeded")
)

// ===== [ Types ] =====

// ILimiter - Rate limit 운영을 위한 인터페이스
type ILimiter interface {
	// Rate limit 초과 여부 검증
	Allow() bool
}

// RateLimiter - Rate limit 운영을 위한 Bucket Wrapper 구조
type RateLimiter struct {
	limiter *limiter.Bucket
}

// ===== [ Implementations ] =====

// Allow - Rate Limit 처리를 위해 Bucket에서 Token 사용이 가능한지를 검증하고, 1개의 Token을 사용한다.
func (rl *RateLimiter) Allow() bool {
	return rl.limiter.TakeAvailable(1) > 0
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewLimiterWithRate - 지정한 비율과 최대 용량을 기준으로 Rate Limiter 생성
func NewLimiterWithRate(maxRate float64, capacity int64) RateLimiter {
	return RateLimiter{
		limiter.NewBucketWithRate(maxRate, capacity),
	}
}

// NewLimiterWithFillInterval - 지정한 기간과 최대 용량을 기준으로 Rate Limiter 생성
func NewLimiterWithFillInterval(fillInterval time.Duration, capacity int64) RateLimiter {
	return RateLimiter{
		limiter.NewBucket(fillInterval, capacity),
	}
}
