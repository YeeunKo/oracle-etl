/**
 * Oracle ETL Pipeline API 클라이언트
 * Fetch 기반 타입 안전 API 호출
 */

import type {
  ApiError,
  CreateTransportRequest,
  ExecuteTransportResponse,
  HealthResponse,
  Job,
  JobListParams,
  OracleStatus,
  OracleTable,
  PaginatedResponse,
  Transport,
  TransportListParams,
} from '@/types/api';
import { useAuthStore } from '@/store/auth';

// ============================================================================
// API 설정
// ============================================================================

/**
 * API 기본 URL
 * Next.js 프록시를 통해 Go 백엔드로 라우팅됨
 */
const API_BASE_URL = '/api';

/**
 * 기본 요청 타임아웃 (밀리초)
 */
const DEFAULT_TIMEOUT = 30000;

// ============================================================================
// 에러 처리
// ============================================================================

/**
 * API 에러 클래스
 */
export class ApiClientError extends Error {
  public readonly status: number;
  public readonly code?: string;
  public readonly details?: Record<string, unknown>;

  constructor(message: string, status: number, code?: string, details?: Record<string, unknown>) {
    super(message);
    this.name = 'ApiClientError';
    this.status = status;
    this.code = code;
    this.details = details;
  }

  /**
   * API 에러 응답으로부터 ApiClientError 생성
   */
  static fromApiError(error: ApiError, status: number): ApiClientError {
    return new ApiClientError(error.message || error.error, status, error.code, error.details);
  }
}

// ============================================================================
// HTTP 클라이언트
// ============================================================================

/**
 * 요청 옵션 타입
 */
interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
  timeout?: number;
  params?: Record<string, string | number | boolean | undefined>;
}

/**
 * URL에 쿼리 파라미터 추가
 */
function buildUrl(path: string, params?: Record<string, string | number | boolean | undefined>): string {
  const url = new URL(path, window.location.origin);

  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        url.searchParams.append(key, String(value));
      }
    });
  }

  return url.pathname + url.search;
}

/**
 * 타임아웃이 적용된 fetch 래퍼
 */
async function fetchWithTimeout(
  url: string,
  options: RequestInit,
  timeout: number
): Promise<Response> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal,
    });
    return response;
  } finally {
    clearTimeout(timeoutId);
  }
}

/**
 * 공통 HTTP 요청 함수
 */
async function request<T>(
  method: string,
  path: string,
  options: RequestOptions = {}
): Promise<T> {
  const { body, timeout = DEFAULT_TIMEOUT, params, headers: customHeaders, ...restOptions } = options;

  // API 키 헤더 추가
  const apiKey = useAuthStore.getState().apiKey;
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...customHeaders as Record<string, string>,
  };

  if (apiKey) {
    headers['X-API-Key'] = apiKey;
  }

  const url = buildUrl(`${API_BASE_URL}${path}`, params);

  const requestOptions: RequestInit = {
    method,
    headers,
    ...restOptions,
  };

  if (body !== undefined) {
    requestOptions.body = JSON.stringify(body);
  }

  try {
    const response = await fetchWithTimeout(url, requestOptions, timeout);

    // 204 No Content 처리
    if (response.status === 204) {
      return undefined as T;
    }

    // 응답 본문 파싱
    const contentType = response.headers.get('content-type');
    let data: unknown;

    if (contentType?.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    // 에러 응답 처리
    if (!response.ok) {
      const apiError = data as ApiError;
      throw ApiClientError.fromApiError(
        apiError,
        response.status
      );
    }

    return data as T;
  } catch (error) {
    // AbortError 처리 (타임아웃)
    if (error instanceof Error && error.name === 'AbortError') {
      throw new ApiClientError('요청 시간이 초과되었습니다.', 408, 'TIMEOUT');
    }

    // 네트워크 에러 처리
    if (error instanceof TypeError && error.message === 'Failed to fetch') {
      throw new ApiClientError('서버에 연결할 수 없습니다.', 0, 'NETWORK_ERROR');
    }

    // ApiClientError는 그대로 전달
    if (error instanceof ApiClientError) {
      throw error;
    }

    // 기타 에러
    throw new ApiClientError(
      error instanceof Error ? error.message : '알 수 없는 에러가 발생했습니다.',
      500,
      'UNKNOWN_ERROR'
    );
  }
}

// ============================================================================
// API 함수들
// ============================================================================

/**
 * 헬스 체크 API
 */
export async function getHealth(): Promise<HealthResponse> {
  return request<HealthResponse>('GET', '/health');
}

/**
 * Oracle 연결 상태 조회 API
 */
export async function getOracleStatus(): Promise<OracleStatus> {
  return request<OracleStatus>('GET', '/oracle/status');
}

/**
 * 테이블 목록 조회 API
 */
export async function getTables(): Promise<OracleTable[]> {
  return request<OracleTable[]>('GET', '/tables');
}

// ============================================================================
// Transport API
// ============================================================================

/**
 * Transport 목록 조회
 */
export async function getTransports(
  params?: TransportListParams
): Promise<PaginatedResponse<Transport>> {
  return request<PaginatedResponse<Transport>>('GET', '/transports', {
    params: params ? {
      offset: params.offset,
      limit: params.limit,
      search: params.search,
    } : undefined,
  });
}

/**
 * Transport 상세 조회
 */
export async function getTransport(id: string): Promise<Transport> {
  return request<Transport>('GET', `/transports/${id}`);
}

/**
 * Transport 생성
 */
export async function createTransport(
  data: CreateTransportRequest
): Promise<Transport> {
  return request<Transport>('POST', '/transports', { body: data });
}

/**
 * Transport 삭제
 */
export async function deleteTransport(id: string): Promise<void> {
  return request<void>('DELETE', `/transports/${id}`);
}

/**
 * Transport 실행
 */
export async function executeTransport(id: string): Promise<ExecuteTransportResponse> {
  return request<ExecuteTransportResponse>('POST', `/transports/${id}/execute`);
}

// ============================================================================
// Job API
// ============================================================================

/**
 * Job 목록 조회
 */
export async function getJobs(params?: JobListParams): Promise<PaginatedResponse<Job>> {
  return request<PaginatedResponse<Job>>('GET', '/jobs', {
    params: params ? {
      transport_id: params.transportId,
      status: params.status,
      offset: params.offset,
      limit: params.limit,
    } : undefined,
  });
}

/**
 * Job 상세 조회
 */
export async function getJob(id: string): Promise<Job> {
  return request<Job>('GET', `/jobs/${id}`);
}

/**
 * Job 취소
 */
export async function cancelJob(id: string): Promise<Job> {
  return request<Job>('POST', `/jobs/${id}/cancel`);
}

/**
 * Job 재시도 (Transport 재실행)
 */
export async function retryJob(id: string): Promise<ExecuteTransportResponse> {
  return request<ExecuteTransportResponse>('POST', `/jobs/${id}/retry`);
}

// ============================================================================
// API 클라이언트 객체 (옵션)
// ============================================================================

/**
 * API 클라이언트 객체
 * 네임스페이스로 그룹화된 API 함수들
 */
export const api = {
  health: {
    get: getHealth,
  },
  oracle: {
    getStatus: getOracleStatus,
  },
  tables: {
    list: getTables,
  },
  transports: {
    list: getTransports,
    get: getTransport,
    create: createTransport,
    delete: deleteTransport,
    execute: executeTransport,
  },
  jobs: {
    list: getJobs,
    get: getJob,
    cancel: cancelJob,
    retry: retryJob,
  },
} as const;
