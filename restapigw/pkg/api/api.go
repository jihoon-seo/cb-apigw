// Package api - ADMIN API 기능을 제공하는 패키지
package api

import (
	"reflect"

	"github.com/asaskevich/govalidator"
)

// ===== [ Constants and Variables ] =====
// ===== [ Types ] =====

type (
	// ConfigurationOperation - Configuration 변경에 연계되는 Operation 형식
	ConfigurationOperation int

	// TODO: Endpoint/Backend Configuration here!!!

	// Definition - API 처리를 위한 Endpoint 설정정보 형식
	Definition struct {
		Name string `mapstructure:"name" yaml:"name" bson:"name" json:"name" valid:"required~name is required,matches(^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$)~name cannot contain non-URL friendly characters"`
	}

	// Configuration - API Definition
	Configuration struct {
		Definitions []*Definition
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

// Validate - Routing Definition 검증
func (d *Definition) Validate() (bool, error) {
	return govalidator.ValidateStruct(d)
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewDefinition - 기본 값으로 설정된 API Routing Definition 인스턴스 생성
func NewDefinition() *Definition {
	return &Definition{}
}
