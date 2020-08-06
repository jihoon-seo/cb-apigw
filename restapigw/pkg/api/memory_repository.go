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
		Sources []*DefinitionMap
	}
)

// ===== [ Implementations ] =====

// getSource - 지정한 소스 경로에 맞는 SourceMap 반환
func (imr *InMemoryRepository) getSource(path string) *DefinitionMap {
	for _, sm := range imr.Sources {
		if sm.Source == path {
			return sm
		}
	}

	return nil
}

// Add adds an api definition to the repository
func (imr *InMemoryRepository) add(source string, ec *config.EndpointConfig) error {
	imr.Lock()
	defer imr.Unlock()

	log := logging.GetLogger()

	isValid, err := ec.Validate()
	if !isValid || nil != err {
		log.WithError(err).Error("Validation errors")
		return err
	}

	sm := imr.getSource(source)
	if nil != sm {
		sm.Definitions = append(sm.Definitions, ec)
		log.Debug(ec.Name + " definition added to " + source + " source")
	} else {
		sm := &DefinitionMap{Source: source, State: NONE, Definitions: make([]*config.EndpointConfig, 0)}
		sm.Definitions = append(sm.Definitions, ec)
		imr.Sources = append(imr.Sources, sm)
	}
	return nil
}

// Close - 사용 중인 Memory Repository 세션 종료
func (imr *InMemoryRepository) Close() error {
	return nil
}

// FindSources - 리포지토리에서 관리하는 API Source 경로들을 반환
func (imr *InMemoryRepository) FindSources() ([]string, error) {
	imr.RLock()
	defer imr.RUnlock()

	sources := make([]string, 0)
	for _, sm := range imr.Sources {
		sources = append(sources, sm.Source)
	}

	return sources, nil
}

// FindAllBySource - 지정한 Source에서 API Routing 설정 정보 반환
func (imr *InMemoryRepository) FindAllBySource(source string) ([]*config.EndpointConfig, error) {
	imr.RLock()
	defer imr.RUnlock()

	endpoints := make([]*config.EndpointConfig, 0)
	sm := imr.getSource(source)
	if nil != sm {
		for _, v := range sm.Definitions {
			endpoints = append(endpoints, v)
		}
	}
	return endpoints, nil
}

// FindAll - 사용 가능한 모든 API Routing 설정 검증 및 반환
func (imr *InMemoryRepository) FindAll() ([]*DefinitionMap, error) {
	imr.RLock()
	defer imr.RUnlock()

	return imr.Sources, nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewInMemoryRepository creates a in memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{Sources: make([]*DefinitionMap, 0)}
}
