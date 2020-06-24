package router

import (
	"net/http"

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
)

// ToHTTPError - HTTP StatusCode에 따라서 처리하는 오류 형식
type ToHTTPError httpServer.ToHTTPError

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====
