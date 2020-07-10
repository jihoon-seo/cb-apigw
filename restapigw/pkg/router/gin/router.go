package gin

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/admin"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
	httpServer "github.com/cloud-barista/cb-apigw/restapigw/pkg/transport/http/server"
	"github.com/gin-gonic/gin"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

type (
	// PipeConfig - API G/W 운영을 위한 Pipeline 구조
	PipeConfig struct {
		Context        context.Context
		Engine         *DynamicEngine
		Repository     api.Repository
		Middlewares    []gin.HandlerFunc
		HandlerFactory HandlerFactory
		ProxyFactory   proxy.Factory
		Logger         logging.Logger

		adminServer        *admin.Server
		currConfigurations *api.Configuration
		configurationChan  chan api.ConfigurationChanged
	}

	// DynamicEngine - 동적 라우팅 구성을 위한 Routing Engine 구조
	DynamicEngine struct {
		engine http.Handler
		mu     sync.Mutex
	}
)

// ===== [ Implementations ] =====

// ServeHTTP -
func (de *DynamicEngine) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	de.engine.ServeHTTP(rw, req)
}

// SetHandler -
func (de *DynamicEngine) SetHandler(h http.Handler) {
	de.mu.Lock()
	defer de.mu.Unlock()

	de.engine = h
}

// close - 동작 중인 Server들을 모두 종료 처리
func (pc PipeConfig) close() {
	// defer s.closeChannel()
	defer pc.adminServer.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		<-ctx.Done()
		if ctx.Err() == context.Canceled {
			logger.Info("terminiating by interrupt signal")
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			panic("[SERVER] Timeout while stopping " + core.AppName + ", killing instance ✝")
		}
	}(ctx)
}

// listenProviders - 지정한 종료 채널을 기준으로 처리 및 변경에 대한 Loop 설정 (stop or configuration change, ...)
func (pc PipeConfig) listenProviders(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-pc.configurationChan:
			if !ok {
				return
			}

			// TODO: 설정 변경 여부 검증
			if pc.currConfigurations.EqualsTo(configMsg.Configurations) {
				logging.GetLogger().Debug("Skipping same configuration")
				continue
			}

			// TODO: 설정 변경 반영
			//s.currConfigurations.Definitions = configMsg.Configurations.Definitions
			//s.handleEvent(configMsg.Configurations)
		}
	}
}

// startAdminServer - Admin Server 설정 및 구동
func (pc PipeConfig) startAdminServer(sConf config.ServiceConfig) error {
	pc.adminServer = admin.New(
		admin.WithLog(&pc.Logger),
		admin.WithConfigurations(pc.currConfigurations),
		admin.WithPort(sConf.Admin.Port),
		admin.WithTLS(sConf.Admin.TLS),
		admin.WithCredentials(sConf.Admin.Credentials),
		//admin.WithProfiler(s.profilingEnabled, s.profilingPublic),
	)

	if err := pc.adminServer.Start(); nil != err {
		return errors.Wrap(err, "[ADMIN SERVER] Could not start Admin API")
	}

	// We're listening to the configuration changes in any case, even if provider does not implement Listener,
	// so we can use "file" storage as memory - all the persistent definitions are loaded on startup,
	// but then API allows to manipulate proxies in memory. Otherwise api calls just stuck because channel is busy.
	go func() {
		ch := make(chan api.ConfigurationMessage)
		listener, providerIsListener := pc.Repository.(api.Listener)
		if providerIsListener {
			listener.Listen(pc.Context, ch)
		}

		for {
			select {
			case c, more := <-pc.adminServer.ConfigurationChan:
				if !more {
					return
				}

				// TODO: Update Endpoints
				//s.updateConfigurations(c)
				//s.handleEvent(s.currConfigurations)

				if providerIsListener {
					ch <- c
				}
			case <-pc.Context.Done():
				close(ch)
				return
			}
		}
	}()

	if watcher, ok := pc.Repository.(api.Watcher); ok {
		watcher.Watch(pc.Context, pc.configurationChan)
	}

	return nil
}

