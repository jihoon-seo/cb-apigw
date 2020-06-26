package balancer

import (
	"errors"
	"math/rand"
)

// ===== [ Constants and Variables ] =====

var (
	// ErrZeroWeight - Backend에 0 가중치가 설정된 경우 오류
	ErrZeroWeight = errors.New("invalid backend, weight 0 given")
)

// ===== [ Types ] =====

type (
	// WeightBalancer - 가중치를 적용하는 로드 밸런서 정보 구조
	WeightBalancer struct{}
)

// ===== [ Implementations ] =====

// Elect - 지정된 Backend Host들을 정보에 따라서 가중치를 적용하여 호출 대상 Backend 선정
func (wb *WeightBalancer) Elect(hosts []*Target) (*Target, error) {
	hostsCount := len(hosts)

	if hostsCount == 0 {
		return nil, ErrEmptyBackendList
	}

	if hostsCount == 1 {
		return hosts[0], nil
	}

	totalWeight := 0
	for _, host := range hosts {
		totalWeight += host.Weight
	}

	if totalWeight <= 0 {
		return nil, ErrZeroWeight
	}

	r := rand.Intn(totalWeight)
	pos := 0

	for _, host := range hosts {
		pos += host.Weight
		if r >= pos {
			continue
		}
		return host, nil
	}

	return nil, ErrCannotElectBackend
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewWeightBalancer - 지정된 Backend Host 정보에 지정된 가중치를 기준으로 로드 밸런서 인스턴 생성
func NewWeightBalancer() *WeightBalancer {
	return &WeightBalancer{}
}
