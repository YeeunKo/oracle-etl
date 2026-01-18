// Package sse는 Server-Sent Events 기능을 제공합니다
package sse

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBroadcaster(t *testing.T) {
	// Broadcaster가 올바르게 초기화되는지 테스트
	broadcaster := NewBroadcaster()

	assert.NotNil(t, broadcaster)
	assert.NotNil(t, broadcaster.register)
	assert.NotNil(t, broadcaster.unregister)
}

func TestBroadcaster_StartAndStop(t *testing.T) {
	// Broadcaster가 시작하고 종료되는지 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	// 약간의 시간 대기 후 종료
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Run이 종료될 시간 대기
	time.Sleep(50 * time.Millisecond)

	// 테스트가 패닉 없이 완료되면 성공
}

func TestBroadcaster_RegisterClient(t *testing.T) {
	// 클라이언트 등록 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	// 클라이언트 등록
	client := broadcaster.Register("TRPID-12345678")

	require.NotNil(t, client)
	assert.NotEmpty(t, client.ID)
	assert.Equal(t, "TRPID-12345678", client.TransportID)
	assert.NotNil(t, client.Events)
	assert.NotNil(t, client.Done)

	// 정리
	broadcaster.Unregister(client.ID)
}

func TestBroadcaster_UnregisterClient(t *testing.T) {
	// 클라이언트 해제 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	// 클라이언트 등록
	client := broadcaster.Register("TRPID-12345678")
	clientID := client.ID

	// 클라이언트 해제
	broadcaster.Unregister(clientID)

	// 해제 완료까지 대기
	time.Sleep(50 * time.Millisecond)

	// 클라이언트가 제거되었는지 확인
	_, exists := broadcaster.clients.Load(clientID)
	assert.False(t, exists, "클라이언트가 해제되어야 합니다")
}

func TestBroadcaster_Broadcast(t *testing.T) {
	// 특정 Transport에 이벤트 브로드캐스트 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	// 동일한 Transport에 대한 두 클라이언트 등록
	client1 := broadcaster.Register("TRPID-12345678")
	client2 := broadcaster.Register("TRPID-12345678")

	// 다른 Transport에 대한 클라이언트 등록
	client3 := broadcaster.Register("TRPID-87654321")

	defer broadcaster.Unregister(client1.ID)
	defer broadcaster.Unregister(client2.ID)
	defer broadcaster.Unregister(client3.ID)

	// 이벤트 브로드캐스트
	event := SSEEvent{
		Event: EventTypeProgress,
		Data: ProgressEvent{
			TransportID:   "TRPID-12345678",
			JobID:         "JOB-001",
			RowsProcessed: 1000,
		},
	}

	broadcaster.Broadcast("TRPID-12345678", event)

	// 클라이언트 1이 이벤트를 받는지 확인
	select {
	case received := <-client1.Events:
		assert.Equal(t, EventTypeProgress, received.Event)
	case <-time.After(time.Second):
		t.Error("클라이언트 1이 이벤트를 받지 못했습니다")
	}

	// 클라이언트 2도 이벤트를 받는지 확인
	select {
	case received := <-client2.Events:
		assert.Equal(t, EventTypeProgress, received.Event)
	case <-time.After(time.Second):
		t.Error("클라이언트 2가 이벤트를 받지 못했습니다")
	}

	// 클라이언트 3은 이벤트를 받지 않아야 함 (다른 Transport)
	select {
	case <-client3.Events:
		t.Error("클라이언트 3이 이벤트를 받으면 안됩니다")
	case <-time.After(100 * time.Millisecond):
		// 예상대로 이벤트 없음
	}
}

func TestBroadcaster_BroadcastProgress(t *testing.T) {
	// BroadcastProgress 헬퍼 메서드 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	client := broadcaster.Register("TRPID-12345678")
	defer broadcaster.Unregister(client.ID)

	// Progress 이벤트 브로드캐스트
	progressEvent := ProgressEvent{
		TransportID:     "TRPID-12345678",
		JobID:           "JOB-001",
		Table:           "VBRP",
		RowsProcessed:   5000,
		RowsTotal:       10000,
		RowsPerSecond:   1000.0,
		BytesWritten:    512000,
		ProgressPercent: 50.0,
	}

	broadcaster.BroadcastProgress(progressEvent)

	// 이벤트 수신 확인
	select {
	case received := <-client.Events:
		assert.Equal(t, EventTypeProgress, received.Event)
		data, ok := received.Data.(ProgressEvent)
		require.True(t, ok)
		assert.Equal(t, "VBRP", data.Table)
		assert.Equal(t, int64(5000), data.RowsProcessed)
	case <-time.After(time.Second):
		t.Error("Progress 이벤트를 받지 못했습니다")
	}
}

