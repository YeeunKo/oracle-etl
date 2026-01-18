/**
 * Oracle 연결 상태 폴링 훅
 * 주기적으로 Oracle 연결 상태를 확인
 */

'use client';

import { useQuery } from '@tanstack/react-query';
import { getOracleStatus, getHealth } from '@/lib/api';
import type { OracleStatus, HealthResponse } from '@/types/api';

// ============================================================================
// 상수
// ============================================================================

/**
 * 기본 폴링 간격 (30초)
 */
const DEFAULT_POLL_INTERVAL = 30 * 1000;

/**
 * 에러 시 재시도 폴링 간격 (10초)
 */
const ERROR_POLL_INTERVAL = 10 * 1000;

// ============================================================================
// Query Keys
// ============================================================================

/**
 * 연결 상태 쿼리 키
 */
export const connectionStatusKeys = {
  all: ['connection'] as const,
  oracle: () => [...connectionStatusKeys.all, 'oracle'] as const,
  health: () => [...connectionStatusKeys.all, 'health'] as const,
};

// ============================================================================
// 훅: Oracle 연결 상태
// ============================================================================

/**
 * Oracle 연결 상태 훅 옵션
 */
interface UseOracleStatusOptions {
  /** 자동 갱신 활성화 (기본값: true) */
  enabled?: boolean;

  /** 폴링 간격 (밀리초, 기본값: 30000) */
  pollInterval?: number;
}

/**
 * Oracle 연결 상태 폴링 훅
 *
 * @param options - 훅 옵션
 * @returns Oracle 상태 쿼리 결과
 *
 * @example
 * const { data, isLoading, isError, error } = useOracleStatus();
 *
 * if (isLoading) return <Skeleton />;
 * if (isError) return <ErrorMessage error={error} />;
 *
 * return <StatusBadge connected={data.connected} />;
 */
export function useOracleStatus(options: UseOracleStatusOptions = {}) {
  const { enabled = true, pollInterval = DEFAULT_POLL_INTERVAL } = options;

  return useQuery<OracleStatus>({
    queryKey: connectionStatusKeys.oracle(),
    queryFn: getOracleStatus,
    enabled,
    refetchInterval: (query) => {
      // 에러 상태에서는 더 빠르게 재시도
      if (query.state.error) {
        return ERROR_POLL_INTERVAL;
      }
      return pollInterval;
    },
    // 윈도우 포커스 시 재조회
    refetchOnWindowFocus: true,
    // 마운트 시 재조회
    refetchOnMount: true,
    // stale time: 10초
    staleTime: 10 * 1000,
    // 재시도 설정
    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
  });
}

// ============================================================================
// 훅: 헬스 체크
// ============================================================================

/**
 * 헬스 체크 훅 옵션
 */
interface UseHealthOptions {
  /** 자동 갱신 활성화 (기본값: true) */
  enabled?: boolean;

  /** 폴링 간격 (밀리초, 기본값: 30000) */
  pollInterval?: number;
}

/**
 * 백엔드 헬스 체크 훅
 *
 * @param options - 훅 옵션
 * @returns 헬스 체크 쿼리 결과
 */
export function useHealth(options: UseHealthOptions = {}) {
  const { enabled = true, pollInterval = DEFAULT_POLL_INTERVAL } = options;

  return useQuery<HealthResponse>({
    queryKey: connectionStatusKeys.health(),
    queryFn: getHealth,
    enabled,
    refetchInterval: (query) => {
      if (query.state.error) {
        return ERROR_POLL_INTERVAL;
      }
      return pollInterval;
    },
    refetchOnWindowFocus: true,
    refetchOnMount: true,
    staleTime: 10 * 1000,
    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
  });
}

// ============================================================================
// 훅: 통합 연결 상태
// ============================================================================

/**
 * 통합 연결 상태 결과 타입
 */
interface ConnectionStatusResult {
  /** 전체 연결 상태 (헬스 + Oracle 모두 연결됨) */
  isConnected: boolean;

  /** 로딩 중 여부 */
  isLoading: boolean;

  /** 에러 발생 여부 */
  isError: boolean;

  /** 백엔드 헬스 상태 */
  health: HealthResponse | undefined;

  /** Oracle 연결 상태 */
  oracle: OracleStatus | undefined;

  /** 에러 메시지 */
  error: string | null;

  /** 마지막 확인 시간 */
  lastChecked: Date | null;
}

/**
 * 통합 연결 상태 훅
 * 백엔드 헬스와 Oracle 연결 상태를 모두 확인
 *
 * @param pollInterval - 폴링 간격 (기본값: 30초)
 * @returns 통합 연결 상태
 *
 * @example
 * const { isConnected, isLoading, oracle, health } = useConnectionStatus();
 *
 * return (
 *   <StatusIndicator
 *     status={isConnected ? 'connected' : 'disconnected'}
 *     loading={isLoading}
 *   />
 * );
 */
export function useConnectionStatus(
  pollInterval: number = DEFAULT_POLL_INTERVAL
): ConnectionStatusResult {
  const healthQuery = useHealth({ pollInterval });
  const oracleQuery = useOracleStatus({ pollInterval });

  // 로딩 상태
  const isLoading = healthQuery.isLoading || oracleQuery.isLoading;

  // 에러 상태
  const isError = healthQuery.isError || oracleQuery.isError;

  // 에러 메시지
  const error = healthQuery.error?.message || oracleQuery.error?.message || null;

  // 전체 연결 상태
  const isConnected =
    healthQuery.data?.status === 'ok' &&
    oracleQuery.data?.connected === true;

  // 마지막 확인 시간
  const lastChecked =
    healthQuery.dataUpdatedAt || oracleQuery.dataUpdatedAt
      ? new Date(Math.max(healthQuery.dataUpdatedAt, oracleQuery.dataUpdatedAt))
      : null;

  return {
    isConnected,
    isLoading,
    isError,
    health: healthQuery.data,
    oracle: oracleQuery.data,
    error,
    lastChecked,
  };
}
