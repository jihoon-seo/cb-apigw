// Package api - CB-Store 기반 Repository
package api

import (
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	cbstore "github.com/cloud-barista/cb-store"
	icbs "github.com/cloud-barista/cb-store/interfaces"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// CbStoreRepository - CB-Store 기반 Repository 관리 정보 형식
	CbStoreRepository struct {
		*InMemoryRepository
		store    icbs.Store
		storeKey string
	}
)

// ===== [ Implementations ] =====

// getStorePath - API Definition 저장을 위한 Store key path 반환
func (csr *CbStoreRepository) getStorePath(path string) string {
	return csr.storeKey + "/" + path
}

// Write - 변경된 리파지토리 내용을 대상 파일로 출력
func (csr *CbStoreRepository) Write(definitionMaps []*DefinitionMap) error {
	csr.Sources = definitionMaps

	for _, dm := range csr.Sources {
		if dm.State == "REMOVED" {
			err := csr.store.Delete(csr.getStorePath(dm.Source))
			if nil != err {
				return err
			}
		} else if dm.State != "" {
			data, err := sourceDefinitions(dm)
			if nil != err {
				return err
			}

			err = csr.store.Put(csr.getStorePath(dm.Source), string(data))
			if nil != err {
				return err
			}
		}
	}

	return nil
}

// Close - 사용 중인 Repository 세션 종료
func (csr *CbStoreRepository) Close() error {
	// TODO: Watcher Closing
	return nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewCbStoreRepository - CB-Store 기반의 Repository 인스턴스 생성
func NewCbStoreRepository(key string) (*CbStoreRepository, error) {
	logger := logging.GetLogger()
	repo := CbStoreRepository{InMemoryRepository: NewInMemoryRepository(), store: cbstore.GetStore(), storeKey: key}

	// TODO: Watching

	// Grab configuration from CB-STORE
	keyValues, _ := repo.store.GetList(key, true)

	for _, kv := range keyValues {
		// Skip Root
		if kv.Key == key {
			continue
		}

		definition := parseEndpoint([]byte(kv.Value))
		for _, def := range definition.Definitions {
			if err := repo.add(core.GetLastPart(kv.Key, "/"), def); nil != err {
				logger.WithField("endpoint", def.Endpoint).WithError(err).Error("Failed during add endpoint to the repository")
				return nil, err
			}
		}
	}

	return &repo, nil
}
