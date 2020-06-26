// Package proxy - Routes 처리를 위한 Proxy 기능 제공 패키지
package proxy

import (
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy/balancer"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
	"go.opencensus.io/plugin/ochttp"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
type (
	// Register - Router에 연계되는 proxy 정보를 관리하는 구조
	Register struct {
		router                    router.Router
		idleConnectionsPerHost    int
		idleConnectionTimeout     time.Duration
		idleConnectionPurgeTicker *time.Ticker
		flushInterval             time.Duration
		matcher                   *router.ListenPathMatcher
		isPublicEndpoint          bool
	}
)

// ===== [ Implementations ] =====

// UpdateRouter - 지정한 Router 정보로 재 설정
func (r *Register) UpdateRouter(router router.Router) {
	r.router = router
}

// Add - 지정한 설정 정보에 맞는 Route 정보 등록
func (r *Register) Add(rd *RouterDefinition) error {
	log := logging.GetLogger()
	log.WithField("balancing_algorithm", rd.Upstreams.Balancing).Debug("Using a load balancing algorithm")
	balancerInstance, err := balancer.New(rd.Upstreams.Balancing)
	if nil != err {
		msg := "Could not create a balancer"
		log.WithError(err).Error(msg)
		return errors.Wrap(err, msg)
	}

	// TODO: Stats Client 처리는?
	handler := NewBalancedReverseProxy(rd.Definition, balancerInstance)
	handler.FlushInterval = r.flushInterval
	handler.Transport = &ochttp.Transport{
		Base: transport.New(
			transport.WithIdleConnectionTime(r.idleConnectionTimeout),
			transport.WithIdleConnectionPurgeTicker(r.idleConnectionPurgeTicker),
			transport.WithInsecureSkipVerify(rd.InsecureSkipVerify),
			transport.WithDialTimeout(time.Duration(rd.ForwardingTimeouts.DialTimeout)),
			transport.WithResponseHeaderTimeout(time.Duration(rd.ForwardingTimeouts.ResponseHeaderTimeout)),
		),
	}

	if r.matcher.Match(rd.ListenPath) {
		r.doRegister(r.matcher.Extract(rd.ListenPath), rd, &ochttp.Handler{
			Handler:          handler,
			IsPublicEndpoint: r.isPublicEndpoint,
		})
	}

	r.doRegister(rd.ListenPath, rd, &ochttp.Handler{Handler: handler, IsPublicEndpoint: r.isPublicEndpoint})
	return nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewRegister - Routes 정보 등록을 위한 Proxy Register 인스턴스 생성
func NewRegister(opts ...RegisterOption) *Register {
	r := Register{
		matcher: router.NewListenPathMatcher(),
	}

	for _, opt := range opts {
		opt(&r)
	}

	return &r
}
