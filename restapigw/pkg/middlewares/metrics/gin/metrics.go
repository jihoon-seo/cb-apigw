package gin

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/metrics"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
	ginRouter "github.com/cloud-barista/cb-apigw/restapigw/pkg/router/gin"
	"github.com/gin-gonic/gin"
	gometricsExp "github.com/rcrowley/go-metrics/exp"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

type (
	// Collector - Metrics 기반의 Collector 구조
	Collector struct {
		*metrics.Producer
	}

	// ginResponseWriter - Gin 의 ResponseWriter에 Metrics 처리를 위한 구조
	ginResponseWriter struct {
		gin.ResponseWriter
		name  string
		begin time.Time
		rm    *metrics.RouterMetrics
	}
)

// ===== [ Implementations ] =====

// end - gin.ResponseWriter 와 Metrics를 end 시점 처리
func (grw *ginResponseWriter) end() {
	duration := time.Since(grw.begin)

	grw.rm.Counter("response", grw.name, "status", strconv.Itoa(grw.Status()), "count").Inc(1)
	grw.rm.Histogram("response", grw.name, "size").Update(int64(grw.Size()))
	grw.rm.Histogram("response", grw.name, "time").Update(int64(duration))
}

// HandlerFactory - 전달된 HandlerFactory 수행 전에 필요한 Metric 관련 처리를 수행하는 HandlerFactory 구성
func (c *Collector) HandlerFactory(hf ginRouter.HandlerFactory, log logging.Logger) ginRouter.HandlerFactory {
	if c.Config == nil || !c.Config.RouterEnabled {
		return hf
	}

	return NewHTTPHandlerFactory(c.Router, hf, log)
}

// RunEndpoint - Gin Engine 기반으로 http.Server 구동
func (c *Collector) RunEndpoint(ctx context.Context, engine *gin.Engine, log logging.Logger) {
	server := &http.Server{
		Addr:    c.Config.ListenAddress,
		Handler: engine,
	}

	go func() {
		// http.Server 실행 및 오류 logging
		log.Error(server.ListenAndServe())
	}()

	go func() {
		// http.Server 종료 처리
		<-ctx.Done()
		// shutting down the stats handler
		log.Info("[METRICS] shutting down the metrics handler.")
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		server.Shutdown(ctx)
		cancel()
	}()
}

// NewEngine - Stats 처리를 위한 Endpoint 역할을 담당하는 Gin Engine 생성
func (c *Collector) NewEngine(debugMode bool) *gin.Engine {
	if debugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true

	engine.GET("/metrics", c.NewExportHandler())
	return engine
}

// NewExportHandler - 수집된 Metrics를 JSON 포맷으로 노출하는 go-metrics에서 제공되는 http.Handler 생성
func (c *Collector) NewExportHandler() gin.HandlerFunc {
	return gin.WrapH(gometricsExp.ExpHandler(*c.Registry))
}

// ===== [ Private Functions ] =====

// ===== [ Public Functions ] =====

// NewHTTPHandlerFactory - Router 단위의 Metrics 처리를 수행하는 HandlerFactory 생성
func NewHTTPHandlerFactory(rm *metrics.RouterMetrics, hf ginRouter.HandlerFactory, log logging.Logger) ginRouter.HandlerFactory {
	return func(eConf *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		next := hf(eConf, p)

		// Endpoint 응답관련 Metric 처리기 등록
		//rm.RegisterResponseWriterMetrics(eConf.Endpoint)

		return func(c *gin.Context) {
			name := eConf.Endpoint
			// Bypass인 경우 실제 호출 URL로 처리
			if eConf.IsBypass {
				name = c.Request.URL.Path
			}

			rm.RegisterResponseWriterMetrics(name)

			// Metric 처리기를 반영한 Gin Response Writer 생성
			rw := &ginResponseWriter{c.Writer, name, time.Now(), rm}
			c.Writer = rw

			// Router 연결 Metric 처리
			rm.Connection()

			next(c)

			rw.end()

			// Router 종료 Metric 처리
			rm.Disconnection()
		}
	}
}

// // // SetupAndRun - Gin 기반으로 동작하는 Metric Producer를 생성하고 Collector 처리를 위한 Gin 기반의 Endpoint Server 구동
// // func SetupAndRun(ctx context.Context, sConf config.ServiceConfig, logger logging.Logger) *Metrics {
// // 	// Metrics Producer 설정 및 생성
// // 	metricProducer := Metrics{metrics.SetupAndCreate(ctx, sConf, logger)}

// // 	if metricProducer.Config != nil && metricProducer.Config.ExposeMetrics {
// // 		// Gin 기반 Server 구동 및 Endpoint 처리 설정
// // 		metricProducer.RunEndpoint(ctx, metricProducer.NewEngine(sConf.Debug, logger), logger)
// // 	}

// // 	return &metricProducer
// // }

// // // NewMetrics - GIN 기반으로 동작하는 Metric Producer응 연계하고 Collector 처리를 위한 Endpoint Server 구동
// // func NewMetrics(ctx context.Context, mp metrics.Metrics, debugMode bool) metrics.IMetrics {
// // 	metrics := Metrics{&mp}
// // 	if nil != metrics.Config && metrics.Config.ExposeMetrics {
// // 		metrics.RunEndpoint(ctx, metrics.NewEngine(debugMode))
// // 	}

// // 	return metrics
// // }

// New - Gin 기반의 Metrics 처리를 위한 Collector 인스턴스 생성
func New(ctx context.Context, mConf config.MWConfig, log logging.Logger, debugMode bool) *Collector {
	mc := Collector{metrics.New(ctx, mConf, log)}

	// 설정에 따라서 Metrics에 대한 Endpoint 설정
	if nil != mc.Config && mc.Config.ExposeMetrics {
		mc.RunEndpoint(ctx, mc.NewEngine(debugMode), log)
	}

	return &mc
}
