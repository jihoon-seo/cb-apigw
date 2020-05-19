// Package server -
package server

import (
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Option - REST API G/W Server 운영에 필요한 옵션 서정을 위한 함수 형식
	Option func(*Server)
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// WithSystemConfig - System Configuration 적용
func WithSystemConfig(sConf *config.ServiceConfig) Option {
	return func(s *Server) {
		s.serviceConf = sConf
	}
}

// WithProvider - 지정한 Configuration Provider 적용
func WithProvider(provider api.Repository) Option {
	return func(s *Server) {
		s.provider = provider
	}
}
