// Package api - CB-Store 기반 Repository
package api

import (
	"io/ioutil"
	"os"

	cbstore "github.com/cloud-barista/cb-store"
	icbs "github.com/cloud-barista/cb-store/interfaces"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// CbStoreRepository - CB-Store 끼반 Repository 관리 정보 형식
	CbStoreRepository struct {
		*InMemoryRepository
		store icbs.Store
	}
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewCbStoreRepository - CB-Store 기반의 Repository 인스턴스 생성
func NewCbStoreRepository(key string) (*CbStoreRepository, error) {
	// logger := logging.GetLogger()
	repo := CbStoreRepository{InMemoryRepository: NewInMemoryRepository(), store: cbstore.GetStore()}

	// Test Data Input
	keyValue, _ := repo.store.Get(key)
	if keyValue.Key == "" || keyValue.Value == "" {
		dir, err := os.Getwd()
		if nil != err {
			return nil, err
		}
		content, err := ioutil.ReadFile(dir + "/conf/apis/cb-restapigw-apis.yaml")
		if nil != err {
			return nil, err
		}
		repo.store.Put(key, string(content))
	}

	// Grab configuration from CB-STORE
	keyValue, _ = repo.store.Get(key)

	// Parse Definitions
	if keyValue.Value != "" {
		// TODO: Parsing Definitions
	}

	return &repo, nil
}
