// Package server -
package server

import (
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/metrics"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Option - REST API G/W Server 운영에 필요한 옵션 서정을 위한 함수 형식
	Option func(*Server)
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// WithLog - Logging에 사용할 Logger 적용
func WithLog(log *logging.Logger) Option {
	return func(s *Server) {
		s.logger = log
	}
}

// WithSystemConfig - System Configuration 적용
func WithSystemConfig(sConf *config.ServiceConfig) Option {
	return func(s *Server) {
		s.serviceConf = sConf
	}
}

// WithProvider - 지정한 Configuration Provider 적용
func WithProvider(provider api.Repository) Option {
	return func(s *Server) {
		s.provider = provider
	}
}

// WithMetricsProducer - 지정한 Metrics Prodcuer 적용
func WithMetricsProducer(mp *metrics.Metrics) Option {
	return func(s *Server) {
		s.metricsProducer = mp
	}
}

// // WithPipeConfig - Route 와 Backend 연계 구성을 위한 Pipeline 설정 적용
// func WithPipeConfig(pc *router.PipeConfig) Option {
// 	return func(s *Server) {
// 		s.pipeConf = pc
// 	}
// }

// WithProxyFactory - Backend 연계를 위한 Proxy Factory 적용
func WithProxyFactory(pf proxy.Factory) Option {
	return func(s *Server) {
		s.proxyFactory = pf
	}
}

// // WithHandlerFactory - OpenCensus와 Middleware 처리를 반영한 Endpoint Handler 적용
// func WithHandlerFactory(hf router.HandlerFactory) Option {
// 	return func(s *Server) {
// 		s.handlerFactory = hf
// 	}
// }