func TestBroadcaster_BroadcastStatus(t *testing.T) {
	// BroadcastStatus 헬퍼 메서드 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	client := broadcaster.Register("TRPID-12345678")
	defer broadcaster.Unregister(client.ID)

	// Status 이벤트 브로드캐스트
	statusEvent := StatusEvent{
		TransportID: "TRPID-12345678",
		JobID:       "JOB-001",
		Status:      StatusRunning,
		Message:     "추출 시작",
	}

	broadcaster.BroadcastStatus(statusEvent)

	// 이벤트 수신 확인
	select {
	case received := <-client.Events:
		assert.Equal(t, EventTypeStatus, received.Event)
		data, ok := received.Data.(StatusEvent)
		require.True(t, ok)
		assert.Equal(t, StatusRunning, data.Status)
	case <-time.After(time.Second):
		t.Error("Status 이벤트를 받지 못했습니다")
	}
}

func TestBroadcaster_BroadcastError(t *testing.T) {
	// BroadcastError 헬퍼 메서드 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	client := broadcaster.Register("TRPID-12345678")
	defer broadcaster.Unregister(client.ID)

	// Error 이벤트 브로드캐스트
	errorEvent := ErrorEvent{
		TransportID: "TRPID-12345678",
		JobID:       "JOB-001",
		Table:       "VBRP",
		Code:        "ORACLE_ERROR",
		Message:     "연결 실패",
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	broadcaster.BroadcastError(errorEvent)

	// 이벤트 수신 확인
	select {
	case received := <-client.Events:
		assert.Equal(t, EventTypeError, received.Event)
		data, ok := received.Data.(ErrorEvent)
		require.True(t, ok)
		assert.Equal(t, "ORACLE_ERROR", data.Code)
	case <-time.After(time.Second):
		t.Error("Error 이벤트를 받지 못했습니다")
	}
}

func TestBroadcaster_BroadcastComplete(t *testing.T) {
	// BroadcastComplete 헬퍼 메서드 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	client := broadcaster.Register("TRPID-12345678")
	defer broadcaster.Unregister(client.ID)

	// Complete 이벤트 브로드캐스트
	completeEvent := CompleteEvent{
		TransportID:   "TRPID-12345678",
		JobID:         "JOB-001",
		TotalRows:     100000,
		TotalBytes:    10485760,
		DurationMs:    8000,
		TablesCount:   5,
		RowsPerSecond: 12500.0,
	}

	broadcaster.BroadcastComplete(completeEvent)

	// 이벤트 수신 확인
	select {
	case received := <-client.Events:
		assert.Equal(t, EventTypeComplete, received.Event)
		data, ok := received.Data.(CompleteEvent)
		require.True(t, ok)
		assert.Equal(t, int64(100000), data.TotalRows)
	case <-time.After(time.Second):
		t.Error("Complete 이벤트를 받지 못했습니다")
	}
}

func TestBroadcaster_ClientCount(t *testing.T) {
	// 클라이언트 수 조회 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	// 초기 상태
	assert.Equal(t, 0, broadcaster.ClientCount())

	// 클라이언트 추가
	client1 := broadcaster.Register("TRPID-001")
	time.Sleep(10 * time.Millisecond) // 등록 완료 대기
	assert.Equal(t, 1, broadcaster.ClientCount())

	client2 := broadcaster.Register("TRPID-002")
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 2, broadcaster.ClientCount())

	// 클라이언트 제거
	broadcaster.Unregister(client1.ID)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, broadcaster.ClientCount())

	broadcaster.Unregister(client2.ID)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, broadcaster.ClientCount())
}

func TestBroadcaster_ClientCountForTransport(t *testing.T) {
	// 특정 Transport의 클라이언트 수 조회 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	// Transport 1에 대한 클라이언트 2개
	client1 := broadcaster.Register("TRPID-001")
	client2 := broadcaster.Register("TRPID-001")

	// Transport 2에 대한 클라이언트 1개
	client3 := broadcaster.Register("TRPID-002")

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 2, broadcaster.ClientCountForTransport("TRPID-001"))
	assert.Equal(t, 1, broadcaster.ClientCountForTransport("TRPID-002"))
	assert.Equal(t, 0, broadcaster.ClientCountForTransport("TRPID-999"))

	broadcaster.Unregister(client1.ID)
	broadcaster.Unregister(client2.ID)
	broadcaster.Unregister(client3.ID)
}

func TestBroadcaster_ConcurrentAccess(t *testing.T) {
	// 동시 접근 안전성 테스트
	broadcaster := NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	var wg sync.WaitGroup
	clientCount := 50

	// 동시에 여러 클라이언트 등록/해제
	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			client := broadcaster.Register("TRPID-CONCURRENT")
			time.Sleep(10 * time.Millisecond)
			broadcaster.Unregister(client.ID)
		}(i)
	}

	// 동시에 이벤트 브로드캐스트
	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			event := ProgressEvent{
				TransportID:   "TRPID-CONCURRENT",
				RowsProcessed: int64(idx * 100),
			}
			broadcaster.BroadcastProgress(event)
		}(i)
	}

	wg.Wait()

	// 모든 고루틴이 완료되면 성공 (패닉이나 데드락 없음)
}
