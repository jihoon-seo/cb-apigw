// Package admin - Admin API 기능 지원 패키지
package admin

import (
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Server - Admin API Server 관리 정보 형식
	Server struct {
		ConfigurationChan chan api.ConfigurationMessage
		apiHandler        *APIHandler

		Port        int `mapstructure:"port"`
		Credentials config.CredentialsConfig
		TLS         config.TLSConfig
	}
)

// ===== [ Implementations ] =====

// isClosedChannel - 이미 채널이 종료되었는지 검증
func (s *Server) isClosedChannel(ch <-chan api.ConfigurationMessage) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

// Start - Admin API Server 구동
func (s *Server) Start() error {
	logging.GetLogger().Info(core.AppName + " Admin API starting...")

	// TODO: API Routing
	// router.DefaultOptions.NotFoundHandler = httpErrors.NotFound
	// r := router.NewChiRouterWithOptions(router.DefaultOptions)
	// go s.listenAndServe(r)

	// s.AddRoutes(r)
	// plugin.EmitEvent(plugin.AdminAPIStartupEvent, plugin.OnAdminAPIStartup{Router: r})

	return nil
}

// Stop - Admin API Server 종료
func (s *Server) Stop() {
	if !s.isClosedChannel(s.ConfigurationChan) {
		close(s.ConfigurationChan)
	}
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// New - Admin Server 구동
func New(opts ...Option) *Server {
	configurationChan := make(chan api.ConfigurationMessage)
	s := Server{
		ConfigurationChan: configurationChan,
		apiHandler:        NewAPIHandler(configurationChan),
	}

	for _, opt := range opts {
		opt(&s)
	}

	return &s
}
