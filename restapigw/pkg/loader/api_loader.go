// Package loader - Proxy 운영을 위한 API 정보 Load 기능 제공 패키지
package loader

import (
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
type (
	// APILoader - Proxy와 연계를 위한 API 정보 Load 구조
	APILoader struct {
		register *proxy.Register
	}
)

// ===== [ Implementations ] =====

// RegisterAPIs - 지정된 Router 정보들 등록
func (al *APILoader) RegisterAPIs(endpoints []*config.EndpointConfig) {
	for _, endpoint := range endpoints {
		al.RegisterAPI(endpoint)
	}
}

// RegisterAPI - 지정된 Router 정보 등록
func (al *APILoader) RegisterAPI(endpoint *config.EndpointConfig) {
	// TODO: Endpoint 등록
	log := logging.GetLogger().WithField("api_endpoint", endpoint.Endpoint)
	log.Debug("Starting RegisterAPI")

	err := endpoint.Validate()
	if nil != err {
		log.WithError(err).Error("Validation errors, Skipping...")
	}

	// TODO: Backend 설정 활성화

	// API 등록
	al.register.Add(endpoint)
	log.Debug("API Registered")
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewAPILoader - API 정보 Loader 인스턴스 생성
func NewAPILoader(register *proxy.Register) *APILoader {
	return &APILoader{register: register}
}
