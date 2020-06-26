// Package router - CHI Mux Routing
package router

import (
	"net/http"

	"github.com/go-chi/chi"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
type (
	// ChiRouter - Router 인터페이스를 적용한 CHI Router 어답터 구조
	ChiRouter struct {
		mux chi.Router
	}
)

// ===== [ Implementations ] =====

// wrapConstructor - 지정한 핸들러에 대한 처리 함수를 Wrapping 처리해서 반환
func (cr *ChiRouter) wrapConstructor(handlers []Constructor) []func(http.Handler) http.Handler {
	var constructors = make([]func(http.Handler) http.Handler, 0)
	for _, h := range handlers {
		constructors = append(constructors, (func(http.Handler) http.Handler)(h))
	}
	return constructors
}

// routesCount - 지정한 Routes 에 대한 등록 Routes 갯수 반환 (Sub Routes 호출 포함)
func (cr *ChiRouter) routesCount(routes chi.Routes) int {
	count := len(routes.Routes())
	for _, route := range routes.Routes() {
		if nil != route.SubRoutes {
			count += cr.routesCount(route.SubRoutes)
		}
	}
	return count
}

// with - 지정한 핸들러들을 인라인으로 구동될 수 있도록 설정 및 반환
func (cr *ChiRouter) with(handlers ...Constructor) chi.Router {
	return cr.mux.With(cr.wrapConstructor(handlers)...)
}

// ServeHTTP - HTTP Request 처리
func (cr *ChiRouter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	cr.mux.ServeHTTP(rw, req)
}

// Handle - Router에 지정한 경로와 메서드를 핸들러에 연결해서 운영될 수 있도록 설정
func (cr *ChiRouter) Handle(method string, path string, handler http.HandlerFunc, handlers ...Constructor) {
	switch method {
	case http.MethodGet:
		cr.GET(path, handler, handlers...)
	case http.MethodPost:
		cr.POST(path, handler, handlers...)
	case http.MethodPut:
		cr.PUT(path, handler, handlers...)
	case http.MethodPatch:
		cr.PATCH(path, handler, handlers...)
	case http.MethodDelete:
		cr.DELETE(path, handler, handlers...)
	case http.MethodHead:
		cr.HEAD(path, handler, handlers...)
	case http.MethodOptions:
		cr.OPTIONS(path, handler, handlers...)
	}
}

// Any  - 지정한 HTTP 모든 메서드로 경로 등록
func (cr *ChiRouter) Any(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Handle(path, handler)
}

// GET - 지정한 HTTP GET 경로 등록
func (cr *ChiRouter) GET(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Get(path, handler)
}

// POST - 지정한 HTTP POST 경로 등록
func (cr *ChiRouter) POST(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Post(path, handler)
}

// PUT - 지정한 HTTP PUT 경로 등록
func (cr *ChiRouter) PUT(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Put(path, handler)
}

// DELETE - 지정한 HTTP DELETE 경로 등록
func (cr *ChiRouter) DELETE(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Delete(path, handler)
}

// PATCH - 지정한 HTTP PATCH 경로 등록
func (cr *ChiRouter) PATCH(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Patch(path, handler)
}

// HEAD - 지정한 HTTP HEAD 경로 등록
func (cr *ChiRouter) HEAD(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Head(path, handler)
}

// OPTIONS - 지정한 HTTP OPTIONS 경로 등록
func (cr *ChiRouter) OPTIONS(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Options(path, handler)
}

// TRACE - 지정한 HTTP TRACE 경로 등록
func (cr *ChiRouter) TRACE(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Trace(path, handler)
}

// CONNECT - 지정한 HTTP CONNECT 경로 등록
func (cr *ChiRouter) CONNECT(path string, handler http.HandlerFunc, handlers ...Constructor) {
	cr.with(handlers...).Connect(path, handler)
}

// Group - 지정한 경로에 해당 Child Router 구성 및 반환
func (cr *ChiRouter) Group(path string) Router {
	return &ChiRouter{cr.mux.Route(path, nil)}
}

// Use - Router에 지정한 미들웨어 핸들러 사용 설정
func (cr *ChiRouter) Use(handlers ...Constructor) Router {
	cr.mux.Use(cr.wrapConstructor(handlers)...)
	return cr
}

// RoutesCount - Router에 등록된 Route 갯수 반환 (Sub Routes 포함)
func (cr *ChiRouter) RoutesCount() int {
	return cr.routesCount(cr.mux)
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewChiRouterWithOptions - 지정한 옵션을 기준으로 CHI Mux Router 인스턴스 생성
func NewChiRouterWithOptions(options Options) *ChiRouter {
	router := chi.NewRouter()
	router.NotFound(options.NotFoundHandler)

	return &ChiRouter{
		mux: router,
	}
}
