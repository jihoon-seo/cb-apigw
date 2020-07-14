// Package cmd -
package cmd

import (
	"context"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/auth"
	cors "github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/cors/gin"
	httpcache "github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/httpcache"
	httpsecure "github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/httpsecure/gin"
	ginMetrics "github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/metrics/gin"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/metrics/influxdb"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/opencensus"
	ginOpencensus "github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/opencensus/router/gin"
	ratelimitProxy "github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/ratelimit/proxy"
	ratelimitRouter "github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/ratelimit/router/gin"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
	ginRouter "github.com/cloud-barista/cb-apigw/restapigw/pkg/router/gin"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/transport/http/client"
	"github.com/gin-gonic/gin"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====

// newEngine - HTTP Server 및 Admin Server 운영을 위한 Router Engine 구성
func newEngine(sConf config.ServiceConfig, log logging.Logger) *ginRouter.DynamicEngine {
	if !sConf.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true

	// CORS Middleware 반영
	cors.New(sConf.Middleware, engine)

	// HTTPSecure Middleware 반영
	if err := httpsecure.Register(sConf.Middleware, engine); err != nil {
		log.Warning(err)
	}

	de := &ginRouter.DynamicEngine{}
	de.SetHandler(engine)

	return de
}

// newHandlerFactory - Middleware들과 Opencensus 처리를 반영한 Gin Endpoint Handler 생성
func newHandlerFactory(log logging.Logger, mc *ginMetrics.Collector) ginRouter.HandlerFactory {
	// Rate Limit 처리용 RouterHandlerFactory 구성
	handlerFactory := ratelimitRouter.HandlerFactory(ginRouter.EndpointHandler, log)
	// TODO: JWT Auth, JWT Rejector

	// 임시로 HMAC을 활용한 Auth 인증 처리용 RouteHandlerFactory 구성
	handlerFactory = auth.HandlerFactory(handlerFactory, log)

	// metrics Collector를 활용하는 RouteHandlerFactory 구성
	handlerFactory = mc.HandlerFactory(handlerFactory, log)

	// Opencensus를 활용하는 RouteHandlerFactory 구성
	handlerFactory = ginOpencensus.HandlerFactory(handlerFactory, log)

	return handlerFactory
}

// newProxyFactory - 지정된 BackendFactory를 기반으로 동작하는 ProxyFactory 생성
func newProxyFactory(logger logging.Logger, backendFactory proxy.BackendFactory, mc *ginMetrics.Collector) proxy.Factory {
	proxyFactory := proxy.NewDefaultFactory(backendFactory, logger)

	// Metrics 연동 기반의 ProxyFactory 설정
	proxyFactory = mc.ProxyFactory("pipe", proxyFactory)

	// Opencensus 연동 기반의 ProxyFactory 설정
	proxyFactory = opencensus.ProxyFactory(proxyFactory)

	return proxyFactory
}

// newBackendFactoryWithContext - 지정된 Context 기반에서 활용가능한 middleware들이 적용된 BackendFactory 생성
func newBackendFactoryWithContext(ctx context.Context, logger logging.Logger, mc *ginMetrics.Collector) proxy.BackendFactory {
	requestExecutorFactory := func(bConf *config.BackendConfig) client.HTTPRequestExecutor {
		var clientFactory client.HTTPClientFactory

		// TODO: Backend Auth

		// HTTPCache 가 적용된 HTTP Client
		clientFactory = httpcache.NewHTTPClient(bConf)

		// Opencensus 와 연계된 HTTP Request Executor
		return opencensus.HTTPRequestExecutor(clientFactory)
	}

	// Opencensus HTTPRequestExecutor를 사용하는 Base BackendFactory 설정
	backendFactory := func(bConf *config.BackendConfig) proxy.Proxy {
		return proxy.NewHTTPProxyWithHTTPExecutor(bConf, requestExecutorFactory(bConf), bConf.Decoder)
	}

	// TODO: Martian for Backend

	// Backend 호출에 대한 Rate Limit Middleware 설정
	backendFactory = ratelimitProxy.BackendFactory(backendFactory)

	// TODO: Circuit-Breaker for Backend

	// Metrics 연동 기반의 BackendFactory 설정
	backendFactory = mc.BackendFactory("backend", backendFactory)

	// Opencensus 연동 기반의 BackendFactory 설정
	backendFactory = opencensus.BackendFactory(backendFactory)

	return backendFactory
}

// ===== [ Public Functions ] =====

// SetupAndRunGinRouter - GIN 기반으로 동작하는 API G/W 및 Admin Server 설정 및 구동
func SetupAndRunGinRouter(ctx context.Context, sConf config.ServiceConfig, repo api.Repository, log logging.Logger) {
	// Metrics 처리릉 위한 Collector 구성
	mc := ginMetrics.New(ctx, sConf.Middleware, log, sConf.Debug)

	// 설정에 따른 InfluxDB 설정
	if nil != mc.Config {
		// Metrics 처리를 위한 InfluxDB 클라이언트 설정
		influxdb.SetupAndRun(ctx, mc.Config.InfluxDB, func() interface{} { return mc.Snapshot() }, &log)
	} else {
		log.Warn("Skip the influxdb setup and running because the no metrics configuration or incorect")
	}

	// Opencensus 설정
	if err := opencensus.Setup(ctx, sConf); nil != err {
		log.Fatal(err)
	}

	ctx = core.ContextWithSignal(ctx)

	// PipeConfig 구성 (Router (Enpoint Handler) - Proxy - Backend 연계 파이프 설정)
	pipeConfig := ginRouter.PipeConfig{
		Context:        ctx,
		Engine:         newEngine(sConf, log),
		Repository:     repo,
		Middlewares:    []gin.HandlerFunc{},
		HandlerFactory: newHandlerFactory(log, mc),
		ProxyFactory:   newProxyFactory(log, newBackendFactoryWithContext(ctx, log, mc), mc),
		Logger:         log,
	}

	// PipeConfig 기반으로 운영되는 Server 구동
	pipeConfig.Run(sConf)
	// defer pipeConfig.Close()

	pipeConfig.Wait()
	pipeConfig.Logger.Info("[SERVER] Terminated.")
}
