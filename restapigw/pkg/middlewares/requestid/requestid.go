// Package requestid - RequestID 기능을 지원하는 패키지
package requestid

import (
	"net/http"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/observability"
	"github.com/gofrs/uuid"
)

// ===== [ Constants and Variables ] =====
const (
	reqIDKey        requestIDKeyType = iota
	requestIDHeader                  = "X-Request-ID"
)

// ===== [ Types ] =====
type (
	requestIDKeyType int
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// RequestID - Http Handler 처리에 RequestID 관련 처리를 지원하는 미들웨어 구성
func RequestID(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestID := req.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = uuid.Must(uuid.NewV4()).String()
		}

		req.Header.Set(requestIDHeader, requestID)
		rw.Header().Set(requestIDHeader, requestID)

		handler.ServeHTTP(rw, req.WithContext(observability.RequestIDToContext(req.Context(), requestID)))
	})
}
