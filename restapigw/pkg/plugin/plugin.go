// Package plugin - REST API G/W에서 사용할 Plugin 운영 지원 패키지
package plugin

import (
	"errors"
	"sync"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
)

// ===== [ Constants and Variables ] =====
var (
	lock sync.RWMutex

	// Plugin 운영을 위한 관리 Map
	plugins = make(map[string]Plugin)

	// 이벤트 처리를 위한 Plugin연계용 Hook 관리 map
	eventHooks = make(map[string][]EventHook)
)

// ===== [ Types ] =====
type (
	// Config - Plugin 설정정보 Map
	Config map[string]interface{}
	// EventHook - Plugin에 전달할 이벤트 처리 함수
	EventHook func(event interface{}) error

	// SetupFunc - Plugin 구성 및 실행을 위한 함수
	SetupFunc func(bConf *config.BackendConfig, rawConfig Config) error
	// ValidateFunc - Plugin에 전달될 설정 정보 검증 함수
	ValidateFunc func(rawConfig Config) (bool, error)

	// Plugin - Plugin 관리 구조
	Plugin struct {
		Action   SetupFunc
		Validate ValidateFunc
	}
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// EmitEvent - EventType을 아큐먼트로 전달하는 이벤트 발생 처리
func EmitEvent(name string, event interface{}) error {
	log := logging.GetLogger()

	hooks, found := eventHooks[name]
	if !found {
		return errors.New("Plugin not found")
	}

	for _, hook := range hooks {
		err := hook(event)
		if nil != err {
			log.WithError(err).WithField("event_name", name).Warn("an error occurred when an event was triggered")
		}
	}

	return nil
}
