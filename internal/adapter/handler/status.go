// Package handler는 HTTP 요청 핸들러를 제공합니다
package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"oracle-etl/internal/adapter/sse"
)

// StatusHandler는 SSE 상태 스트리밍 핸들러입니다
type StatusHandler struct {
	broadcaster *sse.Broadcaster
}

// NewStatusHandler는 새로운 StatusHandler를 생성합니다
func NewStatusHandler(broadcaster *sse.Broadcaster) *StatusHandler {
	return &StatusHandler{
		broadcaster: broadcaster,
	}
}

// GetStatus는 Transport의 실시간 상태를 SSE로 스트리밍합니다
// GET /api/transports/:id/status
func (h *StatusHandler) GetStatus(c *fiber.Ctx) error {
	// fasthttp 버퍼 재사용 문제 방지를 위해 문자열 복사
	transportID := strings.Clone(c.Params("id"))

	// SSE 헤더 설정
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // nginx 프록시 버퍼링 비활성화

	// SSE 스트리밍 시작
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// 클라이언트 등록
		client := h.broadcaster.Register(transportID)
		defer h.broadcaster.Unregister(client.ID)

		// 초기 연결 확인 이벤트 전송
		initialEvent := sse.SSEEvent{
			Event: "connected",
			Data: map[string]string{
				"transport_id": transportID,
				"client_id":    client.ID,
				"message":      "SSE 연결이 설정되었습니다",
			},
		}
		h.writeSSEEvent(w, initialEvent)
		_ = w.Flush() // 초기 이벤트 전송 실패는 무시

		// 이벤트 스트리밍
		for {
			select {
			case event, ok := <-client.Events:
				if !ok {
					// 채널이 닫힘
					return
				}
				h.writeSSEEvent(w, event)
				if err := w.Flush(); err != nil {
					// 클라이언트 연결이 끊어짐
					return
				}

			case <-client.Done:
				// 클라이언트 종료 신호
				return
			}
		}
	})

	return nil
}

// writeSSEEvent는 SSE 형식으로 이벤트를 작성합니다
func (h *StatusHandler) writeSSEEvent(w *bufio.Writer, event sse.SSEEvent) {
	// event: {type}
	fmt.Fprintf(w, "event: %s\n", event.Event)

	// data: {json}
	data, err := json.Marshal(event.Data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", string(data))
}
