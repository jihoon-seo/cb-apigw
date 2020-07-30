// Package server - Router Engine 설정 및 HTTP Server 운영을 지원하는 패키지
package server

import (
	"context"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/admin"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
	httpServer "github.com/cloud-barista/cb-apigw/restapigw/pkg/transport/http/server"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
type (
	// Option - Server 인스턴스에 옵션을 설정하는 함수 형식
	Option func(*Server)

	// Server - API G/W 운영을 위한 서버 구조
	Server struct {
		serviceConfig      config.ServiceConfig
		logger             logging.Logger
		provider           api.Repository
		currConfigurations *api.Configuration
		adminServer        *admin.Server
		router             router.Router

		configChan chan api.ConfigurationChanged
		stopChan   chan struct{}
	}
)

// ===== [ Implementations ] =====

// isConfigChanClosed - Configuration Changed Channel 종료 여부 검증
func (s *Server) isConfigChanClosed(ch <-chan api.ConfigurationChanged) bool {
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
	if !s.isConfigChanClosed(s.configChan) {
		close(s.configChan)
	}
	if !s.isStopChanClosed(s.stopChan) {
		close(s.stopChan)
	}
}

// applyChanges - 수신된 API 변경사항 반영
func (s *Server) applyChanges(cfg *api.Configuration) {
	s.logger.Debug("[SERVER] Refreshing configuration")

	// 신규 라우팅 엔진 생성
	s.router.UpdateEngine(s.serviceConfig)
	// 변경된 Routing 규칙 적용
	s.router.RegisterAPIs(&s.serviceConfig, cfg.Definitions)

	s.logger.Debug("[SERVER] Configuration refreshing complete")
}

// updateConfiguration - 지정된 설정변경 메시지를 관리하는 설정에 반영
func (s *Server) updateConfiguration(cfg api.ConfigurationMessage) {
	currDefinitions := s.currConfigurations.Definitions

	switch cfg.Operation {
	case api.AddedOperation:
		currDefinitions = append(currDefinitions, cfg.Configuration)
	case api.UpdatedOperation:
		for i, d := range currDefinitions {
			if d.Name == cfg.Configuration.Name {
				currDefinitions[i] = cfg.Configuration
			}
		}
	case api.RemovedOperation:
		for i, d := range currDefinitions {
			if d.Name == cfg.Configuration.Name {
				copy(currDefinitions[1:], currDefinitions[i+1:])
				currDefinitions = currDefinitions[:len(currDefinitions)-1]
			}
		}
	}

	s.currConfigurations.Definitions = currDefinitions
}

// listenProviders - Server가 종료되는 시점까지 Repository 변경사항 처리
func (s *Server) listenProviders(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-s.configChan:
			if !ok {
				return
			}

			if s.currConfigurations.EqualsTo(configMsg.Configurations) {
				s.logger.Debug("[SERVER] Changed confiruations is same with current configurations. Skip changes")
				continue
			}

			s.logger.Debug("[SERVER] Configuration change detected by repository")
			s.currConfigurations.Definitions = configMsg.Configurations.Definitions
			s.applyChanges(configMsg.Configurations)
		}
	}
}

// createRouter - API G/W 운영을 위한 Router를 생성한다.
func (s *Server) createRouter(ctx context.Context) router.Router {
	return SetupRouter(ctx, s.serviceConfig, s.logger)
}

// startProvider - 지정한 Context 기반으로 Admin Server 구동 및 변경에 대한 처리
func (s *Server) startProvider(ctx context.Context) error {
	s.adminServer = admin.New(
		admin.WithConfigurations(s.currConfigurations),
		admin.WithPort(s.serviceConfig.Admin.Port),
		admin.WithTLS(s.serviceConfig.Admin.TLS),
		admin.WithCredentials(s.serviceConfig.Admin.Credentials),
		admin.WithLog(s.logger),
		admin.WithProfiler(s.serviceConfig.Admin.ProfilingEnabled, s.serviceConfig.Admin.ProfilingPublic),
	)

	if err := s.adminServer.Start(); nil != err {
		return errors.Wrap(err, "[SERVER] Coluld not start Admin API Server")
	}

	// API 변경 대기
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

				// 변경된 설정 갱신
				s.logger.Debug("[SERVER] Configuration change detected by Admin API")
				s.updateConfiguration(c)
				s.applyChanges(s.currConfigurations)

				if providerIsListener {
					ch <- c
				}
			case <-ctx.Done():
				close(ch)
				return
			}
		}
	}()

	// 파일 변경등의 검사
	if watcher, ok := s.provider.(api.Watcher); ok {
		watcher.Watch(ctx, s.configChan)
	}

	return nil
}

// StartWithContext - 지정한 Context 기반으로 API G/W Server 구동 (Done 발생시 종료)
func (s *Server) StartWithContext(ctx context.Context) error {
	// 종료 처리
	go func() {
		defer s.Close()

		<-ctx.Done()
		reqAcceptGraceTimeout := time.Duration(s.serviceConfig.GraceTimeout)
		if 0 < reqAcceptGraceTimeout {
			s.logger.Infof("[SERVER] Waiting %s for incoming requests to cease", reqAcceptGraceTimeout)
			time.Sleep(reqAcceptGraceTimeout)
		}

		s.logger.Info("[SERVER] Stopping server gracefully")
	}()

	// TODO: Router 구성
	s.router = s.createRouter(ctx)

	// HTTP Server 구동
	go func() {
		httpServer.InitHTTPDefaultTransport(s.serviceConfig)

		if err := httpServer.RunServer(ctx, s.serviceConfig, s.router.Engine()); nil != err {
			s.logger.Error(err.Error())
		}
	}()

	// Listen Admin API Providers
	go s.listenProviders(s.stopChan)

	// API 설정 정보 검색
	defs, err := s.provider.FindAll()
	if nil != err {
		return errors.Wrap(err, "could not find all configurations from the repository")
	}

	// Admin Server 구동
	s.currConfigurations = &api.Configuration{Definitions: defs}
	if err := s.startProvider(ctx); nil != err {
		s.logger.WithError(err).Fatal("Could not start api providers")
	}

	// API Definition에 대한 Router 연계 처리
	s.router.RegisterAPIs(&s.serviceConfig, s.currConfigurations.Definitions)

	s.logger.Info("[SERVER] Started")
	return nil
}

// Close - 지정한 Context 기반으로 API G/W Server 구동 (Done 발생시 종료)
func (s *Server) Close() error {
	defer s.closeChannel()
	defer s.adminServer.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		<-ctx.Done()
		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			panic("[SERVER] Timeout while stopping " + core.AppName + ", killing instance")
		}
	}(ctx)

	//return s.httpServer.Close()
	return nil
}

// Wait - Server shutdown 상태까지 대기 (wait stop signal)
func (s *Server) Wait() {
	<-s.stopChan
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// WithServiceConfig - Service Configuration 설정
func WithServiceConfig(sConf config.ServiceConfig) Option {
	return func(s *Server) {
		s.serviceConfig = sConf
	}
}

// WithLogger - Logger 인스턴스 설정
func WithLogger(logger logging.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithRepository - API Repository 인스턴스 설정
func WithRepository(repo api.Repository) Option {
	return func(s *Server) {
		s.provider = repo
	}
}

// New - API G/W 운영을 위한 Server 인스턴스 생성
func New(opts ...Option) *Server {
	s := Server{
		configChan: make(chan api.ConfigurationChanged, 100),
		stopChan:   make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(&s)
	}

	return &s
}
