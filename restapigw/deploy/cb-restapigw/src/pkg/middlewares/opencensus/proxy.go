package opencensus

import (
	"context"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
	"go.opencensus.io/trace"
)

// ===== [ Constants and Variables ] =====

const (
	// errCtxCanceled - Context가 취소된 경우 오류
	errCtxCanceledMsg = "context canceled"
)

// ===== [ Types ] =====

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// CallChain - 지정한 이름을 기준으로 순차적으로 실행되는 Call chain 함수 반환
func CallChain(name string) proxy.CallChain {
	if !IsProxyEnabled() {
		return proxy.EmptyChain
	}
	return func(next ...proxy.Proxy) proxy.Proxy {
		if len(next) > 1 {
			panic(proxy.ErrTooManyProxies)
		}
		if len(next) < 1 {
			panic(proxy.ErrNotEnoughProxies)
		}

		return func(ctx context.Context, req *proxy.Request) (*proxy.Response, error) {
			var span *trace.Span
			ctx, span = trace.StartSpan(trace.NewContext(ctx, fromContext(ctx)), name)
			resp, err := next[0](ctx, req)
			if err != nil {
				if err.Error() != errCtxCanceledMsg {
					if resp != nil {
						span.SetStatus(trace.Status{Code: int32(resp.Metadata.StatusCode), Message: err.Error()})
					} else {
						span.SetStatus(trace.Status{Code: 500, Message: err.Error()})
					}
				} else {
					span.AddAttributes(trace.BoolAttribute("error", true))
					span.AddAttributes(trace.BoolAttribute("canceled", true))
				}
			}
			span.AddAttributes(trace.BoolAttribute(("complete"), resp != nil && resp.IsComplete))
			span.End()

			return resp, err
		}
	}
}

// ProxyFactory - Opencensus Trace와 연동되는 Proxy Call chain 구성을 위한 팩토리
func ProxyFactory(pf proxy.Factory) proxy.FactoryFunc {
	if !IsProxyEnabled() {
		return pf.New
	}
	return func(eConf *config.EndpointConfig) (proxy.Proxy, error) {
		next, err := pf.New(eConf)
		if err != nil {
			return next, err
		}
		return CallChain("[proxy] " + eConf.Endpoint)(next), nil
	}
}

// BackendFactory - Opencensus Trace와 연동되는 Backend Call chain 구성을 위한 팩토리
func BackendFactory(bf proxy.BackendFactory) proxy.BackendFactory {
	if !IsBackendEnabled() {
		return bf
	}
	return func(bConf *config.BackendConfig) proxy.Proxy {
		return CallChain("[backend] " + bConf.URLPattern)(bf(bConf))
	}
}
