// Package api -
package api

import (
	"sync"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
type (
	// InMemoryRepository - Memory 기반의 Repository 관리 형식
	InMemoryRepository struct {
		sync.RWMutex
		endpoints map[string]*config.EndpointConfig
	}
)

// ===== [ Implementations ] =====

// Add adds an api definition to the repository
func (imr *InMemoryRepository) add(ec *config.EndpointConfig) error {
	imr.Lock()
	defer imr.Unlock()

	err := ec.Validate()
	if err != nil {
		logging.GetLogger().WithError(err).Error("Validation errors")
		return err
	}

	imr.endpoints[ec.Endpoint] = ec

	return nil
}

// Close - 사용 중인 Memory Repository 세션 종료
func (imr *InMemoryRepository) Close() error {
	return nil
}

// FindAll - 사용 가능한 모든 API Routing 설정 검증 및 반환
func (imr *InMemoryRepository) FindAll() ([]*config.EndpointConfig, error) {
	imr.RLock()
	defer imr.RUnlock()

	var endpoints []*config.EndpointConfig
	for _, endpoint := range imr.endpoints {
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewInMemoryRepository creates a in memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{endpoints: make(map[string]*config.EndpointConfig)}
}
