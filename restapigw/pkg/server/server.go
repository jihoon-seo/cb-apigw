// Package server - REST API G/w Server 기능 제공 패키지
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/admin"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
	"github.com/go-chi/chi"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Server - REST API G/W 정보 형식
	Server struct {
		server      *http.Server
		adminServer *admin.Server
		provider    api.Repository

		serviceConf        *config.ServiceConfig
		currConfigurations *api.Configuration

		configurationChan chan api.ConfigurationChanged
		stopChan          chan struct{}
	}
)

// ===== [ Implementations ] =====

// isConfigurationChanClosed - Configuration Changed Channel 종료 여부 검증
func (s *Server) isConfigurationChanClosed(ch <-chan api.ConfigurationChanged) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

// isStopChanClosed - Service Stop Channel 종료 여부 검증
func (s *Server) isStopChanClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

// closeChannel - 설정 변경과 종료에 대한 채널 종료
func (s *Server) closeChannel() {
	if !s.isConfigurationChanClosed(s.configurationChan) {
		close(s.configurationChan)
	}
	if !s.isStopChanClosed(s.stopChan) {
		close(s.stopChan)
	}
}

// updateConfiguration - 지정한 Configuration 변경 메시지 정보를 기준으로 Configuration 정보 갱신 처리
func (s *Server) updateConfigurations(cfg api.ConfigurationMessage) {
	currentDefinitions := s.currConfigurations.Definitions

	// TODO: Update Operation
	// switch cfg.Operation {
	// case api.AddedOperation:
	// 	currentDefinitions = append(currentDefinitions, cfg.Configuration)
	// case api.UpdatedOperation:
	// 	for i, d := range currentDefinitions {
	// 		if d.Name == cfg.Configuration.Name {
	// 			currentDefinitions[i] = cfg.Configuration
	// 		}
	// 	}
	// case api.RemovedOperation:
	// 	for i, d := range currentDefinitions {
	// 		if d.Name == cfg.Configuration.Name {
	// 			copy(currentDefinitions[i:], currentDefinitions[i+1:])
	// 			// currentDefinitions[len(currentDefinitions)-1] = nil // or the zero value of T
	// 			currentDefinitions = currentDefinitions[:len(currentDefinitions)-1]
	// 		}
	// 	}
	// }

	s.currConfigurations.Definitions = currentDefinitions
}

// handleEvent - 지정된 설정(변경된)을 라우트 정보 재 구성
func (s *Server) handleEvent(conf *api.Configuration) {
	logger := logging.GetLogger()
	logger.Debug("Refreshing configuration")
	// TODO: New Routing 적용
	// newRouter := s.createRouter()

	// s.register.UpdateRouter(newRouter)
	// s.apiLoader.RegisterAPIs(conf.Definitions)

	// plugin.EmitEvent(plugin.ReloadEvent, plugin.OnReload{Configurations: conf.Definitions})

	// s.server.Handler = newRouter
	logger.Debug("Configuration refresh done")
}

// listenProviders - 지정한 종료 채널을 기준으로 처리 및 변경에 대한 Loop 설정 (stop or configuration change, ...)
func (s *Server) listenProviders(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-s.configurationChan:
			if !ok {
				return
			}

			// 설정 변경 여부 검증
			if s.currConfigurations.EqualsTo(configMsg.Configurations) {
				logging.GetLogger().Debug("[API Configuration changed] - Skipping same configuration")
				continue
			}

			// 설정 변경 반영
			s.currConfigurations.Definitions = configMsg.Configurations.Definitions
			s.handleEvent(configMsg.Configurations)
		}
	}
}

// listenAndServe - 지정한 핸들러 기준으로 동작하는 HTTP Server들 구동
func (s *Server) listenAndServe(handler http.Handler) error {
	address := fmt.Sprintf(":%v", s.serviceConf.Port)
	logger := logging.GetLogger().WithField("address", address)

	s.server = &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  s.serviceConf.ReadTimeout,
		WriteTimeout: s.serviceConf.WriteTimeout,
		IdleTimeout:  s.serviceConf.IdleTimeout,
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return errors.Wrap(err, "error opening listener")
	}

	if s.serviceConf.TLS.IsHTTPS() {
		s.server.Addr = fmt.Sprintf(":%v", s.serviceConf.TLS.Port)
		if s.serviceConf.TLS.Redirect {
			go func() {
				logger.Info("Listening HTTP redirects to HTTPs")
				logging.GetLogger().Fatal(http.Serve(listener, admin.RedirectHTTPS(s.serviceConf.TLS.Port)))
			}()
		}

		logger.Info("Listening HTTPS")
		return s.server.ServeTLS(listener, s.serviceConf.TLS.PublicKey, s.serviceConf.TLS.PrivateKey)
	}

	logger.Info("Certificate and certificate key were not found, defaulting to HTTP")
	return s.server.Serve(listener)
}

// startHTTPServers - REST API G/W 처리를 위한 HTTP Server 시작
func (s *Server) startHTTPServers(ctx context.Context, r router.Router) error {
	// TODO: Router Engine to plugin
	return s.listenAndServe(chi.ServerBaseContext(ctx, r))
}

