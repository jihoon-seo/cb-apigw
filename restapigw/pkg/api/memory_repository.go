// Package api -
package api

import (
	"strings"
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
		Groups []*DefinitionMap
	}
)

// ===== [ Implementations ] =====

// checkDuplicates - 지정한 API Definition이 중복으로 존재하는지 검증
// - Name : 동일 그룹에 중복 불가
// - Endpoint : 전체 그룹에 중복 불가
func (imr *InMemoryRepository) checkDuplicates(group string, eConf *config.EndpointConfig) bool {
	for _, dm := range imr.Groups {
		if dm.CheckDuplicates(group, eConf) {
			return true
		}
	}
	return false
}

// getGroup - 지정한 소스 경로에 맞는 GroupMap 반환
func (imr *InMemoryRepository) getGroup(path string) *DefinitionMap {
	for _, sm := range imr.Groups {
		if strings.EqualFold(sm.Name, path) {
			return sm
		}
	}

	return nil
}

// Add adds an api definition to the repository
func (imr *InMemoryRepository) add(group string, eConf *config.EndpointConfig) error {
	imr.Lock()
	defer imr.Unlock()

	log := logging.GetLogger()

	// API Definition 검증
	err := eConf.Validate()
	if err != nil {
		log.WithError(err).Error("[REPOSITORY] MEMORY > Validation errors")
		return err
	}

	// API Definition Group 생성 및 등록
	sm := imr.getGroup(group)
	if sm == nil {
		sm = &DefinitionMap{Name: group, State: NONE, Definitions: make([]*config.EndpointConfig, 0)}
		imr.Groups = append(imr.Groups, sm)
	}

	// API Definition 중복 검증 (Name and Endpoint)
	if imr.checkDuplicates(group, eConf) {
		log.Warnf("[REPOSITORY] MEMORY > Duplicates exists. [ Name: %s, Endpoint: %s] Ignored.", eConf.Name, eConf.Endpoint)
		return nil
	} else {
		sm.Definitions = append(sm.Definitions, eConf)
	}

	return nil
}

// Close - 사용 중인 Memory Repository 세션 종료
func (imr *InMemoryRepository) Close() error {
	return nil
}

// FindGroups - 리포지토리에서 관리하는 API Group 경로들을 반환
func (imr *InMemoryRepository) FindGroups() ([]string, error) {
	imr.RLock()
	defer imr.RUnlock()

	groups := make([]string, 0)
	for _, sm := range imr.Groups {
		groups = append(groups, sm.Name)
	}

	return groups, nil
}

// FindAllByGroup - 지정한 Group에서 API Routing 설정 정보 반환
func (imr *InMemoryRepository) FindAllByGroup(group string) ([]*config.EndpointConfig, error) {
	imr.RLock()
	defer imr.RUnlock()

	endpoints := make([]*config.EndpointConfig, 0)
	sm := imr.getGroup(group)
	if sm != nil {
		endpoints = append(endpoints, sm.Definitions...)
	}
	return endpoints, nil
}

// FindAll - 사용 가능한 모든 API Routing 설정 검증 및 반환
func (imr *InMemoryRepository) FindAll() ([]*DefinitionMap, error) {
	imr.RLock()
	defer imr.RUnlock()

	return imr.Groups, nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewInMemoryRepository creates a in memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{Groups: make([]*DefinitionMap, 0)}
}
