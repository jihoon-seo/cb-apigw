// Package limiter - Rate Limit 처리용 Token Bucket 구현 패키지
package limiter

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/logging"
)

// ===== [ Constants and Variables ] =====

const (
	// rateMargin - Rate 계산할 때 허용 가능한 변동율 (1%)
	rateMargin = 0.01
	// infinityDuration - Unlimited 표현을 위한 Int64 최대 값
	infinityDuration time.Duration = 0x7fffffffffffffff
	// nanosec
	nanosec = 1e9
)

var (
	// logger - Logging
	logger = logging.NewLogger()
)

// ===== [ Types ] =====

// IClock - Token Bucket의 시간 처리용 인터페이스
type IClock interface {
	// Now - 현재 시각
	Now() time.Time
	// Sleep - Sleep 처리
	Sleep(d time.Duration)
}

// realClock - 표준 시간 함수를 기준으로 Clock 인터페이스 구현을 위한 구조
type realClock struct{}

/**
 * 용어 정의:
 * - Tick : Time Interrupt (Resolution of the Go time package)기준으로 초당 1e9 개의 Tick (대략 10 nano sec)이 존재한다.
 * - Quantum : 구간마다 채워질 토큰 수 (기본 값 1)
 * - FillInterval : 초당 처리 수를 Tick 단위로 계산한 Tick 기간
 * - LatestTick : 최종 Token을 사용한 Tick
 *
 * 계산 알고리즘:
 * - Bucket의 Rate(Capacity) 를 기준으로 FillInterval을 계산한다.
 * - Bucket에서 Token을 사용(갱신)하는 메서드가 호출되면 LatestTick 처리 및 FillInterval에 따라서 Quantum 만큼 Token을 Bucket에 추가한다.
 * - 초당 처리되는 비율을 Tick 단위로 계산해서, FillInterval로 설정하고 각 FillInterval 마다 Quantum 만큼 Bucket에 Token을 채운다.
 */

// Bucket - Rate Limit 처리를 위한 Token Bucket 구조
type Bucket struct {
	// clock - Bucket 운영에 사용할 Clock (별도 지정이 없으면 시스템 Clock 사용)
	clock IClock
	// startTime - Bucket 생성 시각 (Tick 관리용)
	startTime time.Time
	// capacity -Bucket 최대 Token 용량 (Max Limit 관리용)
	capacity int64
	// quantum - FillInterval에 사용할 Token 수 (기본 값 1)
	quantum int64
	// fillInterval - Token을 추가하기 위한 기간 (Tick 단위)
	fillInterval time.Duration
	// mu - 동시성 관리용
	mu sync.Mutex
	// availableTokens - Latest Tick 기준 현재 사용 가능한 Token 수 (음수인 경우는 Token이 추가될 떄까지 대기 중)
	availableTokens int64
	// latestTick - Token을 처리한 최종 Tick 관리용
	latestTick int64
}

// ===== [ Implementations ] =====

/**
 * RealClock Implements
 */

// Now - 현재 시각 (time.Now)
func (realClock) Now() time.Time {
	return time.Now()
}

