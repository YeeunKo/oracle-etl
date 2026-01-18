/**
 * Oracle ETL Pipeline API 타입 정의
 * Go 백엔드 API와의 타입 동기화
 */

// ============================================================================
// 공통 타입
// ============================================================================

/**
 * API 응답 래퍼 타입
 */
export interface ApiResponse<T> {
  data: T;
  message?: string;
}

/**
 * 페이지네이션 응답 타입
 */
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  offset: number;
  limit: number;
}

/**
 * API 에러 응답 타입
 */
export interface ApiError {
  error: string;
  message: string;
  code?: string;
  details?: Record<string, unknown>;
}

// ============================================================================
// Oracle 연결 상태 타입
// ============================================================================

/**
 * Oracle 연결 상태
 */
export interface OracleStatus {
  connected: boolean;
  version?: string;
  host?: string;
  database?: string;
  lastCheck: string;
  error?: string;
}

/**
 * 헬스 체크 응답
 */
export interface HealthResponse {
  status: 'ok' | 'error';
  timestamp: string;
  version?: string;
}

// ============================================================================
// 테이블 관련 타입
// ============================================================================

/**
 * Oracle 테이블 정보
 */
export interface OracleTable {
  owner: string;
  tableName: string;
  numRows?: number;
  lastAnalyzed?: string;
  tablespaceName?: string;
}

/**
 * 테이블 컬럼 정보
 */
export interface TableColumn {
  columnName: string;
  dataType: string;
  dataLength: number;
  dataPrecision?: number;
  dataScale?: number;
  nullable: boolean;
  defaultValue?: string;
}

// ============================================================================
// Transport 관련 타입
// ============================================================================

/**
 * Transport 상태 열거형
 */
export type TransportStatus =
  | 'idle'        // 대기 중
  | 'running'     // 실행 중
  | 'completed'   // 완료
  | 'failed'      // 실패
  | 'cancelled';  // 취소됨

/**
 * Transport 설정
 */
export interface TransportConfig {
  batchSize: number;
  parallelism: number;
  retryAttempts: number;
  retryDelay: number;
  timeout: number;
}

/**
 * Transport 엔티티
 */
export interface Transport {
  id: string;
  name: string;
  description?: string;
  sourceTable: string;
  sourceSchema: string;
  targetPath: string;
  targetFormat: 'csv' | 'json' | 'parquet';
  config: TransportConfig;
  status: TransportStatus;
  createdAt: string;
  updatedAt: string;
  lastRunAt?: string;
}

/**
 * Transport 생성 요청
 */
export interface CreateTransportRequest {
  name: string;
  description?: string;
  sourceTable: string;
  sourceSchema: string;
  targetPath: string;
  targetFormat: 'csv' | 'json' | 'parquet';
  config?: Partial<TransportConfig>;
}

/**
 * Transport 실행 응답
 */
export interface ExecuteTransportResponse {
  jobId: string;
  transportId: string;
  status: 'started';
  message: string;
}

// ============================================================================
// Job 관련 타입
// ============================================================================

/**
 * Job 상태 열거형
 */
export type JobStatus =
  | 'pending'     // 대기 중
  | 'running'     // 실행 중
  | 'completed'   // 완료
  | 'failed'      // 실패
  | 'cancelled';  // 취소됨

/**
 * Extraction (추출 단계) 정보
 */
export interface Extraction {
  id: string;
  jobId: string;
  tableName: string;
  startedAt: string;
  completedAt?: string;
  rowsExtracted: number;
  bytesExtracted: number;
  status: JobStatus;
  error?: string;
}

/**
 * Job 진행 상황
 */
export interface JobProgress {
  phase: 'initializing' | 'extracting' | 'transforming' | 'loading' | 'completed';
  percentage: number;
  currentTable?: string;
  rowsProcessed: number;
  totalRows: number;
  bytesProcessed: number;
  estimatedTimeRemaining?: number;
}

/**
 * Job 엔티티
 */
export interface Job {
  id: string;
  transportId: string;
  transportName: string;
  status: JobStatus;
  progress: JobProgress;
  extractions: Extraction[];
  startedAt: string;
  completedAt?: string;
  error?: string;
  logs?: string[];
}

/**
 * Job 목록 조회 파라미터
 */
export interface JobListParams {
  transportId?: string;
  status?: JobStatus;
  offset?: number;
  limit?: number;
}

// ============================================================================
// SSE 이벤트 타입
// ============================================================================

/**
 * SSE 상태 업데이트 이벤트
 */
export interface SSEStatusEvent {
  type: 'status';
  jobId: string;
  status: JobStatus;
  progress: JobProgress;
  timestamp: string;
}

/**
 * SSE 로그 이벤트
 */
export interface SSELogEvent {
  type: 'log';
  jobId: string;
  level: 'info' | 'warn' | 'error';
  message: string;
  timestamp: string;
}

/**
 * SSE 완료 이벤트
 */
export interface SSECompleteEvent {
  type: 'complete';
  jobId: string;
  status: JobStatus;
  summary: {
    totalRows: number;
    totalBytes: number;
    duration: number;
  };
  timestamp: string;
}

/**
 * SSE 에러 이벤트
 */
export interface SSEErrorEvent {
  type: 'error';
  jobId: string;
  error: string;
  timestamp: string;
}

/**
 * SSE 이벤트 유니온 타입
 */
export type SSEEvent =
  | SSEStatusEvent
  | SSELogEvent
  | SSECompleteEvent
  | SSEErrorEvent;

// ============================================================================
// API 요청 타입
// ============================================================================

/**
 * Transport 목록 조회 파라미터
 */
export interface TransportListParams {
  offset?: number;
  limit?: number;
  search?: string;
}
