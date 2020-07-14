// Package admin - Admin API 기능 지원 패키지
package admin

import (
	"fmt"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Server - Admin API Server 관리 정보 형식
	Server struct {
		ConfigurationChan chan api.ConfigurationMessage

		apiHandler *APIHandler
		logger     *logging.Logger

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
	s.logger.Info("[API SERVER] Admin API starting...")

	// TODO: API Routing
	// router.DefaultOptions.NotFoundHandler = httpErrors.NotFound
	// r := router.NewChiRouterWithOptions(router.DefaultOptions)
	// go s.listenAndServe(r)

	// s.AddRoutes(r)
	// plugin.EmitEvent(plugin.AdminAPIStartupEvent, plugin.OnAdminAPIStartup{Router: r})

	s.logger.Info("[API SERVER] Admin API started.")
	return nil
}

// Stop - Admin API Server 종료
func (s *Server) Stop() {
	if s == nil {
		fmt.Println("server is null")
		return
	}

	if !s.isClosedChannel(s.ConfigurationChan) {
		logging.GetLogger().Info("admin configuration channel closing.")
		close(s.ConfigurationChan)
	}
	s.logger.Info("[API G/W] Admin API stoped.")
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// New - Admin Server 구동
func New(opts ...Option) *Server {
	configurationChan := make(chan api.ConfigurationMessage)
	s := &Server{
		ConfigurationChan: configurationChan,
		apiHandler:        NewAPIHandler(configurationChan),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}
