// Package balancer - Upstream을 선택하는데 사용할 간단한 로드 밸런서 기능 제공 패키지
package balancer

import (
	"errors"
	"math/rand"
	"reflect"
	"time"
)

// ===== [ Constants and Variables ] =====
var (
	// ErrEmptyBackendList - 로드 밸런싱 대상인 Upstream의 Backend list가 존재하지 않는 오류
	ErrEmptyBackendList = errors.New("can not elect backend, Backends empty")
	// ErrCannotElectBackend - 로드 밸런싱 대상 Backend를 선정할 수 없는 오류
	ErrCannotElectBackend = errors.New("can't elect backend")
	// ErrUnsupportedAlgorithm - 로드 밸런싱에 지원되지 않는 알고리즘이 지정된 경우 오류
	ErrUnsupportedAlgorithm = errors.New("unsupported balancing algorithm")

	// typeRegistry - 로드 밸런싱에 적용할 알고리듬 관리용
	typeRegistry = make(map[string]reflect.Type)
)

// ===== [ Types ] =====
type (
	// Balancer - 여러 가지 알고리듬에 적용될 로드 밸런싱 정보 처리용 인터페이스
	Balancer interface {
		Elect(hosts []*Target) (*Target, error)
	}

	// Target - 선정된 Backend Service에 적용할 로드 밸런싱 정보 관리 구조
	Target struct {
		Target string
		Weight int
	}
)

// ===== [ Implementations ] =====
// ===== [ Private Functions ] =====

// init - 패키지 초기화
func init() {
	rand.Seed(time.Now().UnixNano())
	typeRegistry["roundrobin"] = reflect.TypeOf(RoundrobinBalancer{})
	typeRegistry["rr"] = reflect.TypeOf(RoundrobinBalancer{})
	typeRegistry["weigth"] = reflect.TypeOf(WeightBalancer{})
}

// ===== [ Public Functions ] =====

// New - 지정한 알고리즘에 맞는 밸런서 인스턴스 생성
func New(balance string) (Balancer, error) {
	alg, ok := typeRegistry[balance]
	if !ok {
		return nil, ErrUnsupportedAlgorithm
	}
	return reflect.New(alg).Elem().Addr().Interface().(Balancer), nil
}
