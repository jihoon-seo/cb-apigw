// Package admin -
package admin

import "github.com/cloud-barista/cb-apigw/restapigw/pkg/api"

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// APIHandler - Admin API 운영을 위한 REST Handler 정보 형식
	APIHandler struct {
		configurationChan chan<- api.ConfigurationMessage

		Configs *api.Configuration
	}
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewAPIHandler - 지정한 Configuration 변경을 설정한 Admin API Handler 인스턴스 생성
func NewAPIHandler(configurationChan chan<- api.ConfigurationMessage) *APIHandler {
	return &APIHandler{
		configurationChan: configurationChan,
	}
}
