package plugin

import (
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/router"
)

// ===== [ Constants and Variables ] =====
const (
	// StartupEvent - startup 이벤트 식별자
	StartupEvent = "startup"
	// AdminAPIStartupEvent - Admin API startup 이벤트 식별자
	AdminAPIStartupEvent = "admin_startup"
	// ReloadEvent - 설정 다시 로드 이벤트 식별자
	ReloadEvent = "reload"
	// ShutdownEvent - 종료 이벤트 식별자
	ShutdownEvent = "shutdown"
	// SetupEvent - 구성 이벤트 식별자
	SetupEvent = "setup"
)

// ===== [ Types ] =====
type (
	// OnStartup - REST API G/W 시작 시점의 이벤트 구조
	OnStartup struct {
		Register       *proxy.Register
		Config         *config.ServiceConfig
		Configurations []*config.EndpointConfig
	}

	// OnReload - REST API G/W의 설정이 변경되어 다시 로드되는 시점의 이벤트 구조
	OnReload struct {
		Configurations []*config.EndpointConfig
	}

	// OnAdminAPIStartup - REST API G/W 의 관리자 API 시작 시점의 이벤트 구조
	OnAdminAPIStartup struct {
		Router router.Router
	}
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====