// Run - 서비스 설정을 기준으로 Gin Engine 구성 / Route - Endpoint Handler 추가 / HTTP Server 구동
func (pc PipeConfig) Run(sConf config.ServiceConfig) {
	pc.Logger.Info("[SERVER] Starting.")
	// 설정된 Middleware 들을 사용하도록 설정
	pc.Engine.engine.(*gin.Engine).Use(pc.Middlewares...)

	// 종료 처리
	go func() {
		defer pc.close()

		<-pc.Context.Done()
		reqAcceptGraceTimeout := time.Duration(sConf.GraceTimeout)
		if reqAcceptGraceTimeout > 0 {
			pc.Logger.Infof("[SERVER] Waiting %s for incoming requests to cease", reqAcceptGraceTimeout)
			time.Sleep(reqAcceptGraceTimeout)
		}

		pc.Logger.Info("[SERVER] Stopping server gracefully")
	}()

	// TODO: Endpoints Registry
	// if sConf.Debug {
	// 	// Debug Endpoint Handler 구성
	// 	pc.registerDebugEndpoints()
	// }

	// // Endpoint Handler 구성
	// pc.registerEndpoints(sConf.Endpoints)

	// pc.Engine.NoRoute(func(c *gin.Context) {
	// 	c.Header(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
	// })

	go func() {
		// TODO: 서버 생성 및 실행
		if err := httpServer.RunServer(pc.Context, sConf, pc.Engine, pc.Logger); nil != err {
			pc.Logger.Error(err)
		}
	}()

	// TODO: Routing 구성???

	// Provider 구성 (Admin)
	//go pc.listenProviders(s.stopChan)

	// TODO: Checking Repository Definitions
	defs, err := pc.Repository.FindAll()
	if nil != err {
		pc.Logger.WithError(err).Warning("[API SERVER] Could not find all configurations from the repository")
	}

	// Definition 기준으로 Admin API 구동
	pc.currConfigurations = &api.Configuration{Definitions: defs}
	if err := pc.startAdminServer(sConf); nil != err {
		logger.WithError(err).Fatal("[API SERVER] Could not start Admin API Repository proviers")
	}

	// svr := server.New(
	// 	server.WithServiceConfig(sConf),
	// 	server.WithProvider(repo),
	// )

	// svr.StartWithContext(ctx)
	// defer svr.Close()

	// svr.Wait()
	pc.Logger.Info("[SERVER] Started.")

	<-pc.Context.Done()
	pc.Logger.Info("[SERVER] Terminated")

	// HTTP Server 환경 초기화
	//httpServer.InitHTTPDefaultTransport(sConf)

	// // Gin Engine을 Handler로 사용하는 HTTP Server 구동
	// if err := httpServer.RunServer(pc.Context, sConf, pc.Engine); err != nil {
	// 	pc.Logger.Error(err.Error())
	// }

}

//// registerDebugEndpoints - debug 옵션이 활성화된 경우에 `__debug` 로 호출되는 Enpoint Handler 구성
//func (pc PipeConfig) registerDebugEndpoints() {
// 	handler := DebugHandler(*pc.Logger)
// 	pc.Engine.GET("/__debug/*param", handler)
// 	pc.Engine.POST("/__debug/*param", handler)
// 	pc.Engine.PUT("/__debug/*param", handler)
// }

// // registerEndpoints - 지정한 Endpoint 설정들을 기준으로 Endpoint Handler 구성
// func (pc PipeConfig) registerEndpoints(eConfs []*config.EndpointConfig) {
// 	for _, eConf := range eConfs {
// 		// Endpoint에 연결되어 동작할 수 있도록 ProxyFactory의 Call chain에 대한 인스턴스 생성 (ProxyStack)
// 		proxyStack, err := pc.ProxyFactory.New(eConf)
// 		if err != nil {
// 			pc.Logger.Error("calling the ProxyFactory", err.Error())
// 			continue
// 		}

// 		if eConf.IsBypass {
// 			// Bypass case
// 			pc.registerGroup(eConf.Endpoint, pc.HandlerFactory(eConf, proxyStack), len(eConf.Backend))
// 		} else {
// 			// Normal case
// 			pc.registerEndpoint(eConf.Method, eConf.Endpoint, pc.HandlerFactory(eConf, proxyStack), len(eConf.Backend))
// 		}
// 	}
// }

// // registerGroup - Bypass인 경우는 Group 단위로 Gin Engine에 Endpoint Handler 등록
// func (pc PipeConfig) registerGroup(path string, handler gin.HandlerFunc, totBackends int) {
// 	if totBackends > 1 {
// 		pc.Logger.Error("Bypass endpoint must have a single backend! Ignoring", path)
// 		return
// 	}

// 	// Bypass에 적합한 Group 정보 조정 및 Route 등록
// 	suffix := "/" + core.Bypass
// 	group := strings.TrimSuffix(path, suffix)

// 	groupRoute := pc.Engine.Group(group)
// 	groupRoute.Any(suffix, handler)
// }

// // registerEndpoint - 지정한 정보를 기준으로 Gin Engine에 Endpoint Handler 등록
// func (pc PipeConfig) registerEndpoint(method, path string, handler gin.HandlerFunc, totBackends int) {
// 	method = strings.ToTitle(method)
// 	if method != http.MethodGet && totBackends > 1 {
// 		pc.Logger.Error(method, "endpoints must have a single backend! Ignoring", path)
// 		return
// 	}
// 	switch method {
// 	case http.MethodGet:
// 		pc.Engine.GET(path, handler)
// 	case http.MethodPost:
// 		pc.Engine.POST(path, handler)
// 	case http.MethodPut:
// 		pc.Engine.PUT(path, handler)
// 	case http.MethodPatch:
// 		pc.Engine.PATCH(path, handler)
// 	case http.MethodDelete:
// 		pc.Engine.DELETE(path, handler)
// 	default:
// 		pc.Logger.Error("Unsupported method", method)
// 	}
// }

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====
