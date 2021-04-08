package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/encoding"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/transport/http/client"
)

// ===== [ Constants and Variables ] =====

var (
	logger = logging.NewLogger()
)

// ===== [ Types ] =====

// responseError - Defines interface for response error
type responseError interface {
	Error() string
	Name() string
	StatusCode() int
}

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// NewHTTPProxyWithHTTPExecutor - 지정된 BackendConfig 와 HTTP Request Executor와 응답 처리에 사용할 Decoder를 설정한 Proxy 반환
func NewHTTPProxyWithHTTPExecutor(bconf *config.BackendConfig, hre client.HTTPRequestExecutor, dec encoding.Decoder) Proxy {
	if bconf.Encoding == encoding.NOOP {
		return NewHTTPProxyDetailed(bconf, hre, client.NoOpHTTPStatusHandler, NoOpHTTPResponseParser)
	}

	ef := NewEntityFormatter(bconf)
	rp := DefaultHTTPResponseParserFactory(HTTPResponseParserConfig{dec, ef})
	return NewHTTPProxyDetailed(bconf, hre, client.GetHTTPStatusHandler(bconf), rp)
}

// NewHTTPProxyDetailed - 지정된 BackendConfig와 HTTP Reqeust Executor와 응답 처리에 사용할 StatusHandler, Response Parser를 설정한 Proxy 반환
func NewHTTPProxyDetailed(bconf *config.BackendConfig, hre client.HTTPRequestExecutor, hsh client.HTTPStatusHandler, hrp HTTPResponseParser) Proxy {
	return func(ctx context.Context, req *Request) (*Response, error) {
		reqToBackend, err := http.NewRequest(strings.ToTitle(req.Method), req.URL.String(), req.Body)
		if err != nil {
			return nil, err
		}

		// Backend 호출에 필요한 Header 정보 설정
		reqToBackend.Header = make(map[string][]string, len(req.Headers))
		for k, v := range req.Headers {
			tmp := make([]string, len(v))
			copy(tmp, v)
			reqToBackend.Header[k] = tmp
		}

		// Body Size 정보 설정
		if req.Body != nil {
			if v, ok := req.Headers["Content-Length"]; ok && len(v) == 1 && v[0] != "chunked" {
				if size, err := strconv.Atoi(v[0]); err == nil {
					reqToBackend.ContentLength = int64(size)
				}
			}
		}

		// Backend 호출에 필요한 Query String 정보 설정
		if req.Query != nil {
			q := reqToBackend.URL.Query()
			for k, v := range req.Query {
				q.Add(k, v[0])
			}
			reqToBackend.URL.RawQuery = q.Encode()
		}

		logger.Debugf("[CallChain] HTTP to Backend > %s", reqToBackend.URL.String())

		// Backed 호출
		resp, err := hre(ctx, reqToBackend)
		if reqToBackend.Body != nil {
			reqToBackend.Body.Close()
		}

		select {
		case <-ctx.Done():
			// 호출이 종료된 경우 (ex. timeout)
			return nil, ctx.Err()
		default:
		}

		//
		if resp == nil && err != nil {
			return nil, err
		}

		// Response Status 처리
		resp, err = hsh(ctx, resp)
		if err != nil {
			// HTTPResponseError 구현체인지 검증
			if t, ok := err.(responseError); ok {
				if !req.IsBypass {
					// return_error_details로 처리되는 경우는 오류 제거하고 정상적인 반환으로 처리
					return &Response{
						Data: map[string]interface{}{
							fmt.Sprintf("error_%s", t.Name()): t,
						},
						Metadata: Metadata{StatusCode: t.StatusCode(), Message: t.Error()},
					}, nil
				} else if resp != nil {
					if parseData, parseError := hrp(ctx, resp); parseError != nil {
						return parseData, parseError
					} else {
						// Proxy Response 정보에 오류 상태 설정 및 반환
						parseData.Metadata.StatusCode = t.StatusCode()
						parseData.IsComplete = false
						return parseData, err
					}
				} else {
					return nil, err
				}
			}

			// WrappedError 여부 검증
			if we, ok := err.(core.WrappedError); ok {
				// Bypass 이면서 Response가 존재하는 경우
				if req.IsBypass && resp != nil {
					if parseData, parseError := hrp(ctx, resp); parseError != nil {
						return parseData, parseError
					} else {
						// Proxy Response 정보에 오류 상태 설정 및 반환
						parseData.Metadata.StatusCode = we.Code()
						parseData.IsComplete = false
						return parseData, we
					}
				}
				// Bypass가 아니거나 Response가 없는 경우
				return nil, we
			}

			// 기타 오류 상태
			return nil, err
		}

		// Response Parser 호출후 Proxy Response 정보 반환
		return hrp(ctx, resp)
	}
}

// NewRequestBuilderChain - Request 파라미터와 Backend Path를 설정한 Proxy 호출 체인을 생성한다.
func NewRequestBuilderChain(bConf *config.BackendConfig) CallChain {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, req *Request) (*Response, error) {
			r := req.Clone()
			// Bypass가 아닌 경우는 Path와 Method를 설정에 맞도록 재 구성
			if !req.IsBypass {
				r.GeneratePath(bConf.URLPattern)
				r.Method = bConf.Method
			}

			return next[0](ctx, &r)
		}
	}
}
