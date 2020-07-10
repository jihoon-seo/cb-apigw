// Package mux - Gorilla Mux 기반의 Metrics 기능 제공 패키지
package mux // ===== [ Constants and Variables ] =====
// ===== [ Types ] =====
import (
	"context"
	"net/http"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/metrics"
)

type (
	// Metrics - Mux 기반의 Routing 처리세 적용할 Metrics 구조
	Metrics struct {
		*metrics.Producer
	}
)

// ===== [ Implementations ] =====

// RunEndpoint - Gorilla Mux Engine 기반으로 http.Server 구동
func (m *Metrics) RunEndpoint(ctx context.Context, s *http.ServeMux, log logging.Logger) {
	server := &http.Server{
		Addr:    m.Config.ListenAddress,
		Handler: s,
	}

	go func() {
		// http.Server 실행 및 오류 logging
		log.Error(server.ListenAndServe())
	}()

	go func() {
		// http.Server 종료 처리
		<-ctx.Done()
		// shutting down the stats handler
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		server.Shutdown(ctx)
		cancel()
	}()
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// SetupAndRun - Gorilla Mux 기반으로 동작하는 Metric Producer를 생성하고 Collector 처리를 위한 Mux 기반의 Endpoint Server 구동
// func SetupAndRun(ctx context.Context, sConf config.ServiceConfig, log logging.Logger) *Metrics {
// 	// Metrics Producer 설정 및 생성
// 	metricProducer := Metrics{metrics.NewMetricsProducer(ctx, sConf, log)}

// 	if metricProducer.Config != nil && metricProducer.Config.ExposeMetrics {
// 		// Gin 기반 Server 구동 및 Endpoint 처리 설정
// 		metricProducer.RunEndpoint(ctx, metricProducer.NewEngine(sConf.Debug, log), log)
// 	}

// 	return &metricProducer
// }