// Sleep - Sleep 처리 (time.Sleep)
func (realClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

/**
 * Bucket Implements
 */

// currentTick - Bucket의 생성 시각을 기준으로 지정된 시각까지의 시간을 FillInterval로 나눈 Tick 반환
func (tb *Bucket) currentTick(now time.Time) int64 {
	return int64(now.Sub(tb.startTime) / tb.fillInterval)
}

// adjustAvailableTokens - 지정된 Tick 정보를 기준으로 LatestTick 대비로 Bucket 내에 사용 가능한 Token 수 조정
func (tb *Bucket) adjustAvailableTokens(tick int64) {
	lastTick := tb.latestTick
	tb.latestTick = tick

	if tb.availableTokens >= tb.capacity {
		return
	}
	tb.availableTokens += (tick - lastTick) * tb.quantum
	if tb.availableTokens > tb.capacity {
		tb.availableTokens = tb.capacity
	}

	logger.Debugf("Adjust Available Token on TokenBucket - [Quantum : %d, fillInterval : %d, lastTick: %d, tick: %d, Added Token: %d]\n", tb.quantum, tb.fillInterval, lastTick, tick, (tick-lastTick)*tb.quantum)

	return
}

// take - 현재 시각을 기준으로 Blocking 없이 Bucket에서 지정한 갯수만큼의 Token을 가져오고, Token이 추가되어 사용 가능할 때까지의 대기 시간 반환 (지정한 최대 대기 시간을 초과하는 경우는 즉시 반환)
func (tb *Bucket) take(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool) {
	if count <= 0 {
		return 0, true
	}

	// 지정한 시각을 기준으로 사용 가능한 Token 수 조정
	tick := tb.currentTick(now)
	tb.adjustAvailableTokens(tick)

	// 사용 가능 Token이 존재하는 경우는 대기 없이 즉시 반환
	availableCount := tb.availableTokens - count
	if availableCount >= 0 {
		tb.availableTokens = availableCount
		return 0, true
	}

	// 사용 가능 Token이 부족한 경우는 대기 시간을 계산, 최대 대기 시간을 넘어가면 사용 불가 상태로 즉시 반환
	endTick := tick + (-availableCount+tb.quantum-1)/tb.quantum
	endTime := tb.startTime.Add(time.Duration(endTick) * tb.fillInterval)
	waitTime := endTime.Sub(now)
	if waitTime > maxWait {
		return 0, false
	}

	// 대기 시간과 사용가능 Token수 설정
	tb.availableTokens = availableCount
	return waitTime, true
}

// takeAvailable - 현재 시간 기준으로 지정한 갯수의 Token을 사용하는 것으로 처리하고 사용 가능한 Token 갯수 반환
func (tb *Bucket) takeAvailable(now time.Time, count int64) int64 {
	if count <= 0 {
		return 0
	}

	// 현재 Tick 기준으로 사용 가능 Token 개수 조정
	tb.adjustAvailableTokens(tb.currentTick(now))
	if tb.availableTokens <= 0 {
		return 0
	}

	// 부족한 경우는 부족한 수량만 반환
	if count > tb.availableTokens {
		count = tb.availableTokens
	}

	tb.availableTokens -= count
	return count
}

// available - 현재 시각 기준으로 Bucket 내의 가용 Token 수 반환
func (tb *Bucket) available(now time.Time) int64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 현재 Tick 기준으로 사용 가능 Token 개수 조정
	tb.adjustAvailableTokens(tb.currentTick(now))
	return tb.availableTokens
}

// Take - Blocking 없이 Bucket에서 지정한 갯수의 Token을 가져오고, Token이 추가되어 사용 가능할 떄까지의 대기 시간 반환
func (tb *Bucket) Take(count int64) time.Duration {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	d, _ := tb.take(tb.clock.Now(), count, infinityDuration)
	return d
}

// TakeAvailable - 현재 시간 기준으로 지정한 갯수의 Token을 사용하는 것으로 처리하고 사용 가능한 Token 갯수 반환
func (tb *Bucket) TakeAvailable(count int64) int64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	return tb.takeAvailable(tb.clock.Now(), count)
}

// TakeMaxDuration - 전달한 갯수의 Token을 얻기 위해 대기하는 시간과 Token 존재 여부 검증 (지정한 최대 대기 시간을 초과하면 처리 불가로 즉시 반환)
func (tb *Bucket) TakeMaxDuration(count int64, maxWait time.Duration) (time.Duration, bool) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	return tb.take(tb.clock.Now(), count, maxWait)
}

// Wait - Bucket 내의 가용 Token 수를 기준으로 전달된 갯수의 Token이 사용 가능할 때까지 대기 처리
func (tb *Bucket) Wait(count int64) {
	if d := tb.Take(count); d > 0 {
		tb.clock.Sleep(d)
	}
}

// WaitMaxDuration - Bucket 내의 가용 Token 수를 기준으로 전달된 갯수의 Token이 사용 가능할 떄까지 최대 대기 (최대 대기 시간이 초과되면 Token을 삭제하고 삭제될 Token이 없는 경우는 즉시 반환)
func (tb *Bucket) WaitMaxDuration(count int64, maxWait time.Duration) bool {
	d, ok := tb.TakeMaxDuration(count, maxWait)
	if d > 0 {
		tb.clock.Sleep(d)
	}
	return ok
}

// Rate - 초당 채워지는 Token 비율 반환
func (tb *Bucket) Rate() float64 {
	return nanosec * float64(tb.quantum) / float64(tb.fillInterval)
}

// Available - Bucket 내의 가용 Token 개수 반환
func (tb *Bucket) Available() int64 {
	return tb.available(tb.clock.Now())
}

