// Package server -
package server

import (
	"context"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	ginMetrics "github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/metrics/gin"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
	ginRouter "github.com/cloud-barista/cb-apigw/restapigw/pkg/router/gin"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====

// setupGinRouter - Gin 기반으로 동작하는 Router 설정
func setupGinRouter(ctx context.Context, sConf *config.ServiceConfig, logger logging.Logger) router.Router {
	mc := ginMetrics.New(ctx, sConf.Middleware, logger, sConf.Debug)

	// API G/W Server 구동
	return ginRouter.New(sConf,
		ginRouter.WithLogger(logger),
		ginRouter.WithHandlerFactory(setupGinHandlerFactory(logger, mc)),
		ginRouter.WithProxyFactory(setupGinProxyFactory(logger, setupGinBackendFactoryWithContext(ctx, logger, mc), mc)),
	)
}

// ===== [ Public Functions ] =====

// SetupRouter - API G/W 운영을 위한 Router 설정
func SetupRouter(ctx context.Context, sConf *config.ServiceConfig, logger logging.Logger) router.Router {
	switch sConf.RouterEngine {
	default:
		return setupGinRouter(ctx, sConf, logger)
	}
}
