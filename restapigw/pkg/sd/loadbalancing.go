// Package sd -
package sd

import (
	"runtime"
	"sync/atomic"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/core/rand"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/errors"
)

// ===== [ Constants and Variables ] =====
const ()

var (
	// ErrNoHosts - Load Balancing 처리 대상 Host 가 지정되지 않은 경우 오류
	ErrNoHosts = errors.New("no host available")
)

// ===== [ Types ] =====
type (
	// Balancer - Backend Host에 대한 Load Balancing 전략 적용을 위한 인터페이스
	Balancer interface {
		Host() (string, error)
	}

	// balancer - Subscriber와 연계하기 위한 Balancer 구조
	balancer struct {
		subscriber Subscriber
	}

	// roundRobinLB - Roundrobin 정책 구조
	roundRobinLB struct {
		balancer
		counter uint64
	}

	// randomLB - Random 정책 구조
	randomLB struct {
		balancer
		rand func(uint32) uint32
	}

	// noBalancer - 처리대상 Load Balancer가 없는 경우 처리용 형식
	noBalancer string
)

// ===== [ Implementations ] =====

// Host - 연계된 Subscriber를 통해 대상 Host 반환
func (rrb *roundRobinLB) Host() (string, error) {
	hosts, err := rrb.hosts()
	if nil != err {
		return "", err
	}
	offset := (atomic.AddUint64(&rrb.counter, 1) - 1) % uint64(len(hosts))
	return hosts[offset], nil
}

// Host - 연계된 Subscriber를 통해 대상 Host 반환
func (rb *randomLB) Host() (string, error) {
	hosts, err := rb.hosts()
	if err != nil {
		return "", err
	}
	return hosts[int(rb.rand(uint32(len(hosts))))], nil
}

// hosts - 관리 중인 Subscriber의 Hosts 반환
func (b *balancer) hosts() ([]string, error) {
	hosts, err := b.subscriber.Hosts()
	if nil != err {
		return hosts, err
	}
	if 0 >= len(hosts) {
		return hosts, ErrNoHosts
	}
	return hosts, nil
}

// Host - Load Balancing 대상이 없는 경우에 처리할 Dummy Host 반환
func (nb noBalancer) Host() (string, error) {
	return string(nb), nil
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewBalancer - 지정한 Subscriber 정보를 기준으로 Load Balancer 생성
func NewBalancer(subscriber Subscriber) Balancer {
	// 사용 가능한 Process 수를 기준으로 Roundrobin 또는 Processor 별로 처리하는 Random Load balancer 구성
	if p := runtime.GOMAXPROCS(-1); p == 1 {
		return NewRoundRobinLB(subscriber)
	}
	return NewRandomLB(subscriber)
}

// NewRoundRobinLB - 지정한 Subscriber 정보를 기준으로 Roundrobin 정책을 적용한 Load Balancer 생성
func NewRoundRobinLB(subscriber Subscriber) Balancer {
	if hosts, ok := subscriber.(FixedSubscriber); ok && 1 == len(hosts) {
		return noBalancer(hosts[0].Host)
	}
	return &roundRobinLB{
		balancer: balancer{subscriber: subscriber},
		counter:  0,
	}
}

// NewRandomLB - 난수 생성을 통해서 랜덤으로 처리하는 Load Balancer 생성
func NewRandomLB(subscriber Subscriber) Balancer {
	if hosts, ok := subscriber.(FixedSubscriber); ok && 1 == len(hosts) {
		return noBalancer(hosts[0].Host)
	}
	return &randomLB{
		balancer: balancer{subscriber: subscriber},
		rand:     rand.Uint32n,
	}
}
