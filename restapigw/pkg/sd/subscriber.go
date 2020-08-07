// Package sd - Service Discovery 기능 제공 패키지
package sd

import "github.com/cloud-barista/cb-apigw/restapigw/pkg/config"

// ===== [ Constants and Variables ] =====
const ()

var ()

// ===== [ Types ] =====
type (
	// FixedSubscriber - Backend Hosts 정보 관리 형식 (갱신되지 않는 형식)
	FixedSubscriber []*config.HostConfig

	// Subscriber - Backend Hosts 정보 관리 인터페이스
	Subscriber interface {
		Hosts() ([]string, error)
	}

	// SubscriberFunc - Subcriber 사용을 위한 함수 어뎁터
	SubscriberFunc func() ([]string, error)

	// SubscriberFactory - 지정된 Backend 설정에 따른 Subscriber 구성 팩토리 함수 형식
	SubscriberFactory func(*config.BackendConfig) Subscriber
)

// ===== [ Implementations ] =====

// Hosts - Subscriber에서 관리되는 Hosts 반환
func (sf SubscriberFunc) Hosts() ([]string, error) { return sf() }

// Hosts - FixedSubcriber에서 관리되는 Hosts 반환
func (fs FixedSubscriber) Hosts() ([]string, error) {
	hosts := make([]string, 0)
	for _, hc := range fs {
		hosts = append(hosts, hc.Host)
	}
	return hosts, nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// FixedSubscrberFactory - 지정된 Backend 설정에 따른 Fixed Subscriber를 구성
func FixedSubscrberFactory(bConf *config.BackendConfig) Subscriber {
	return FixedSubscriber(bConf.Hosts)
}