func (s *Server) startAdminServer(ctx context.Context) error {
	s.adminServer = admin.New(
		admin.WithConfigurations(s.currConfigurations),
		admin.WithPort(s.serviceConf.Admin.Port),
		admin.WithTLS(s.serviceConf.Admin.TLS),
		admin.WithCredentials(s.serviceConf.Admin.Credentials),
		//admin.WithProfiler(s.profilingEnabled, s.profilingPublic),
	)

	if err := s.adminServer.Start(); err != nil {
		return errors.Wrap(err, "could not start "+core.AppName+" Admin API")
	}

	// We're listening to the configuration changes in any case, even if provider does not implement Listener,
	// so we can use "file" storage as memory - all the persistent definitions are loaded on startup,
	// but then API allows to manipulate proxies in memory. Otherwise api calls just stuck because channel is busy.
	go func() {
		ch := make(chan api.ConfigurationMessage)
		listener, providerIsListener := s.provider.(api.Listener)
		if providerIsListener {
			listener.Listen(ctx, ch)
		}

		for {
			select {
			case c, more := <-s.adminServer.ConfigurationChan:
				if !more {
					return
				}

				s.updateConfigurations(c)
				s.handleEvent(s.currConfigurations)

				if providerIsListener {
					ch <- c
				}
			case <-ctx.Done():
				close(ch)
				return
			}
		}
	}()

	if watcher, ok := s.provider.(api.Watcher); ok {
		// 리파지토리 파일 변경 감시 및 처리
		watcher.Watch(ctx, s.configurationChan)
	}

	return nil
}

// Wait - REST API G/W Server 대기 (종료될 떄까지)
func (s *Server) Wait() {
	<-s.stopChan
}

// Close - REST API G/W Server 종료 (Admin Server 종료 포함)
func (s *Server) Close() error {
	logger := logging.GetLogger()

	defer s.closeChannel()
	defer s.adminServer.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		<-ctx.Done()
		if ctx.Err() == context.Canceled {
			logger.Info("terminiating by interrupt signal")
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			panic("Timeout while stopping " + core.AppName + ", killing instance ✝")
		}
	}(ctx)

	return s.server.Close()
}

// Start - REST API G/W Server 시작 (Background Context 사용)
// func (s *Server) Start() error {
// 	return s.StartWithContext(context.Background())
// }

// StartWithContext - REST API G/W Server 시작 (Cancellable Context 사용)
func (s *Server) StartWithContext(ctx context.Context) error {
	logger := logging.NewLogger() //logging.GetLogger()
	logger.Info(core.AppName + " starting...")

	var closed = false

	go func() {
		defer s.Close()

		<-ctx.Done()
		closed = true

		logger.Info("I have to go...")
		reqAcceptGraceTimeout := time.Duration(s.serviceConf.GraceTimeout)
		if reqAcceptGraceTimeout > 0 {
			logger.Infof("Waiting %s for incoming requests to cease", reqAcceptGraceTimeout)
			time.Sleep(reqAcceptGraceTimeout)
		}

		logger.Info("Stopping server gracefully")
	}()

	// TODO: Router 구성
	// r := s.createRouter()
	// s.register = proxy.NewRegister(
	// 	proxy.WithRouter(r),
	// 	proxy.WithFlushInterval(s.sysConfig.BackendFlushInterval),
	// 	proxy.WithIdleConnectionsPerHost(s.sysConfig.MaxIdelConnsPerHost),
	// 	proxy.WithIdleConnectionTimeout(s.sysConfig.IdleConnTimeout),
	// 	proxy.WithIdleConnectionPurgeTicker(s.sysConfig.ConnPurgeInterval),
	// 	proxy.WithStatsClient(s.statsClient),
	// 	proxy.WithIsPublicEndpoint(s.sysConfig.Tracing.IsPublicEndpoint),
	// )

	// // 경합 현상을 피하기 위해 API Loader를 초기화
	// s.apiLoader = loader.NewAPILoader(s.register)

	// HTTP Server 시작
	go func() {
		logger.Debugf(">>>>> Closed??? =====> %s", closed)
		if !closed {
			if err := s.startHTTPServers(ctx, nil); err != nil {
				logger.WithError(err).Fatal("Could not start http servers")
			}
		}
	}()

	// Provider 구성 (Admin)
	go s.listenProviders(s.stopChan)

	// TODO: CHecking Definitions
	endpoints, err := s.provider.FindAll()
	if nil != err {
		return errors.Wrap(err, "could not find all configurations from the provider")
	}

	// Definition 기준으로 Admin API 구동
	s.currConfigurations = &api.Configuration{Definitions: endpoints}
	if err := s.startAdminServer(ctx); nil != err {
		logger.WithError(err).Fatal("Could not start Admin API Providers")
	}

	// TODO: Plugin 처리
	// event := plugin.OnStartup{
	// 	StatsClient:   s.statsClient,
	// 	Register:      s.register,
	// 	Config:        s.sysConfig,
	// 	Configuration: definitions,
	// }

	// if mgoRepo, ok := s.provider.(*api.MongoRepository); ok {
	// 	event.MongoSession = mgoRepo.Session
	// }

	// plugin.EmitEvent(plugin.StartupEvent, event)
	// s.apiLoader.RegisterAPIs(definitions)

	logger.Info(core.AppName + " started.")

	return nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// New - REST API G/W 운영을 위한 Server 인스턴스 생성
func New(opts ...Option) *Server {
	s := Server{
		configurationChan: make(chan api.ConfigurationChanged, 100),
		stopChan:          make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(&s)
	}

	return &s
}
