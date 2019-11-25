package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/encoding"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/transport/http/client"
)

// ===== [ Constants and Variables ] =====

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

		reqToBackend.Header = make(map[string][]string, len(req.Headers))
		for k, v := range req.Headers {
			tmp := make([]string, len(v))
			copy(tmp, v)
			reqToBackend.Header[k] = tmp
		}

		if req.Body != nil {
			if v, ok := req.Headers["Content-Length"]; ok && len(v) == 1 && v[0] != "chunked" {
				if size, err := strconv.Atoi(v[0]); err == nil {
					reqToBackend.ContentLength = int64(size)
				}
			}
		}

		// Backed 호출
		resp, err := hre(ctx, reqToBackend)
		if reqToBackend.Body != nil {
			reqToBackend.Body.Close()
		}

		select {
		case <-ctx.Done():
			// 호출에 문제가 생긴 경우
			return nil, ctx.Err()
		default:
		}

		// 응답이 없어서 상태 확인이 안되는 경우
		if resp == nil && err != nil {
		 	return nil, err
		}

		// Response Status 처리
		resp, err = hsh(ctx, resp)
		if err != nil {
			if t, ok := err.(responseError); ok {
				return &Response{
					Data: map[string]interface{}{
						fmt.Sprintf("error_%s", t.Name()): t,
					},
					Metadata: Metadata{StatusCode: t.StatusCode()},
				}, nil
			}
			return nil, err
		}

		// Response Parser 호출
		return hrp(ctx, resp)
	}
}

// NewRequestBuilderChain - Creates a proxy chain that parses the request params received from the outter layer and generates the path to the backend endpoint
func NewRequestBuilderChain(bConf *config.BackendConfig) CallChain {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, req *Request) (*Response, error) {
			r := req.Clone()
			r.GeneratePath(bConf.URLPattern)
			r.Method = bConf.Method

			// =====
			// FIXME: Loadbalancer가 추가되면 아래의 부분은 Loadbalancer 처리로 대체
			// Builds the URL for Backend call (Currently use only first url)
			var b strings.Builder
			b.WriteString(bConf.Host[0])
			b.WriteString(r.Path) // generated path replaced with paramters
			newURL, err := url.Parse(b.String())
			if err != nil {
				return nil, err
			}
			r.URL = newURL
			// =====
			return next[0](ctx, &r)
		}
	}
}
