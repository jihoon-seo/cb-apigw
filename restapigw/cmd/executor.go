package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/api"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"

	// Opencensus 연동을 위한 Exporter 로드 및 초기화
	_ "github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/opencensus/exporters/jaeger"
)

// ===== [ Constants and Variables ] =====

// ===== [ Types ] =====

type ()

// ===== [ Implementations ] =====

// ===== [ Private Functions ] =====

// contextWithSignal - System Interrupt Signal을 반영한 Context 생
func contextWithSignal(ctx context.Context) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-signals:
			fmt.Println("Received System Interrupt")
			cancel()
			close(signals)
			fmt.Println("Signals channel closed")
		}
	}()
	return newCtx
}

// setupRepository - HTTP Server와 Admin Server 운영에 사용할 API Route 정보 리파지토리 구성
func setupRepository(sConf config.ServiceConfig, log logging.Logger) (api.Repository, error) {
	// API Routing 정보를 관리하는 Repository 구성
	repo, err := api.BuildRepository(sConf.Repository.DSN, sConf.Cluster.UpdateFrequency)
	if nil != err {
		return nil, errors.Wrap(err, "could not build a repository for the database or file or CB-Store")
	}

	return repo, nil
}

// ===== [ Public Functions ] =====

// SetupAndRun - API Gateway 서비스를 위한 Router 및 Pipeline 구성과 구동
func SetupAndRun(ctx context.Context, sConf config.ServiceConfig) error {
	// Sets up the Logger (CB-LOG)
	log := logging.NewLogger()

	// API 운영을 위한 라우팅 리파지토리 구성
	repo, err := setupRepository(sConf, *log)
	if nil != err {
		return err
	}
	defer repo.Close()

	// API G/W 및 Admin 구동을 위한 Router 설정 및 구동

	switch sConf.RouterEngine {
	default:
		SetupAndRunGinRouter(ctx, sConf, repo, *log)
	}

	return nil
}
