# SPEC-ETL-001: Implementation Plan

## Metadata

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-ETL-001 |
| Document Type | Implementation Plan |
| Created | 2026-01-18 |
| Status | Planned |

---

## Implementation Strategy

### Development Approach

본 프로젝트는 **TDD (Test-Driven Development)** 접근법을 채택하여 개발합니다:

1. **RED**: 각 기능에 대한 테스트 코드 작성 (실패 상태)
2. **GREEN**: 테스트를 통과하는 최소한의 코드 구현
3. **REFACTOR**: 코드 품질 개선 및 최적화

### Architecture Pattern

**Clean Architecture** 패턴을 적용하여 관심사 분리:

```
cmd/
├── server/
│   └── main.go              # Application entry point

internal/
├── domain/                   # Business entities
│   ├── transport.go
│   ├── job.go
│   └── extraction.go
├── usecase/                  # Business logic
│   ├── transport_service.go
│   ├── extraction_service.go
│   └── job_service.go
├── adapter/
│   ├── handler/              # HTTP handlers (Fiber)
│   │   ├── health.go
│   │   ├── oracle.go
│   │   ├── table.go
│   │   ├── transport.go
│   │   └── job.go
│   ├── oracle/               # Oracle repository
│   │   ├── connection.go
│   │   ├── pool.go
│   │   └── extractor.go
│   ├── gcs/                  # GCS repository
│   │   ├── client.go
│   │   └── uploader.go
│   └── sse/                  # SSE broadcaster
│       └── broadcaster.go
├── config/                   # Configuration
│   └── config.go
└── middleware/               # HTTP middleware
    ├── auth.go
    ├── logging.go
    └── recovery.go

pkg/
├── jsonl/                    # JSONL encoder
│   └── encoder.go
└── compress/                 # Gzip compression
    └── gzip.go
```

---

## Milestones

### Milestone 1: Foundation (Priority: High)

**목표**: 프로젝트 기반 구조 및 핵심 인프라 구축

#### Tasks

| ID | Task | Description | Dependencies |
|----|------|-------------|--------------|
| M1-01 | 프로젝트 초기화 | Go module, 디렉토리 구조, 의존성 설정 | - |
| M1-02 | 설정 관리 | 환경 변수 및 config.yaml 파싱 | M1-01 |
| M1-03 | 로깅 시스템 | 구조화된 JSON 로깅 (zerolog) | M1-01 |
| M1-04 | Fiber 서버 설정 | HTTP 서버, 미들웨어, 라우팅 | M1-02 |
| M1-05 | Health 엔드포인트 | GET /api/health 구현 | M1-04 |

#### Deliverables
- 실행 가능한 빈 서버
- 설정 로딩 시스템
- 구조화된 로깅
- Health check API

---

### Milestone 2: Oracle Integration (Priority: High)

**목표**: Oracle DB 연결 및 데이터 추출 기능 구현

#### Tasks

| ID | Task | Description | Dependencies |
|----|------|-------------|--------------|
| M2-01 | Oracle 커넥션 풀 | godror 기반 커넥션 풀 구현 | M1-02 |
| M2-02 | mTLS 연결 설정 | wallet 기반 보안 연결 | M2-01 |
| M2-03 | 연결 상태 체크 | GET /api/oracle/status 구현 | M2-02 |
| M2-04 | 테이블 목록 조회 | GET /api/tables 구현 (row count 포함) | M2-02 |
| M2-05 | 샘플 데이터 조회 | GET /api/tables/:name/sample 구현 | M2-04 |
| M2-06 | 배치 데이터 추출 | FetchArraySize=1000 배치 페치 | M2-04 |
| M2-07 | 청크 스트리밍 | 10,000 rows 청크 스트리밍 | M2-06 |

#### Deliverables
- Oracle 커넥션 풀 관리
- mTLS 보안 연결
- 테이블 메타데이터 조회
- 고성능 데이터 추출

#### Technical Approach

