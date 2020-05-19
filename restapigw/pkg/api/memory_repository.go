// Package api -
package api

import (
	"sync"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
type (
	// InMemoryRepository - Memory 기반의 Repository 관리 형식
	InMemoryRepository struct {
		sync.RWMutex
		definitions map[string]*Definition
	}
)

// ===== [ Implementations ] =====

// Add adds an api definition to the repository
func (imr *InMemoryRepository) add(d *Definition) error {
	imr.Lock()
	defer imr.Unlock()

	isValid, err := d.Validate()
	if false == isValid && err != nil {
		logging.GetLogger().WithError(err).Error("Validation errors")
		return err
	}

	imr.definitions[d.Name] = d

	return nil
}

// Close - 사용 중인 Memory Repository 세션 종료
func (imr *InMemoryRepository) Close() error {
	return nil
}

// FindAll - 사용 가능한 모든 API Routing 설정 검증 및 반환
func (imr *InMemoryRepository) FindAll() ([]*Definition, error) {
	imr.RLock()
	defer imr.RUnlock()

	var definitions []*Definition
	for _, definition := range imr.definitions {
		definitions = append(definitions, definition)
	}

	return definitions, nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewInMemoryRepository creates a in memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{definitions: make(map[string]*Definition)}
}