// Capacity - Bucket 의 최대 용량 반환
func (tb *Bucket) Capacity() int64 {
	return tb.capacity
}

// ===== [ Private Functions ] =====

// nextQuantum - 전달된 Quantum 보다 큰 최소의 정수 반환
func nextQuantum(q int64) int64 {
	q1 := q * 11 / 10
	if q1 == q {
		q1++
	}
	return q1
}

// ===== [ Public Functions ] =====

// NewBucket - 지정한 최대 용량까지 지정한 기간마다 1개의 Token을 채우는 Token Bucket 생성 (생성된 Bucket은 최대 용량의 Token이 설정되어 있다)
func NewBucket(fillInterval time.Duration, capacity int64) *Bucket {
	return NewBucketWithClock(fillInterval, capacity, nil)
}

// NewBucketWithRate - 지정한 비율로 최대 용량까지 Token을 채우는 Token Bucket 생성 (Clock 정확도가 제한되므로 높은 비율을 지정했을 경우 실제 비율은 1% 정도의 차이가 발생할 수 있다)
func NewBucketWithRate(rate float64, capacity int64) *Bucket {
	return NewBucketWithRateAndClock(rate, capacity, nil)
}

// NewBucketWithRateAndClock - NewBucketWithRate와 동일하며, 운영을 위한 Clock 인터페이스 구현 적용된 Token Bucket 생성
func NewBucketWithRateAndClock(rate float64, capacity int64, clock IClock) *Bucket {
	// 1ns 단위로 1개의 Token을 최대 용량까지 처리하는 Token Bucket 생성
	tb := NewBucketWithQuantumAndClock(1, capacity, 1, clock)

	// 지정한 비율로 Quantum과 FillInterval 결정 (1 ~ 양의 정수 기준 최대 값까지 검증)
	for quantum := int64(1); quantum < 1<<50; quantum = nextQuantum(quantum) {
		// 1초에 허용되는 비율을 기준으로 Token이 채워지는 기간을 계산한다.
		fillInterval := time.Duration(nanosec * float64(quantum) / rate)
		if fillInterval <= 0 {
			continue
		}
		tb.fillInterval = fillInterval
		tb.quantum = quantum

		// 계산된 Quantum과 FillInterval을 기준으로 허용되는 비율과 지정한 비율의 편차를 검증한다. (허용 편차를 벗어나는 경우는 재 계산)
		if diff := math.Abs(tb.Rate() - rate); diff/rate <= rateMargin {
			logger.Debugf("[Quantum : %d, fillInterval : %d, Specified Rate: %f, Bucket Rate: %f]\n", quantum, fillInterval, rate, tb.Rate())
			return tb
		}
	}

	panic("Cannot find suitable quantum for " + strconv.FormatFloat(rate, 'g', -1, 64))
}

// NewBucketWithClock - NewBucket과 동일하며, 운영을 위한 Clock 인터페이스 구현 적용된 Token Bucket 생성
func NewBucketWithClock(fillInterval time.Duration, capacity int64, clock IClock) *Bucket {
	return NewBucketWithQuantumAndClock(fillInterval, capacity, 1, clock)
}

// NewBucketWithQuantum - NewBucket과 유사하지만, 지정한 기간마다 채워지는 Token의 수를 지정한 Token Bucket 생성
func NewBucketWithQuantum(fillInterval time.Duration, capacity, quantum int64) *Bucket {
	return NewBucketWithQuantumAndClock(fillInterval, capacity, quantum, nil)
}

// NewBucketWithQuantumAndClock - NewBucketWithQuantum과 동일하며, 운영을 위한 Clock 인터페이스 구현 적용된 Token Bucket 쌩성 (Clock이 지정되지 않으면 시스템 사용)
func NewBucketWithQuantumAndClock(fillInterval time.Duration, capacity, quantum int64, clock IClock) *Bucket {
	if clock == nil {
		clock = realClock{}
	}
	if fillInterval <= 0 {
		panic("The token bucket fill interval must be positive.")
	}
	if capacity <= 0 {
		panic("The token bucket capacity must be positive.")
	}
	if quantum <= 0 {
		panic("The token bucket quantum must be positive.")
	}

	return &Bucket{
		clock:           clock,
		startTime:       clock.Now(),
		latestTick:      0,
		fillInterval:    fillInterval,
		capacity:        capacity,
		quantum:         quantum,
		availableTokens: capacity,
	}
}
