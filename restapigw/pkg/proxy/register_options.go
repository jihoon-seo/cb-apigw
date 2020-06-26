package proxy

import (
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
type (
	// RegisterOption - Register Option 구성 함수
	RegisterOption func(*Register)
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// WithRouter - 지정한 Router 설정
func WithRouter(router router.Router) RegisterOption {
	return func(r *Register) {
		r.router = router
	}
}

// WithFlushInterval - 연결을 비워(출력)하는 기간 설정
func WithFlushInterval(d time.Duration) RegisterOption {
	return func(r *Register) {
		r.flushInterval = d
	}
}

// WithIdleConnectionsPerHost - 각 호스트당 유지되는 유휴 연결 갯수 설정
func WithIdleConnectionsPerHost(count int) RegisterOption {
	return func(r *Register) {
		r.idleConnectionsPerHost = count
	}
}

// WithIdleConnectionTimeout - 유휴 연결이 유지되는 최대 시간 설정
func WithIdleConnectionTimeout(d time.Duration) RegisterOption {
	return func(r *Register) {
		r.idleConnectionTimeout = d
	}
}

// WithIdleConnectionPurgeTicker - 유휴 연결을 정리하는 기간 설정
func WithIdleConnectionPurgeTicker(d time.Duration) RegisterOption {
	var ticker *time.Ticker

	if d != 0 {
		ticker = time.NewTicker(d)
	}

	return func(t *Register) {
		t.idleConnectionPurgeTicker = ticker
	}
}
