// Package server -
package server

import (
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
	ginRouter "github.com/cloud-barista/cb-apigw/restapigw/pkg/router/gin"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
type ()

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====

// setupGinRouter - Gin 기반으로 동작하는 Router 설정
func setupGinRouter(sConf config.ServiceConfig, logger logging.Logger) router.Router {
	// API G/W Server 구동
	return ginRouter.New(sConf,
		ginRouter.WithLogger(logger),
	//ginRouter.WithHandlerFactory(hf),
	//ginRouter.withProxyFactory(pf),
	)
}

// ===== [ Public Functions ] =====

// SetupRouter - API G/W 운영을 위한 Router 설정
func SetupRouter(sConf config.ServiceConfig, logger logging.Logger) router.Router {
	switch sConf.RouterEngine {
	default:
		return setupGinRouter(sConf, logger)
	}
}
