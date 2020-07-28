// Package admin - Admin API 기능 지원 패키지
package admin

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"

	ginAdapter "github.com/cloud-barista/cb-apigw/restapigw/pkg/core/adapters/gin"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/jwt"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	ginCors "github.com/rs/cors/wrapper/gin"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Server - Admin API Server 관리 정보 형식
	Server struct {
		ConfigurationChan chan api.ConfigurationMessage

		apiHandler *APIHandler
		logger     logging.Logger

		Port             int `mapstructure:"port"`
		Credentials      config.CredentialsConfig
		TLS              config.TLSConfig
		profilingEnabled bool
		profilingPublic  bool
	}
)

// ===== [ Implementations ] =====

// addInternalPublicRoutes - ADMIN Server의 외부에 노출되는 Routes 설정
func (s *Server) addInternalPublicRoutes(ge *gin.Engine) {
	ge.GET("/", gin.WrapH(Home()))
	ge.GET("/status", gin.WrapH(NewOverviewHandler(s.apiHandler.Configs, s.logger)))
	ge.GET("/status/{name}", gin.WrapH(NewStatusHandler(s.apiHandler.Configs, s.logger)))
	// TODO: Metrics 정보 설정은? Prometheus???
}

// addInternalAUthRoutes - ADMIN Server의 Auth 처리용 Routes 설정
func (s *Server) addInternalAuthRoutes(ge *gin.Engine, guard jwt.Guard) {
	handlers := jwt.Handler{Guard: guard}
	ge.POST("/login", gin.WrapH(handlers.Login(s.Credentials, s.logger)))
	authGroup := ge.Group("/auth")
	{
		authGroup.GET("/refresh_token", gin.WrapH(handlers.Refresh()))
	}
}

// addInternalRoutes - API Definition 처리를 위한 Routes 설정
func (s *Server) addInternalRoutes(ge *gin.Engine, guard jwt.Guard) {
	s.logger.Debug("[ADMIN Server] Loading API Endpoints")

	// APIs endpoints
	groupAPI := ge.Group("/apis")
	groupAPI.Use(ginAdapter.Wrap(jwt.NewMiddleware(guard).Handler))
	{
		groupAPI.GET("/", gin.WrapH(s.apiHandler.Get()))
		groupAPI.GET("/{name}", gin.WrapH(s.apiHandler.GetBy()))
		groupAPI.POST("/", gin.WrapH(s.apiHandler.Post()))
		groupAPI.PUT("/{name}", gin.WrapH(s.apiHandler.PutBy()))
		groupAPI.DELETE("/{name}", gin.WrapH(s.apiHandler.DeleteBy()))
	}

	if s.profilingEnabled {
		groupProfiler := ge.Group("/debug/pprof")
		if !s.profilingPublic {
			groupProfiler.Use(ginAdapter.Wrap(jwt.NewMiddleware(guard).Handler))
		}
		{
			groupProfiler.GET("/", gin.WrapF(pprof.Index))
			groupProfiler.GET("/cmdline", gin.WrapF(pprof.Cmdline))
			groupProfiler.GET("/profile", gin.WrapF(pprof.Profile))
			groupProfiler.GET("/symbol", gin.WrapF(pprof.Symbol))
			groupProfiler.GET("/trace", gin.WrapF(pprof.Trace))
		}
	}
}

// isClosedChannel - 이미 채널이 종료되었는지 검증
func (s *Server) isClosedChannel(ch <-chan api.ConfigurationMessage) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

// listenAndServe - GIN Router를 기반으로 HTTP Server 구동
func (s *Server) listenAndServe(h http.Handler) error {
	address := fmt.Sprintf(":%v", s.Port)

	if s.TLS.IsHTTPS() {
		addressTLS := fmt.Sprintf(":%v", s.TLS.Port)
		if s.TLS.Redirect {
			go func() {
				s.logger.WithField("address", address).Info("Listening HTTP redirects to HTTPS")
				s.logger.Fatal(http.ListenAndServe(address, RedirectHTTPS(s.TLS.Port)))
			}()
		}

		s.logger.WithField("address", addressTLS).Info("LIstening HTTPS")
		return http.ListenAndServeTLS(addressTLS, s.TLS.PublicKey, s.TLS.PrivateKey, h)
	}

	s.logger.WithField("address", address).Info("Certificate and certificate key where not found, defaulting to HTTP")
	return http.ListenAndServe(address, h)
}

// Start - Admin API Server 구동
func (s *Server) Start() error {
	s.logger.Info("[API SERVER] Admin API starting...")

	// Gin Router 생성
	engine := gin.Default()

	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "API_NOT_FOUND", "message": "API not found"})
	})

	// Router 설정 및 구동
	go s.listenAndServe(engine)

	s.AddRoutes(engine)

	// TODO: Plugin setup signal

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
		close(s.ConfigurationChan)
	}
	s.logger.Info("[API G/W] Admin API stoped.")
}

// AddRoutes - ADMIN Routes 정보 구성
func (s *Server) AddRoutes(ge *gin.Engine) {
	guard := jwt.NewGuard(s.Credentials)

	// Cors 적용
	ge.Use(
		ginCors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}),
	)

	s.addInternalPublicRoutes(ge)
	s.addInternalAuthRoutes(ge, guard)
	s.addInternalRoutes(ge, guard)
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
