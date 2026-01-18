// Package sse는 Server-Sent Events 기능을 제공합니다
package sse

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// Client는 SSE 연결을 나타냅니다
type Client struct {
	ID          string        // 고유 클라이언트 ID
	TransportID string        // 필터링을 위한 Transport ID
	Events      chan SSEEvent // 이벤트 수신 채널
	Done        chan struct{} // 종료 신호 채널
}

// registrationMessage는 등록 메시지입니다
type registrationMessage struct {
	client *Client
	done   chan struct{} // 등록 완료 신호
}

// Broadcaster는 SSE 연결을 관리합니다
type Broadcaster struct {
	clients     sync.Map                   // map[clientID]*Client (스레드 안전)
	register    chan registrationMessage   // 클라이언트 등록 채널
	unregister  chan string                // 클라이언트 해제 채널
	clientCount int32                      // 총 클라이언트 수 (atomic)
	mu          sync.RWMutex //nolint:unused // 동기화용 뮤텍스
}

// NewBroadcaster는 새로운 Broadcaster를 생성합니다
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		register:   make(chan registrationMessage, 100),
		unregister: make(chan string, 100),
	}
}

// Run은 Broadcaster를 실행합니다
// 컨텍스트가 취소될 때까지 실행됩니다
func (b *Broadcaster) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// 모든 클라이언트에게 종료 신호 전송
			b.clients.Range(func(key, value interface{}) bool {
				if client, ok := value.(*Client); ok {
					select {
					case <-client.Done:
						// 이미 닫힘
					default:
						close(client.Done)
					}
				}
				return true
			})
			return

		case msg := <-b.register:
			b.clients.Store(msg.client.ID, msg.client)
			atomic.AddInt32(&b.clientCount, 1)
			// 등록 완료 신호
			close(msg.done)

		case clientID := <-b.unregister:
			if value, ok := b.clients.LoadAndDelete(clientID); ok {
				atomic.AddInt32(&b.clientCount, -1)
				if client, ok := value.(*Client); ok {
					// Done 채널을 먼저 닫아 클라이언트에게 종료 신호
					select {
					case <-client.Done:
						// 이미 닫힘
					default:
						close(client.Done)
					}
				}
			}
		}
	}
}

// Register는 새 클라이언트를 등록합니다
func (b *Broadcaster) Register(transportID string) *Client {
	client := &Client{
		ID:          uuid.New().String(),
		TransportID: transportID,
		Events:      make(chan SSEEvent, 100), // 버퍼링된 채널
		Done:        make(chan struct{}),
	}

	done := make(chan struct{})
	b.register <- registrationMessage{client: client, done: done}
	
	// 등록 완료 대기
	<-done

	return client
}

// Unregister는 클라이언트를 해제합니다
func (b *Broadcaster) Unregister(clientID string) {
	b.unregister <- clientID
}

// Broadcast는 특정 Transport의 모든 클라이언트에게 이벤트를 전송합니다
func (b *Broadcaster) Broadcast(transportID string, event SSEEvent) {
	b.clients.Range(func(key, value interface{}) bool {
		client, ok := value.(*Client)
		if !ok {
			return true
		}

		// TransportID가 일치하는 클라이언트에게만 전송
		if client.TransportID == transportID {
			// Done 채널 확인하여 종료된 클라이언트에게 전송 방지
			select {
			case <-client.Done:
				// 클라이언트가 종료됨
				return true
			default:
			}

			// 논블로킹 전송 (채널이 가득 찬 경우 이벤트 드롭)
			select {
			case client.Events <- event:
			case <-client.Done:
				// 클라이언트가 종료됨
			default:
				// 채널이 가득 찬 경우 이벤트 스킵
			}
		}
		return true
	})
}

// BroadcastProgress는 Progress 이벤트를 브로드캐스트합니다
func (b *Broadcaster) BroadcastProgress(event ProgressEvent) {
	b.Broadcast(event.TransportID, SSEEvent{
		Event: EventTypeProgress,
		Data:  event,
	})
}

// BroadcastStatus는 Status 이벤트를 브로드캐스트합니다
func (b *Broadcaster) BroadcastStatus(event StatusEvent) {
	b.Broadcast(event.TransportID, SSEEvent{
		Event: EventTypeStatus,
		Data:  event,
	})
}

// BroadcastError는 Error 이벤트를 브로드캐스트합니다
func (b *Broadcaster) BroadcastError(event ErrorEvent) {
	b.Broadcast(event.TransportID, SSEEvent{
		Event: EventTypeError,
		Data:  event,
	})
}

// BroadcastComplete는 Complete 이벤트를 브로드캐스트합니다
func (b *Broadcaster) BroadcastComplete(event CompleteEvent) {
	b.Broadcast(event.TransportID, SSEEvent{
		Event: EventTypeComplete,
		Data:  event,
	})
}

// ClientCount는 현재 연결된 클라이언트 수를 반환합니다
func (b *Broadcaster) ClientCount() int {
	return int(atomic.LoadInt32(&b.clientCount))
}

// ClientCountForTransport는 특정 Transport에 연결된 클라이언트 수를 반환합니다
func (b *Broadcaster) ClientCountForTransport(transportID string) int {
	count := 0
	b.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*Client); ok {
			if client.TransportID == transportID {
				count++
			}
		}
		return true
	})
	return count
}
