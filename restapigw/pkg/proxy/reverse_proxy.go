package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/observability"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy/balancer"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
	"github.com/go-chi/chi"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====

// addTraceAttribute - 지정한 Request에 추가적인 추적 정보 설정
func addTraceAttribute(req *http.Request) {
	log := logging.GetLogger()

	ctx := req.Context()
	span := trace.FromContext(ctx)
	if nil == span {
		return
	}

	host, err := os.Hostname()
	if host == "" || nil != err {
		log.WithError(err).Debug("Failed to get host name")
		host = "unknown"
	}

	span.AddAttributes(
		trace.StringAttribute("http.host", host),
		trace.StringAttribute("http.referrer", req.Referer()),
		trace.StringAttribute("http.remote_address", req.RemoteAddr),
		trace.StringAttribute("request.id", observability.RequestIDFromContext(ctx)),
	)
}

// applyParameters - 지정한 Request, 경로에 지장한 파라미터들을 설정
func applyParameters(req *http.Request, path string, paramNames []string) (string, error) {
	for _, paramName := range paramNames {
		paramValue := chi.URLParam(req, paramName)

		if len(paramValue) == 0 {
			return "", errors.Errorf("unable to extract {%s} from request", paramName)
		}

		path = strings.Replace(path, fmt.Sprintf("{%s}", paramName), paramValue, -1)
	}

	return path, nil
}

// singleJoiningSlash - 2개의 URL 연결 처리
func singleJoiningSlash(a, b string) string {
	a = cleanSlashes(a)
	b = cleanSlashes(b)

	aSlash := strings.HasSuffix(a, "/")
	bSlash := strings.HasPrefix(b, "/")

	switch {
	case aSlash && bSlash:
		return a + b[1:]
	case !aSlash && !bSlash:
		if len(b) > 0 {
			return a + "/" + b
		}
		return a
	}
	return a + b
}

// cleanSlashes - URL에 존재하는 중복된 '/'를 하나로 조정
func cleanSlashes(val string) string {
	endSlash := strings.HasSuffix(val, "//")
	startSlash := strings.HasPrefix(val, "//")

	if startSlash {
		val = "/" + strings.TrimPrefix(val, "//")
	}
	if endSlash {
		val = strings.TrimSuffix(val, "//") + "/"
	}

	return val
}

// createDirector - 지정한 정보를 기준으로 Request 전달 대상을 처리하는 Director 인스턴스 생성
func createDirector(d *Definition, balancer balancer.Balancer) func(req *http.Request) {
	log := logging.GetLogger()
	paramNameExtractor := router.NewListenPathParamNameExtractor()
	matcher := router.NewListenPathMatcher()

	return func(req *http.Request) {
		upstream, err := balancer.Elect(d.Upstreams.Targets.ToBalancerTargets())
		if nil != err {
			log.WithError(err).Error("Could not elect on upstream")
			return
		}

		targetURL := upstream.Target
		paramNames := paramNameExtractor.Extract(targetURL)
		parameterizedPath, err := applyParameters(req, targetURL, paramNames)
		if nil != err {
			log.WithError(err).Warn("Unable to extract param from request")
		} else {
			targetURL = parameterizedPath
		}

		log.WithField("target", targetURL).Debug("Target upstream elected")

		target, err := url.Parse(targetURL)
		if nil != err {
			log.WithError(err).WithField("upstream_url", targetURL).Error("Could not parse the target URL")
			return
		}

		originalURI := req.RequestURI
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		path := target.Path

		if d.AppendPath {
			log.Debug("Appending listen path to the target url")
			path = singleJoiningSlash(target.Path, req.URL.Path)
		}

		if d.StripPath {
			path = singleJoiningSlash(target.Path, req.URL.Path)
			listenPath := matcher.Extract(d.ListenPath)

			log.WithField("listen_path", listenPath).Debug("Stripping listen path")
			path = strings.Replace(path, listenPath, "", 1)
			if !strings.HasSuffix(target.Path, "/") && strings.HasSuffix(path, "/") {
				path = path[:len(path)-1]
			}
		}

		log.WithField("path", path).Debug("Upstream Path")
		req.URL.Path = path

		// HOST header에 대한 SSL 검증 문제 회피용
		if d.PreserveHost {
			log.Debug("Preserving the host header")
		} else {
			req.Host = target.Host
		}

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}

		// 복제된 Request를 수정한 후에는 원본 Request에 대한 정보를 확인할 수 없으므로 Log로 처리해 둔다.
		log.SetFields(logging.Fields{
			"request":          originalURI,
			"request-id":       observability.RequestIDFromContext(req.Context()),
			"upstream-host":    req.URL.Host,
			"upstream-request": req.URL.RequestURI(),
		}).Info("Proxing request to the following upstream")

		// TODO: State 처리

		// 추가적인 추적 정보 설정
		addTraceAttribute(req)

		// TODO: Opencensus Tag 처리
		ctx, _ := tag.New(req.Context(), tag.Insert(observability.KeyUpstreamPath, upstream.Target))
		*req = *req.WithContext(ctx)
	}
}

// ===== [ Public Functions ] =====

// TODO: State Client는?

// NewBalancedReverseProxy - 로드 밸런싱 처리에 대한 Reverse Proxy 인스턴스 생성
func NewBalancedReverseProxy(d *Definition, balancer balancer.Balancer) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: createDirector(d, balancer),
	}
}
