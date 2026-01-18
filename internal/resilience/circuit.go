// Package resilience는 시스템 복원력을 위한 패턴을 제공합니다.
// M7-03: Circuit Breaker
package resilience

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen은 Circuit Breaker가 열려있을 때 반환되는 에러입니다
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State는 Circuit Breaker의 상태를 나타냅니다
type State int

const (
	// StateClosed는 정상 상태 (요청 허용)
	StateClosed State = iota
	// StateOpen은 차단 상태 (요청 거부)
	StateOpen
	// StateHalfOpen은 테스트 상태 (일부 요청 허용)
	StateHalfOpen
)

// String은 State를 문자열로 변환합니다
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitConfig는 Circuit Breaker 설정입니다
type CircuitConfig struct {
	// FailureThreshold는 Open 상태로 전환하기 위한 연속 실패 횟수
	FailureThreshold int
	// SuccessThreshold는 Closed 상태로 전환하기 위한 연속 성공 횟수
	SuccessThreshold int
	// Timeout은 Open 상태에서 Half-Open으로 전환하기까지의 대기 시간
	Timeout time.Duration
}

// DefaultCircuitConfig는 기본 Circuit Breaker 설정을 반환합니다
func DefaultCircuitConfig() CircuitConfig {
	return CircuitConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
	}
}

// CircuitMetrics는 Circuit Breaker 메트릭입니다
type CircuitMetrics struct {
	TotalRequests int64 // 총 요청 수
	Successes     int64 // 성공 수
	Failures      int64 // 실패 수
	Rejections    int64 // 거부 수 (Open 상태에서)
}

// CircuitBreaker는 Circuit Breaker 패턴을 구현합니다
type CircuitBreaker struct {
	mu sync.RWMutex

	config CircuitConfig
	state  State

	failures  int // 연속 실패 횟수
	successes int // 연속 성공 횟수 (Half-Open 상태에서)

	lastFailure time.Time // 마지막 실패 시간

	// 메트릭
	totalRequests int64
	totalSuccesses int64
	totalFailures int64
	totalRejections int64
}

// NewCircuitBreaker는 새로운 Circuit Breaker를 생성합니다
func NewCircuitBreaker(config CircuitConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// State는 현재 Circuit Breaker 상태를 반환합니다
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Open 상태에서 타임아웃 확인
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) >= cb.config.Timeout {
			return StateHalfOpen
		}
	}

	return cb.state
}

// Execute는 주어진 함수를 Circuit Breaker로 보호하여 실행합니다
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Open 상태 확인 및 타임아웃 처리
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) >= cb.config.Timeout {
			// Half-Open으로 전환
			cb.state = StateHalfOpen
			cb.successes = 0
		} else {
			// 여전히 Open - 요청 거부
			cb.totalRejections++
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}

	cb.totalRequests++
	cb.mu.Unlock()

	// 함수 실행
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}

	return err
}

// onSuccess는 성공 시 호출됩니다 (락 보유 상태)
func (cb *CircuitBreaker) onSuccess() {
	cb.totalSuccesses++
	cb.failures = 0 // 연속 실패 리셋

	switch cb.state {
	case StateClosed:
		// Closed 상태에서 성공 - 상태 유지
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			// 충분한 성공 - Closed로 전환
			cb.state = StateClosed
			cb.successes = 0
		}
	}
}

// onFailure는 실패 시 호출됩니다 (락 보유 상태)
func (cb *CircuitBreaker) onFailure() {
	cb.totalFailures++
	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.FailureThreshold {
			// 임계값 초과 - Open으로 전환
			cb.state = StateOpen
		}
	case StateHalfOpen:
		// Half-Open에서 실패 - 다시 Open으로
		cb.state = StateOpen
		cb.successes = 0
	}
}

// Metrics는 Circuit Breaker 메트릭을 반환합니다
func (cb *CircuitBreaker) Metrics() *CircuitMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return &CircuitMetrics{
		TotalRequests: cb.totalRequests,
		Successes:     cb.totalSuccesses,
		Failures:      cb.totalFailures,
		Rejections:    cb.totalRejections,
	}
}

// Reset은 Circuit Breaker를 초기 상태로 리셋합니다
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.totalRequests = 0
	cb.totalSuccesses = 0
	cb.totalFailures = 0
	cb.totalRejections = 0
}
