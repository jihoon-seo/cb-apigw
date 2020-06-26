package balancer

import "sync"

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// RoundrobinBalancer - 라운드로빈 알고리듬을 적용하는 관리 구조
	RoundrobinBalancer struct {
		current int // 현재 Backend 위치 식별용
		mu      sync.RWMutex
	}
)

// ===== [ Implementations ] =====

// Elect - 지정한 Backend Host 정보들 중에 다음 호출에 사용될 대상 선출
func (rrb *RoundrobinBalancer) Elect(hosts []*Target) (*Target, error) {
	hostsCount := len(hosts)
	if hostsCount == 0 {
		return nil, ErrEmptyBackendList
	}

	if hostsCount == 1 {
		return hosts[0], nil
	}

	if rrb.current >= hostsCount {
		rrb.current = 0
	}

	host := hosts[rrb.current]
	rrb.mu.Lock()
	defer rrb.mu.Unlock()

	rrb.current++

	return host, nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewRoundrobinBalancer - 라우드로빈 방식의 로드 밸런서 인스턴스 생성
func NewRoundrobinBalancer() *RoundrobinBalancer {
	return &RoundrobinBalancer{}
}