```go
// Oracle Connection Pool Configuration
type OracleConfig struct {
    WalletPath     string
    TNSName        string
    Username       string
    Password       string
    PoolMinConns   int
    PoolMaxConns   int
    FetchArraySize int
}

// Connection Pool Implementation
func NewOraclePool(cfg OracleConfig) (*sql.DB, error) {
    connStr := fmt.Sprintf(`user="%s" password="%s" connectString="%s"
        poolMinSessions=%d poolMaxSessions=%d
        libDir="/opt/oracle/instantclient"`,
        cfg.Username, cfg.Password, cfg.TNSName,
        cfg.PoolMinConns, cfg.PoolMaxConns)

    db, err := sql.Open("godror", connStr)
    if err != nil {
        return nil, err
    }

    // Set FetchArraySize via context
    ctx := godror.ContextWithParams(context.Background(),
        godror.FetchArraySize(cfg.FetchArraySize))

    return db, nil
}
```

---

### Milestone 3: GCS Upload (Priority: High)

**목표**: Google Cloud Storage 스트리밍 업로드 구현

#### Tasks

| ID | Task | Description | Dependencies |
|----|------|-------------|--------------|
| M3-01 | GCS 클라이언트 설정 | 서비스 계정 인증 | M1-02 |
| M3-02 | JSONL 인코더 | JSON Lines 직렬화 | M1-01 |
| M3-03 | Gzip 압축 스트림 | 압축 스트림 구현 | M3-02 |
| M3-04 | 스트리밍 업로드 | 청크 기반 스트리밍 업로드 | M3-01, M3-03 |
| M3-05 | Resumable 업로드 | 재개 가능한 업로드 구현 | M3-04 |
| M3-06 | 업로드 진행률 추적 | 바이트 전송량 메트릭 | M3-04 |

#### Deliverables
- GCS 클라이언트 래퍼
- JSONL + Gzip 스트리밍 파이프라인
- Resumable upload 지원
- 업로드 메트릭 수집

#### Technical Approach

```go
// Streaming Pipeline: Oracle -> JSONL -> Gzip -> GCS
type StreamPipeline struct {
    reader     *sql.Rows
    jsonlEnc   *jsonl.Encoder
    gzipWriter *gzip.Writer
    gcsWriter  *storage.Writer
}

func (p *StreamPipeline) Stream(ctx context.Context) error {
    // Pipe: rows -> JSONL -> gzip -> GCS
    pr, pw := io.Pipe()

    gzWriter := gzip.NewWriter(pw)
    jsonlEnc := jsonl.NewEncoder(gzWriter)

    go func() {
        defer pw.Close()
        defer gzWriter.Close()

        for p.reader.Next() {
            row := p.scanRow()
            jsonlEnc.Encode(row)
        }
    }()

    // Stream to GCS
    _, err := io.Copy(p.gcsWriter, pr)
    return err
}
```

---

### Milestone 4: Transport Management (Priority: High)

**목표**: Transport 및 Job 관리 시스템 구현

#### Tasks

| ID | Task | Description | Dependencies |
|----|------|-------------|--------------|
| M4-01 | Transport 도메인 모델 | Transport, Job, Extraction 엔티티 | M1-01 |
| M4-02 | In-Memory 저장소 | 초기 개발용 메모리 저장소 | M4-01 |
| M4-03 | Transport CRUD | POST/GET /api/transport 구현 | M4-02 |
| M4-04 | Job 관리 | Job 생성, 상태 관리, 버전닝 | M4-02 |
| M4-05 | Transport 실행 | POST /api/transport/:id/execute | M2-07, M3-05, M4-04 |
| M4-06 | Job 이력 조회 | GET /api/jobs 구현 | M4-04 |

#### Deliverables
- Transport CRUD API
- Job 버전 관리
- Transport 실행 오케스트레이션
- Job 이력 조회

---

### Milestone 5: Real-time Monitoring (Priority: Medium)

**목표**: SSE 기반 실시간 진행률 모니터링

#### Tasks

| ID | Task | Description | Dependencies |
|----|------|-------------|--------------|
| M5-01 | SSE Broadcaster | SSE 이벤트 브로드캐스터 | M1-04 |
| M5-02 | Progress 이벤트 | 진행률 이벤트 발행 | M5-01 |
| M5-03 | Status 엔드포인트 | GET /api/transport/:id/status (SSE) | M5-01, M4-05 |
| M5-04 | 메트릭 수집 | rows/sec, bytes 등 메트릭 계산 | M5-02 |
| M5-05 | Error 이벤트 | 오류 이벤트 발행 및 처리 | M5-02 |

