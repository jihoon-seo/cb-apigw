package router

import (
	"net/http"
	"sync"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
	httpServer "github.com/cloud-barista/cb-apigw/restapigw/pkg/transport/http/server"
)

// ===== [ Constants and Variables ] =====

const (
	// HeaderCompleteResponseValue - 정상적으로 종료된 Response들에 대한 Complete Header ("X-CB-RESTAPIGW-COMPLETED")로 설정될 값
	HeaderCompleteResponseValue = httpServer.HeaderCompleteResponseValue
	// HeaderIncompleteResponseValue - 비 정상적으로 종료된 Response들에 대한 Complete Header로 설정될 값
	HeaderIncompleteResponseValue = httpServer.HeaderIncompleteResponseValue
)

var (
	// MessageResponseHeaderName - Response 오류인 경우 클라이언트에 표시할 Header 명
	MessageResponseHeaderName = httpServer.MessageResponseHeaderName
	// CompleteResponseHeaderName - 정상/비정상 종료에 대한 Header 정보를 클라이언트에 알리기 위한 Header 명
	CompleteResponseHeaderName = httpServer.CompleteResponseHeaderName
	// HeadersToSend - Route로 전달된 Request에서 Proxy로 전달할 설정에 지정된 Header들 정보
	HeadersToSend = httpServer.HeadersToSend
	// HeadersToNotSend - Router로 전달된 Request에서 Proxy로 전달죄지 않을 Header들 정보
	HeadersToNotSend = httpServer.HeadersToNotSend
	// UserAgentHeaderValue - Proxy Request에 설정할 User-Agent Header 값
	UserAgentHeaderValue = httpServer.UserAgentHeaderValue

	// DefaultToHTTPError - 항상 InternalServerError로 처리하는 기본 오류
	DefaultToHTTPError = httpServer.DefaultToHTTPError
	// ErrInternalError - 문제가 발생했을 때 InternalServerError로 표현되는 오류
	ErrInternalError = httpServer.ErrInternalError
)

// ===== [ Types ] =====

type (
	// HandlerFactory - 사용할 Router에 Proxy 연계를 위한 함수
	HandlerFactory func(*config.EndpointConfig, proxy.Proxy) http.HandlerFunc

	// ToHTTPError - HTTP StatusCode에 따라서 처리하는 오류 형식
	ToHTTPError httpServer.ToHTTPError

	// Constructor - Middleware 형식의 운영을 위한 함수 형식
	Constructor func(http.Handler) http.Handler

	// Router - Router 운영에 필요한 인터페이스 형식
	Router interface {
		ServeHTTP(rw http.ResponseWriter, req *http.Request)
		Handle(method string, path string, handler http.HandlerFunc, handlers ...Constructor)
		Any(path string, handler http.HandlerFunc, handlers ...Constructor)
		GET(path string, handler http.HandlerFunc, handlers ...Constructor)
		POST(path string, handler http.HandlerFunc, handlers ...Constructor)
		PUT(path string, handler http.HandlerFunc, handlers ...Constructor)
		DELETE(path string, handler http.HandlerFunc, handlers ...Constructor)
		PATCH(path string, handler http.HandlerFunc, handlers ...Constructor)
		HEAD(path string, handler http.HandlerFunc, handlers ...Constructor)
		OPTIONS(path string, handler http.HandlerFunc, handlers ...Constructor)
		TRACE(path string, handler http.HandlerFunc, handlers ...Constructor)
		CONNECT(path string, handler http.HandlerFunc, handlers ...Constructor)
		Group(path string) Router
		Use(handlers ...Constructor) Router

		RoutesCount() int
	}

	// DynamicHandler - description
	DynamicHandler struct {
		h  http.Handler
		mu sync.Mutex
	}
)

// ===== [ Implementations ] =====

// ServeHTTP -
func (dr *DynamicHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	dr.h.ServeHTTP(rw, req)
}

// SetHandler -
func (dr *DynamicHandler) SetHandler(h http.Handler) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	dr.h = h
}

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// // WithServer - Router 연계를 위한 HTTP Server 설정
// func WithServer(server *http.Server) Option {
// 	return func(c *Config) {
// 		c.server = server
// 	}
// }

// // WithLog - Logging에 사용할 Logger 적용
// func WithLog(log *logging.Logger) Option {
// 	return func(c *Config) {
// 		c.logger = log
// 	}
// }

// // WithSystemConfig - System Configuration 적용
// func WithSystemConfig(sConf *config.ServiceConfig) Option {
// 	return func(c *Config) {
// 		c.serviceConf = sConf
// 	}
// }

// // WithProvider - 지정한 Configuration Provider 적용
// func WithProvider(provider api.Repository) Option {
// 	return func(c *Config) {
// 		c.provider = provider
// 	}
// }

// // WithMetricsProducer - 지정한 Metrics Prodcuer 적용
// func WithMetricsProducer(mp *metrics.Metrics) Option {
// 	return func(c *Config) {
// 		c.metricsProducer = mp
// 	}
// }

// // New - Router 구성을 위한 Router Configuration 인스턴스 생성 및 옵션 설정
// func New(opts ...Option) *Config {
// 	c := Config{}

// 	for _, opt := range opts {
// 		opt(&c)
// 	}

// 	return &c
// }
