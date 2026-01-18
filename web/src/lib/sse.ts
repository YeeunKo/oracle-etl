/**
 * Server-Sent Events (SSE) 유틸리티
 * Transport 실행 상태를 실시간으로 수신
 */

import type { SSEEvent } from '@/types/api';
import { useAuthStore } from '@/store/auth';

// ============================================================================
// SSE 설정
// ============================================================================

/**
 * SSE 기본 URL
 */
const SSE_BASE_URL = '/api';

/**
 * 기본 재연결 지연 시간 (밀리초)
 */
const DEFAULT_RECONNECT_DELAY = 3000;

/**
 * 최대 재연결 지연 시간 (밀리초)
 */
const MAX_RECONNECT_DELAY = 30000;

/**
 * 최대 재연결 시도 횟수
 */
const MAX_RECONNECT_ATTEMPTS = 10;

// ============================================================================
// SSE 클라이언트 타입
// ============================================================================

/**
 * SSE 이벤트 핸들러 타입
 */
export interface SSEEventHandlers {
  onMessage?: (event: SSEEvent) => void;
  onError?: (error: Error) => void;
  onOpen?: () => void;
  onClose?: () => void;
}

/**
 * SSE 클라이언트 옵션
 */
export interface SSEClientOptions {
  autoReconnect?: boolean;
  reconnectDelay?: number;
  maxReconnectAttempts?: number;
}

/**
 * SSE 클라이언트 인터페이스
 */
export interface SSEClient {
  connect: () => void;
  disconnect: () => void;
  isConnected: () => boolean;
}

// ============================================================================
// SSE 클라이언트 구현
// ============================================================================

/**
 * SSE 클라이언트 생성 함수
 *
 * @param transportId - Transport ID
 * @param handlers - 이벤트 핸들러
 * @param options - 클라이언트 옵션
 * @returns SSE 클라이언트 인스턴스
 *
 * @example
 * const client = createSSEClient('transport-123', {
 *   onMessage: (event) => console.log('Received:', event),
 *   onError: (error) => console.error('Error:', error),
 * });
 *
 * client.connect();
 * // ... 나중에
 * client.disconnect();
 */
export function createSSEClient(
  transportId: string,
  handlers: SSEEventHandlers,
  options: SSEClientOptions = {}
): SSEClient {
  const {
    autoReconnect = true,
    reconnectDelay = DEFAULT_RECONNECT_DELAY,
    maxReconnectAttempts = MAX_RECONNECT_ATTEMPTS,
  } = options;

  let eventSource: EventSource | null = null;
  let reconnectAttempts = 0;
  let reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
  let isManuallyDisconnected = false;

  /**
   * 재연결 지연 시간 계산 (지수 백오프)
   */
  function getReconnectDelay(): number {
    const delay = Math.min(
      reconnectDelay * Math.pow(2, reconnectAttempts),
      MAX_RECONNECT_DELAY
    );
    return delay;
  }

  /**
   * SSE 연결
   */
  function connect(): void {
    // 이미 연결되어 있으면 무시
    if (eventSource?.readyState === EventSource.OPEN) {
      return;
    }

    isManuallyDisconnected = false;

    // API 키 포함한 URL 생성
    const apiKey = useAuthStore.getState().apiKey;
    const url = new URL(`${SSE_BASE_URL}/transports/${transportId}/status`, window.location.origin);
    if (apiKey) {
      url.searchParams.set('api_key', apiKey);
    }

    try {
      eventSource = new EventSource(url.toString());

      // 연결 성공
      eventSource.onopen = () => {
        reconnectAttempts = 0;
        handlers.onOpen?.();
      };

      // 메시지 수신
      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as SSEEvent;
          handlers.onMessage?.(data);

          // 완료 또는 에러 이벤트 시 연결 종료
          if (data.type === 'complete' || data.type === 'error') {
            disconnect();
          }
        } catch (parseError) {
          console.error('[SSE] 메시지 파싱 실패:', parseError);
        }
      };

      // 에러 처리
      eventSource.onerror = (event) => {
        console.error('[SSE] 연결 에러:', event);

        const error = new Error('SSE 연결 에러가 발생했습니다.');
        handlers.onError?.(error);

        // 연결 종료
        eventSource?.close();
        eventSource = null;

        // 자동 재연결
        if (autoReconnect && !isManuallyDisconnected && reconnectAttempts < maxReconnectAttempts) {
          const delay = getReconnectDelay();
          reconnectAttempts++;

          console.log(`[SSE] ${delay}ms 후 재연결 시도 (${reconnectAttempts}/${maxReconnectAttempts})`);

          reconnectTimeoutId = setTimeout(() => {
            connect();
          }, delay);
        } else if (reconnectAttempts >= maxReconnectAttempts) {
          const maxAttemptsError = new Error('SSE 최대 재연결 시도 횟수를 초과했습니다.');
          handlers.onError?.(maxAttemptsError);
          handlers.onClose?.();
        }
      };
    } catch (error) {
      console.error('[SSE] 연결 생성 실패:', error);
      handlers.onError?.(error instanceof Error ? error : new Error('SSE 연결 생성 실패'));
    }
  }

  /**
   * SSE 연결 종료
   */
  function disconnect(): void {
    isManuallyDisconnected = true;

    // 재연결 타이머 취소
    if (reconnectTimeoutId) {
      clearTimeout(reconnectTimeoutId);
      reconnectTimeoutId = null;
    }

    // EventSource 종료
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }

    handlers.onClose?.();
  }

  /**
   * 연결 상태 확인
   */
  function isConnected(): boolean {
    return eventSource?.readyState === EventSource.OPEN;
  }

  return {
    connect,
    disconnect,
    isConnected,
  };
}

// ============================================================================
// React Hook을 위한 SSE 헬퍼
// ============================================================================

/**
 * SSE 구독 헬퍼 함수
 * React 컴포넌트에서 사용하기 편리한 형태로 래핑
 *
 * @param transportId - Transport ID
 * @param onEvent - 이벤트 콜백
 * @returns 구독 해제 함수
 */
export function subscribeToTransportStatus(
  transportId: string,
  onEvent: (event: SSEEvent) => void,
  onError?: (error: Error) => void
): () => void {
  const client = createSSEClient(transportId, {
    onMessage: onEvent,
    onError,
  });

  client.connect();

  return () => {
    client.disconnect();
  };
}