#### Deliverables
- SSE 실시간 스트리밍
- 진행률 메트릭 (rows/sec, bytes)
- 에러 알림 시스템

#### Technical Approach

```go
// SSE Event Types
type SSEEvent struct {
    Event string      `json:"-"`
    Data  interface{} `json:"data"`
}

type ProgressEvent struct {
    Table          string  `json:"table"`
    RowsProcessed  int64   `json:"rows_processed"`
    RowsPerSecond  float64 `json:"rows_per_second"`
    ProgressPercent float64 `json:"progress_percent"`
}

// SSE Broadcaster
type Broadcaster struct {
    clients    map[string]chan SSEEvent
    register   chan Client
    unregister chan string
    events     chan SSEEvent
}
```

---

### Milestone 6: Security & Auth (Priority: Medium)

**목표**: API 인증 및 보안 강화

#### Tasks

| ID | Task | Description | Dependencies |
|----|------|-------------|--------------|
| M6-01 | API Key 인증 | X-API-Key 헤더 인증 | M1-04 |
| M6-02 | Bearer Token | Authorization 헤더 지원 | M6-01 |
| M6-03 | Rate Limiting | 요청 제한 미들웨어 | M1-04 |
| M6-04 | 민감 데이터 마스킹 | 로그에서 자격증명 제거 | M1-03 |
| M6-05 | CORS 설정 | Cross-Origin 요청 허용 설정 | M1-04 |

#### Deliverables
- API Key/Bearer 인증
- Rate limiting
- 보안 로깅

---

### Milestone 7: Error Handling & Resilience (Priority: Medium)

**목표**: 견고한 에러 처리 및 복구 메커니즘

#### Tasks

| ID | Task | Description | Dependencies |
|----|------|-------------|--------------|
| M7-01 | 구조화된 에러 | 표준 에러 응답 포맷 | M1-04 |
| M7-02 | Retry 로직 | Exponential backoff 재시도 | M2-07, M3-04 |
| M7-03 | Circuit Breaker | Oracle 연결 circuit breaker | M2-01 |
| M7-04 | Graceful Shutdown | 시그널 핸들링 및 정리 | M1-04 |
| M7-05 | Panic Recovery | 패닉 복구 미들웨어 | M1-04 |

#### Deliverables
- 통일된 에러 응답
- 자동 재시도 메커니즘
- 안전한 서버 종료

---

### Milestone 8: Performance Optimization (Priority: Medium)

**목표**: 100,000 rows/second 목표 달성

#### Tasks

| ID | Task | Description | Dependencies |
|----|------|-------------|--------------|
| M8-01 | 병렬 테이블 처리 | 다중 테이블 동시 추출 | M4-05 |
| M8-02 | 버퍼 최적화 | IO 버퍼 크기 튜닝 | M3-04 |
| M8-03 | 메모리 프로파일링 | 메모리 사용량 분석 | M4-05 |
| M8-04 | 벤치마크 테스트 | 성능 측정 및 병목 식별 | M4-05 |
| M8-05 | 최적화 적용 | 벤치마크 기반 최적화 | M8-04 |

#### Deliverables
- 병렬 처리 구현
- 성능 벤치마크 결과
- 최적화 문서

---

### Milestone 9: Testing & Quality (Priority: High)

**목표**: 85%+ 테스트 커버리지 및 품질 보증

#### Tasks

| ID | Task | Description | Dependencies |
|----|------|-------------|--------------|
| M9-01 | Unit Tests | 각 패키지 단위 테스트 | All |
| M9-02 | Integration Tests | Oracle/GCS 통합 테스트 | M2-07, M3-05 |
| M9-03 | E2E Tests | 전체 ETL 플로우 테스트 | M4-05 |
| M9-04 | Performance Tests | 성능 목표 검증 테스트 | M8-04 |
| M9-05 | Linting & Format | golangci-lint, gofmt | M1-01 |
| M9-06 | Security Scan | gosec 보안 스캔 | All |

#### Deliverables
- 85%+ 테스트 커버리지
- CI/CD 파이프라인
- 품질 리포트

---

## Technical Approach

