// Package api - ADMIN API 기능을 제공하는 패키지
package api

import (
	"net/http"
	"reflect"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
)

// ===== [ Constants and Variables ] =====

const (
	// RemovedOperation - 설정 제거 작업
	RemovedOperation ConfigurationOperation = iota
	// UpdatedOperation - 설정 변경 작업
	UpdatedOperation
	// AddedOperation - 설정 등록 작업
	AddedOperation
)

var (
	// ErrAPIDefinitionNotFound is used when the api was not found in the datastore
	ErrAPIDefinitionNotFound = errors.NewWithCode(http.StatusNotFound, "api definition not found")

	// ErrAPINameExists is used when the API name is already registered on the datastore
	ErrAPINameExists = errors.NewWithCode(http.StatusConflict, "api name is already registered")

	// ErrAPIListenPathExists is used when the API listen path is already registered on the datastore
	ErrAPIListenPathExists = errors.NewWithCode(http.StatusConflict, "api listen path is already registered")

	// ErrDBContextNotSet is used when the database request context is not set
	ErrDBContextNotSet = errors.NewWithCode(http.StatusInternalServerError, "DB context was not set for this request")
)

// ===== [ Types ] =====

type (
	// ConfigurationOperation - Configuration 변경에 연계되는 Operation 형식
	ConfigurationOperation int

	// Configuration - API Endpoints
	Configuration struct {
		Definitions []*config.EndpointConfig
	}

	// ConfigurationChanged - Configuration 변경이 발생했을 떄 전송되는 메시지 처리 형식
	ConfigurationChanged struct {
		Configurations *Configuration
	}

	// ConfigurationMessage - Configuration 변경시 사용할 메시지 형식
	ConfigurationMessage struct {
		Operation     ConfigurationOperation
		Configuration *config.EndpointConfig
	}
)

// ===== [ Implementations ] =====

// EqualsTo - 현재 동작 중인 설정과 지정된 설정이 동일한지 여부 검증
func (c *Configuration) EqualsTo(tc *Configuration) bool {
	return reflect.DeepEqual(c, tc)
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewDefinition - 기본 값으로 설정된 API Routing Endpoint 인스턴스 생성
func NewDefinition() *config.EndpointConfig {
	return &config.EndpointConfig{}
}
