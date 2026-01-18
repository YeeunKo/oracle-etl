/**
 * Job 실시간 모니터링을 위한 SSE 훅
 * Server-Sent Events를 통해 Job 상태를 실시간으로 업데이트
 */

'use client';

import { useEffect, useCallback, useState, useRef } from 'react';
import type { SSEEvent, JobProgress, JobStatus } from '@/types/api';
import { useAuthStore } from '@/store/auth';

// ============================================================================
// 타입 정의
// ============================================================================

/**
 * SSE 로그 항목
 */
export interface SSELogEntry {
  id: string;
  level: 'info' | 'warn' | 'error';
  message: string;
  timestamp: string;
}

/**
 * SSE 연결 상태
 */
export type SSEConnectionState = 'connecting' | 'connected' | 'disconnected' | 'error';

/**
 * useJobSSE 훅 옵션
 */
export interface UseJobSSEOptions {
  /** Job 완료 시 콜백 */
  onComplete?: (summary: { totalRows: number; totalBytes: number; duration: number }) => void;
  /** 에러 발생 시 콜백 */
  onError?: (error: string) => void;
  /** 자동 연결 여부 (기본값: true) */
  autoConnect?: boolean;
}

/**
 * useJobSSE 훅 반환 타입
 */
export interface UseJobSSEResult {
  /** 현재 Job 상태 */
  status: JobStatus | null;
  /** 진행 상황 */
  progress: JobProgress | null;
  /** 로그 목록 */
  logs: SSELogEntry[];
  /** SSE 연결 상태 */
  connectionState: SSEConnectionState;
  /** 에러 메시지 */
  error: string | null;
  /** 수동 연결 함수 */
  connect: () => void;
  /** 연결 해제 함수 */
  disconnect: () => void;
  /** 로그 초기화 함수 */
  clearLogs: () => void;
}

// ============================================================================
// 상수
// ============================================================================

const SSE_BASE_URL = '/api';
const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_DELAY = 3000;

// ============================================================================
// useJobSSE 훅
// ============================================================================

/**
 * Job 실시간 모니터링 훅
 *
 * @param jobId - 모니터링할 Job ID
 * @param options - 훅 옵션
 * @returns SSE 상태 및 제어 함수
 *
 * @example
 * const { status, progress, logs, connectionState } = useJobSSE('job-123', {
 *   onComplete: (summary) => console.log('완료:', summary),
 *   onError: (error) => console.error('에러:', error),
 * });
 */
export function useJobSSE(
  jobId: string | null,
  options: UseJobSSEOptions = {}
): UseJobSSEResult {
  const { onComplete, onError, autoConnect = true } = options;

  // 상태
  const [status, setStatus] = useState<JobStatus | null>(null);
  const [progress, setProgress] = useState<JobProgress | null>(null);
  const [logs, setLogs] = useState<SSELogEntry[]>([]);
  const [connectionState, setConnectionState] = useState<SSEConnectionState>('disconnected');
  const [error, setError] = useState<string | null>(null);

  // Refs
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  /**
   * 로그 항목 추가
   */
  const addLog = useCallback((entry: Omit<SSELogEntry, 'id'>) => {
    const newLog: SSELogEntry = {
      ...entry,
      id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
    };
    setLogs((prev) => [...prev, newLog]);
  }, []);

  /**
   * 로그 초기화
   */
  const clearLogs = useCallback(() => {
    setLogs([]);
  }, []);

  /**
   * SSE 연결 해제
   */
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    setConnectionState('disconnected');
    reconnectAttemptsRef.current = 0;
  }, []);

  /**
   * SSE 연결
   */
  const connect = useCallback(() => {
    if (!jobId) return;

    // 기존 연결 정리
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    setConnectionState('connecting');
    setError(null);

    // API 키 포함한 URL 생성
    const apiKey = useAuthStore.getState().apiKey;
    const url = new URL(`${SSE_BASE_URL}/jobs/${jobId}/stream`, window.location.origin);
    if (apiKey) {
      url.searchParams.set('api_key', apiKey);
    }

    try {
      const eventSource = new EventSource(url.toString());
      eventSourceRef.current = eventSource;

      // 연결 성공
      eventSource.onopen = () => {
        setConnectionState('connected');
        reconnectAttemptsRef.current = 0;
        addLog({
          level: 'info',
          message: 'SSE 연결이 설정되었습니다.',
          timestamp: new Date().toISOString(),
        });
      };

      // 메시지 수신
      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as SSEEvent;

          switch (data.type) {
            case 'status':
              setStatus(data.status);
              setProgress(data.progress);
              break;

            case 'log':
              addLog({
                level: data.level,
                message: data.message,
                timestamp: data.timestamp,
              });
              break;

            case 'complete':
              setStatus(data.status);
              addLog({
                level: 'info',
                message: `작업 완료: ${data.summary.totalRows.toLocaleString()} 행, ${(data.summary.totalBytes / 1024 / 1024).toFixed(2)} MB`,
                timestamp: data.timestamp,
              });
              onComplete?.(data.summary);
              disconnect();
              break;

            case 'error':
              setError(data.error);
              addLog({
                level: 'error',
                message: data.error,
                timestamp: data.timestamp,
              });
              onError?.(data.error);
              disconnect();
              break;
          }
        } catch (parseError) {
          console.error('[SSE] 메시지 파싱 실패:', parseError);
        }
      };

      // 에러 처리
      eventSource.onerror = () => {
        setConnectionState('error');
        eventSource.close();
        eventSourceRef.current = null;

        // 재연결 시도
        if (reconnectAttemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
          reconnectAttemptsRef.current++;
          const delay = RECONNECT_DELAY * reconnectAttemptsRef.current;

          addLog({
            level: 'warn',
            message: `연결 끊김. ${delay / 1000}초 후 재연결 시도... (${reconnectAttemptsRef.current}/${MAX_RECONNECT_ATTEMPTS})`,
            timestamp: new Date().toISOString(),
          });

          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, delay);
        } else {
          setError('최대 재연결 시도 횟수를 초과했습니다.');
          addLog({
            level: 'error',
            message: '최대 재연결 시도 횟수를 초과했습니다.',
            timestamp: new Date().toISOString(),
          });
          onError?.('최대 재연결 시도 횟수를 초과했습니다.');
        }
      };
    } catch (err) {
      console.error('[SSE] 연결 생성 실패:', err);
      setConnectionState('error');
      setError('SSE 연결 생성에 실패했습니다.');
    }
  }, [jobId, addLog, disconnect, onComplete, onError]);

  // 자동 연결
  useEffect(() => {
    if (autoConnect && jobId) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [jobId, autoConnect, connect, disconnect]);

  return {
    status,
    progress,
    logs,
    connectionState,
    error,
    connect,
    disconnect,
    clearLogs,
  };
}
