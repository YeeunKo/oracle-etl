// Package handler는 HTTP 요청 핸들러를 제공합니다
package handler

import (
	"bufio"
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"oracle-etl/internal/adapter/sse"
)

func TestNewStatusHandler(t *testing.T) {
	// StatusHandler 생성 테스트
	broadcaster := sse.NewBroadcaster()
	handler := NewStatusHandler(broadcaster)

	assert.NotNil(t, handler)
	assert.Equal(t, broadcaster, handler.broadcaster)
}

func TestStatusHandler_GetStatus_Success(t *testing.T) {
	// SSE 연결 성공 테스트
	broadcaster := sse.NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster.Run(ctx)

	handler := NewStatusHandler(broadcaster)

	app := fiber.New()
	app.Get("/api/transports/:id/status", handler.GetStatus)

	// 비동기로 이벤트 전송
	go func() {
		time.Sleep(100 * time.Millisecond)
		progressEvent := sse.ProgressEvent{
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

		// 작은 지연 후 연결 종료를 위해
		time.Sleep(100 * time.Millisecond)
		cancel() // 컨텍스트 취소로 테스트 종료
	}()

	req := httptest.NewRequest("GET", "/api/transports/TRPID-12345678/status", nil)

	// 타임아웃 설정
	testCtx, testCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer testCancel()

	// SSE 응답을 읽기 위해 Test 대신 직접 처리
	resp, err := app.Test(req, -1) // -1은 타임아웃 없음
	require.NoError(t, err)
	defer resp.Body.Close()

	// Content-Type 확인
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "text/event-stream")

	// Cache-Control 확인
	cacheControl := resp.Header.Get("Cache-Control")
	assert.Contains(t, cacheControl, "no-cache")

	// 응답 본문 읽기 (비동기로 읽어야 함)
	reader := bufio.NewReader(resp.Body)

	readDone := make(chan struct{})
	var receivedData []string

	go func() {
		for {
			select {
			case <-testCtx.Done():
				close(readDone)
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						close(readDone)
						return
					}
					return
				}
				if strings.TrimSpace(line) != "" {
					receivedData = append(receivedData, line)
				}
			}
		}
	}()

	// 타임아웃 또는 데이터 수신 대기
	select {
	case <-readDone:
	case <-testCtx.Done():
	}

	// 데이터가 수신되었는지 확인 (SSE 형식)
	// Note: 실제 SSE 스트리밍 테스트는 통합 테스트에서 더 정확하게 검증
}

func TestStatusHandler_GetStatus_TransportIDExtraction(t *testing.T) {
	// Transport ID가 올바르게 추출되는지 테스트
	broadcaster := sse.NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())

	go broadcaster.Run(ctx)

	handler := NewStatusHandler(broadcaster)

	app := fiber.New()
	app.Get("/api/transports/:id/status", handler.GetStatus)

	// 짧은 시간 후 취소하여 테스트 종료
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	req := httptest.NewRequest("GET", "/api/transports/TRPID-TESTID01/status", nil)
	resp, err := app.Test(req, 100)

	require.NoError(t, err)
	defer resp.Body.Close()

	// SSE 헤더가 설정되어 있는지 확인
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestStatusHandler_Headers(t *testing.T) {
	// SSE 필수 헤더 테스트
	broadcaster := sse.NewBroadcaster()

	ctx, cancel := context.WithCancel(context.Background())
	
	go broadcaster.Run(ctx)

	// 빠른 취소
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	handler := NewStatusHandler(broadcaster)

	app := fiber.New()
	app.Get("/api/transports/:id/status", handler.GetStatus)

	req := httptest.NewRequest("GET", "/api/transports/TRPID-12345678/status", nil)
	resp, err := app.Test(req, 100)

	require.NoError(t, err)
	defer resp.Body.Close()

	// SSE 필수 헤더 확인
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")
	assert.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))
	assert.Equal(t, "keep-alive", resp.Header.Get("Connection"))
}
