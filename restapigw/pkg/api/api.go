// Package api - ADMIN API 기능을 제공하는 패키지
package api

import (
	"reflect"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/config"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/proxy"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// Plugin - API 사용될 Plugin 정보 구조
	Plugin struct {
		Name    string                 `mapstructure:"name" bson:"name" json:"name"`
		Enabled bool                   `mapstructure:"enabled" bson:"enabled" json:"enabled"`
		Config  map[string]interface{} `mapstructure:"config" bson:"config" json:"config"`
	}

	// HealthCheck - Health Checking 설정 구조
	HealthCheck struct {
		URL     string `mapstructure:"url" bson:"url" json:"url"  valid:"url"`
		Timeout int    `mapstructure:"timeout" bson:"timeout" json:"timeout"`
	}

	// Definition - Proxy로 처리할 API 구조
	Definition struct {
		Name        string            `mapstructure:"name" bson:"name" json:"name" valid:"required-name is required,matches(^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$)~name cannot contain non-URL friendly characters"`
		Active      bool              `mapstructure:"active" bson:"active" json:"active"`
		Proxy       *proxy.Definition `mapstructure:"proxy" bson:"proxy" json:"proxy" valid:"required"`
		Plugins     []Plugin          `mapstructure:"plugins" bson:"plugins" json:"plugins"`
		HealthCheck HealthCheck       `mapstructure:"health_check" bson:"heal_check" json:"health_check"`
	}

	// ConfigurationOperation - Configuration 변경에 연계되는 Operation 형식
	ConfigurationOperation int

	// Configuration - API Endpoints
	Configuration struct {
		Definitions []*config.EndpointConfig
	}

	// ConfigurationChanged - Configuration 변경이 발생했을 떄 전송되는 메시지 처리 형식
	ConfigurationChanged struct {
		Configurations *Configuration
	}

	// ConfigurationMessage - Configuration 변경시 사용할 메시지 형식
	ConfigurationMessage struct {
		Operation     ConfigurationOperation
		Configuration *Configuration
	}
)

// ===== [ Implementations ] =====

// EqualsTo - 현재 동작 중인 설정과 지정된 설정이 동일한지 여부 검증
func (c *Configuration) EqualsTo(tc *Configuration) bool {
	return reflect.DeepEqual(c, tc)
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewEndpoint - 기본 값으로 설정된 API Routing Endpoint 인스턴스 생성
func NewEndpoint() *config.EndpointConfig {
	return &config.EndpointConfig{}
}