### Key Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| gofiber/fiber | v2.52+ | Web framework |
| godror/godror | v0.44+ | Oracle driver |
| cloud.google.com/go/storage | v1.40+ | GCS SDK |
| rs/zerolog | v1.32+ | Structured logging |
| spf13/viper | v1.18+ | Configuration |
| stretchr/testify | v1.9+ | Testing |

### Performance Strategies

1. **Batch Fetching**: FetchArraySize=1000으로 라운드트립 최소화
2. **Streaming Pipeline**: 메모리 복사 없는 파이프라인 처리
3. **Parallel Processing**: 테이블별 goroutine 병렬 처리
4. **Buffer Optimization**: 64KB 버퍼로 IO 효율화
5. **Connection Pooling**: 커넥션 재사용으로 오버헤드 제거

### Error Handling Strategy

```go
// Structured Error Response
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
    TraceID string `json:"trace_id"`
}

// Error Codes
const (
    ErrCodeOracleConnection = "ORACLE_CONNECTION_ERROR"
    ErrCodeGCSUpload        = "GCS_UPLOAD_ERROR"
    ErrCodeTransportNotFound = "TRANSPORT_NOT_FOUND"
    ErrCodeValidation       = "VALIDATION_ERROR"
    ErrCodeAuth             = "AUTHENTICATION_ERROR"
)
```

---

## Risk Assessment

### Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Oracle 연결 불안정 | Medium | High | Circuit breaker, 연결 풀 모니터링 |
| 메모리 부족 (대용량 테이블) | Medium | High | 청크 스트리밍, 메모리 프로파일링 |
| GCS 업로드 실패 | Low | Medium | Resumable upload, 재시도 로직 |
| 성능 목표 미달 | Medium | High | 병렬 처리, 버퍼 최적화, 벤치마크 |

### Mitigation Strategies

1. **연결 불안정**: Connection pool health check, circuit breaker 패턴
2. **메모리 부족**: 청크 기반 스트리밍, 압축 전송, pprof 모니터링
3. **업로드 실패**: Resumable upload API, exponential backoff
4. **성능 미달**: 병렬화 확대, 버퍼 튜닝, 프로파일링 기반 최적화

---

## Dependencies

### External Dependencies

| Dependency | Type | Notes |
|------------|------|-------|
| Oracle ATP/On-Premise | Database | mTLS 연결, wallet 필요 |
| Google Cloud Storage | Storage | 서비스 계정 인증 |
| Oracle Instant Client | Library | godror 필수 의존성 |

### Internal Dependencies

```
Milestone 1 (Foundation)
    └── Milestone 2 (Oracle) ─┐
    └── Milestone 3 (GCS) ────┼── Milestone 4 (Transport) ── Milestone 5 (Monitoring)
                              │
                              └── Milestone 6 (Security)
                              └── Milestone 7 (Resilience)
                              └── Milestone 8 (Performance)

All Milestones ── Milestone 9 (Testing)
```

---

## Definition of Done

각 마일스톤은 다음 기준을 충족해야 완료로 간주됩니다:

1. **기능 완성**: 모든 Task가 구현됨
2. **테스트 통과**: 관련 테스트 모두 통과
3. **코드 리뷰**: 코드 리뷰 완료 및 승인
4. **문서화**: API 문서 및 주석 완성
5. **품질 게이트**: linter 경고 0개, 보안 스캔 통과

---

## Traceability

| Milestone | Requirements Covered |
|-----------|---------------------|
| M1 | NFR-04.1 |
| M2 | FR-01.1, FR-01.2, FR-01.3, FR-02.1, FR-02.2, FR-02.3, FR-02.4 |
| M3 | FR-03.1, FR-03.2, FR-03.3, FR-03.4 |
| M4 | FR-04.1, FR-04.2, FR-04.3, FR-04.4, FR-05.5, FR-05.6 |
| M5 | FR-06.1, FR-06.2, FR-06.3 |
| M6 | NFR-03.3, NFR-03.4 |
| M7 | NFR-02.1, NFR-02.2, NFR-02.3, NFR-02.4 |
| M8 | NFR-01.1, NFR-01.2, NFR-01.3, NFR-01.4 |
| M9 | All (Verification) |
