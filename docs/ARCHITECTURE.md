# 아키텍처 문서

Oracle ETL Pipeline의 시스템 아키텍처 및 설계 문서

## 목차

- [개요](#개요)
- [Clean Architecture](#clean-architecture)
- [디렉토리 구조](#디렉토리-구조)
- [데이터 흐름](#데이터-흐름)
- [컴포넌트 상세](#컴포넌트-상세)
- [회복성 패턴](#회복성-패턴)
- [성능 최적화](#성능-최적화)

---

## 개요

Oracle ETL Pipeline은 Clean Architecture 원칙을 따르는 Go 기반 ETL 시스템입니다.

### 시스템 구성

```
┌─────────────────────────────────────────────────────────────────────┐
│                           클라이언트                                  │
│                    (REST API / SSE 클라이언트)                        │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Oracle ETL API 서버                           │
│  ┌─────────────┬──────────────┬──────────────┬────────────────────┐ │
│  │   Handler   │  Middleware   │   UseCase    │      Domain       │ │
│  │  (Fiber)    │  (Auth/Log)   │  (Service)   │     (Model)       │ │
│  └─────────────┴──────────────┴──────────────┴────────────────────┘ │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   Oracle DB     │  │ Google Cloud    │  │   In-Memory     │
│   (godror)      │  │ Storage (GCS)   │  │   Repository    │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

### 핵심 기술 스택

| 계층 | 기술 | 용도 |
|------|------|------|
| HTTP 서버 | Fiber v2 | 고성능 웹 프레임워크 |
| Oracle 연결 | godror | Oracle 데이터베이스 드라이버 |
| GCS 클라이언트 | cloud.google.com/go/storage | 클라우드 스토리지 업로드 |
| 설정 관리 | Viper | YAML/환경변수 설정 로드 |
| 로깅 | zerolog | 구조화된 JSON 로깅 |
| 인증 | golang-jwt/jwt | JWT 토큰 검증 |

---

## Clean Architecture

### 계층 구조

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Frameworks & Drivers                            │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │  HTTP Handler (Fiber) │ Oracle Adapter │ GCS Adapter │ SSE    │  │
│  └───────────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────────┤
│                      Interface Adapters                              │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │      Repository Interface      │      Handler Interface       │  │
│  └───────────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────────┤
│                       Use Cases (Application)                        │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │    TransportService     │     JobService     │   Executor     │  │
│  └───────────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────────┤
│                          Entities (Domain)                           │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │      Transport       │        Job        │     Extraction     │  │
│  └───────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### 의존성 규칙

- **Domain**: 핵심 비즈니스 엔티티, 외부 의존성 없음
- **UseCase**: 비즈니스 로직, Domain에만 의존
- **Adapter**: 외부 시스템 연결, UseCase와 Domain에 의존
- **Framework**: HTTP, DB 드라이버, Adapter와 UseCase에 의존

---

## 디렉토리 구조

```
oracle-etl/
├── cmd/
│   └── server/
│       └── main.go                 # 서버 진입점
│
├── internal/                       # 내부 패키지 (외부 노출 불가)
│   │
│   ├── domain/                     # 도메인 계층 (비즈니스 엔티티)
│   │   ├── transport.go            # Transport 엔티티
│   │   ├── job.go                  # Job 엔티티
│   │   └── extraction.go           # Extraction 엔티티 및 관련 타입
│   │
│   ├── usecase/                    # 유스케이스 계층 (비즈니스 로직)
│   │   ├── transport_service.go    # Transport CRUD 로직
│   │   ├── job_service.go          # Job 관리 로직
│   │   └── parallel_executor.go    # 병렬 실행 로직
│   │
│   ├── repository/                 # 저장소 인터페이스
│   │   ├── transport_repo.go       # Transport 저장소 인터페이스
│   │   ├── job_repo.go             # Job 저장소 인터페이스
│   │   └── memory/                 # In-Memory 구현
│   │       ├── transport_repo.go
│   │       └── job_repo.go
│   │
│   ├── adapter/                    # 어댑터 계층 (외부 시스템 연결)
│   │   ├── handler/                # HTTP 핸들러
│   │   │   ├── health.go           # 헬스체크 핸들러
│   │   │   ├── transport.go        # Transport API 핸들러
│   │   │   ├── job.go              # Job API 핸들러
│   │   │   ├── status.go           # SSE 상태 핸들러
│   │   │   ├── oracle.go           # Oracle 상태 핸들러
│   │   │   └── table.go            # 테이블 정보 핸들러
│   │   │
│   │   ├── oracle/                 # Oracle 어댑터
│   │   │   ├── pool.go             # 커넥션 풀 관리
│   │   │   ├── repository.go       # 데이터 조회
│   │   │   ├── extractor.go        # 데이터 추출
│   │   │   └── mock_repository.go  # 테스트용 Mock
│   │   │
│   │   ├── gcs/                    # GCS 어댑터
│   │   │   ├── client.go           # GCS 클라이언트
│   │   │   └── uploader.go         # 스트리밍 업로더
│   │   │
│   │   └── sse/                    # SSE 어댑터
│   │       ├── broadcaster.go      # 이벤트 브로드캐스터
│   │       ├── event.go            # 이벤트 타입 정의
│   │       └── metrics.go          # SSE 메트릭
│   │
│   ├── config/                     # 설정 관리
│   │   └── config.go               # Viper 기반 설정 로드
│   │
│   ├── middleware/                 # HTTP 미들웨어
│   │   ├── auth.go                 # 인증 미들웨어
│   │   ├── logging.go              # 요청 로깅
│   │   ├── recovery.go             # 패닉 복구
│   │   ├── ratelimit.go            # Rate Limiting
│   │   ├── cors.go                 # CORS 처리
│   │   └── masking.go              # 민감 정보 마스킹
│   │
│   ├── errors/                     # 에러 처리
│   │   └── errors.go               # 구조화된 에러 응답
│   │
│   └── resilience/                 # 회복성 패턴
│       ├── retry.go                # 재시도 로직
│       ├── circuit.go              # Circuit Breaker
│       └── shutdown.go             # Graceful Shutdown
│
├── pkg/                            # 공개 유틸리티 패키지
│   ├── buffer/                     # 버퍼 관리
│   │   └── config.go
│   ├── compress/                   # 압축 유틸리티
│   │   └── gzip.go
│   ├── jsonl/                      # JSONL 인코딩
│   │   └── encoder.go
│   └── pool/                       # 워커 풀
│       └── worker.go
│
├── benchmarks/                     # 벤치마크 테스트
│   ├── extraction_benchmark_test.go
│   ├── buffer_benchmark_test.go
│   ├── pool_benchmark_test.go
│   └── parallel_executor_benchmark_test.go
│
├── tests/                          # 테스트
│   ├── integration/                # 통합 테스트
│   │   ├── oracle_integration_test.go
│   │   └── gcs_integration_test.go
│   └── e2e/                        # E2E 테스트
│       └── etl_flow_test.go
│
├── config.yaml                     # 설정 파일
├── go.mod                          # Go 모듈 정의
└── go.sum                          # 의존성 체크섬
```

### 패키지 설명

| 패키지 | 설명 | 의존성 |
|--------|------|--------|
| `domain` | 비즈니스 엔티티 | 없음 |
| `usecase` | 비즈니스 로직 | domain, repository |
| `repository` | 데이터 저장소 | domain |
| `adapter/handler` | HTTP 핸들러 | usecase, domain |
| `adapter/oracle` | Oracle 클라이언트 | domain, godror |
| `adapter/gcs` | GCS 클라이언트 | domain, cloud.google.com |
| `adapter/sse` | SSE 브로드캐스터 | 없음 |
| `middleware` | HTTP 미들웨어 | config |
| `config` | 설정 관리 | viper |
| `errors` | 에러 처리 | 없음 |
| `resilience` | 회복성 패턴 | 없음 |
| `pkg/*` | 공개 유틸리티 | 없음 |

---

## 데이터 흐름

### ETL 실행 흐름

```
┌──────────────┐
│   Client     │
│  POST /api/  │
│ transports/  │
│ :id/execute  │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│                      TransportHandler.Execute                     │
│  1. Transport 존재 확인                                           │
│  2. 실행 가능 상태 확인 (enabled && status != running)            │
│  3. JobService.CreateJob() 호출                                  │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│                        JobService.CreateJob                       │
│  1. 새 Job ID 생성 (JOB-YYYYMMDD-HHMMSS-xxxx)                    │
│  2. Transport별 버전 번호 증가                                    │
│  3. Job 저장소에 저장                                             │
│  4. Job 반환                                                      │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│                     ParallelExecutor (비동기)                     │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  goroutine pool (parallel_tables 설정만큼)                  │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │  │
│  │  │  Worker 1   │  │  Worker 2   │  │  Worker 3   │  ...    │  │
│  │  │ (Table A)   │  │ (Table B)   │  │ (Table C)   │         │  │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │  │
│  │         │                │                │                 │  │
│  └─────────┴────────────────┴────────────────┴─────────────────┘  │
└──────────────────────────────┬───────────────────────────────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Oracle Extractor                              │
│  1. 테이블에서 청크 단위(10,000 rows)로 데이터 추출               │
│  2. 배치 페치(fetch_array_size: 1000)로 효율적 조회               │
│  3. JSONL 형식으로 직렬화                                        │
│  4. Gzip 압축                                                    │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      GCS Uploader                                 │
│  1. 스트리밍 업로드 (16MB 청크)                                   │
│  2. 경로: gs://bucket/{transport_id}/v{version}/{table}.jsonl.gz │
│  3. Resumable upload 지원                                        │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      SSE Broadcaster                              │
│  1. 진행률 이벤트 브로드캐스트                                     │
│  2. 완료/실패 이벤트 전송                                         │
│  3. 연결된 클라이언트에 실시간 전달                                │
└─────────────────────────────────────────────────────────────────┘
```

### 데이터 변환 파이프라인

```
Oracle DB Row
     │
     ▼
┌─────────────────────────────────────────┐
│  map[string]interface{}                  │
│  { "ID": 1, "NAME": "Test", "DATE": ... }│
└──────────────────────┬───────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────┐
│  JSONL Encoder                           │
│  {"ID":1,"NAME":"Test","DATE":"..."}    │
│  {"ID":2,"NAME":"Test2","DATE":"..."}   │
└──────────────────────┬───────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────┐
│  Gzip Compressor                         │
│  ~60-70% 압축률                          │
└──────────────────────┬───────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────┐
│  GCS Upload                              │
│  gs://bucket/TRPID-xxx/v001/TABLE.jsonl.gz
└─────────────────────────────────────────┘
```

---

## 컴포넌트 상세

### Domain Layer

#### Transport

```go
type Transport struct {
    ID          string          // TRPID-{uuid[:8]}
    Name        string          // 이름
    Description string          // 설명
    Tables      []string        // 추출 대상 테이블
    Enabled     bool            // 활성화 여부
    Schedule    *CronSchedule   // Cron 스케줄 (선택)
    Status      TransportStatus // idle | running | failed
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### Job

```go
type Job struct {
    ID          string       // JOB-{YYYYMMDD-HHMMSS}-{random}
    TransportID string       // 연결된 Transport
    Version     int          // Transport별 버전 (v001, v002, ...)
    Status      JobStatus    // pending | running | completed | failed | cancelled
    StartedAt   *time.Time
    CompletedAt *time.Time
    Extractions []Extraction // 테이블별 추출 결과
    Error       *string
    Metrics     JobMetrics   // 실행 메트릭
    CreatedAt   time.Time
}
```

#### Extraction

```go
type Extraction struct {
    ID          string           // 추출 ID
    JobID       string           // 연결된 Job
    TableName   string           // 테이블 이름
    Status      ExtractionStatus // pending | running | completed | failed
    RowCount    int64            // 추출된 row 수
    ByteCount   int64            // 전송된 바이트
    GCSPath     string           // GCS 객체 경로
    StartedAt   *time.Time
    CompletedAt *time.Time
    Error       *string
}
```

### Adapter Layer

#### Oracle Pool

- 커넥션 풀 관리 (pool_min ~ pool_max)
- mTLS Wallet 기반 연결
- 헬스체크 및 상태 모니터링

#### GCS Uploader

- Resumable Upload 지원
- 16MB 청크 단위 업로드
- 타임아웃 및 재시도 처리

#### SSE Broadcaster

- Pub/Sub 패턴 구현
- 클라이언트별 채널 관리
- 연결 종료 시 자동 정리

### Middleware Layer

| 미들웨어 | 역할 |
|----------|------|
| Recovery | 패닉 복구 및 로깅 |
| Logging | 요청/응답 로깅 |
| Auth | API Key / JWT 인증 |
| RateLimit | 요청 제한 |
| CORS | Cross-Origin 처리 |
| Masking | 민감 정보 마스킹 |

---

## 회복성 패턴

### Retry Pattern

```
┌───────────────────────────────────────────────────────────────┐
│                     Exponential Backoff                        │
│                                                                │
│  시도 1 ──[실패]──> 1초 대기 ──> 시도 2 ──[실패]──> 2초 대기   │
│       ──> 시도 3 ──[실패]──> 4초 대기 ──> 시도 4 (최종 실패)   │
│                                                                │
│  설정:                                                         │
│  - retry_attempts: 3 (기본값)                                  │
│  - retry_backoff: 1s (초기 대기 시간)                          │
│  - max_backoff: 30s (최대 대기 시간)                           │
└───────────────────────────────────────────────────────────────┘
```

### Circuit Breaker Pattern

```
┌─────────────────────────────────────────────────────────────────┐
│                      Circuit Breaker States                      │
│                                                                  │
│  ┌──────────┐     실패 임계값 초과    ┌──────────┐              │
│  │  CLOSED  │ ─────────────────────> │   OPEN   │              │
│  │  (정상)   │ <───────────────────── │  (차단)   │              │
│  └──────────┘     성공                └────┬─────┘              │
│       ▲                                     │                    │
│       │                              타임아웃 후                 │
│       │                                     │                    │
│       │           ┌──────────────┐          │                    │
│       └───────────│  HALF-OPEN   │<─────────┘                    │
│         성공      │   (테스트)     │                              │
│                   └──────────────┘                               │
│                          │                                       │
│                   실패 시 다시 OPEN                               │
└─────────────────────────────────────────────────────────────────┘
```

### Graceful Shutdown

```
┌─────────────────────────────────────────────────────────────────┐
│                     Graceful Shutdown Flow                       │
│                                                                  │
│  1. SIGINT/SIGTERM 수신                                         │
│  2. 새 요청 수신 중단                                            │
│  3. SSE Broadcaster 종료 (클라이언트 연결 해제)                   │
│  4. 진행 중인 요청 완료 대기 (30초 타임아웃)                      │
│  5. Oracle 커넥션 풀 정리                                        │
│  6. 서버 종료                                                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## 성능 최적화

### Oracle 데이터 추출 최적화

| 설정 | 값 | 효과 |
|------|-----|------|
| `fetch_array_size` | 1000 | 네트워크 라운드트립 감소 |
| `prefetch_count` | 1000 | 메모리 버퍼링으로 속도 향상 |
| `chunk_size` | 10000 | 청크 단위 처리로 메모리 효율화 |
| `parallel_tables` | 4 | 동시 테이블 처리 |

### 병렬 처리 아키텍처

```
┌─────────────────────────────────────────────────────────────────┐
│                     Parallel Executor                            │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                      Worker Pool                          │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐      │   │
│  │  │ Worker1 │  │ Worker2 │  │ Worker3 │  │ Worker4 │      │   │
│  │  │ TABLE_A │  │ TABLE_B │  │ TABLE_C │  │ (idle)  │      │   │
│  │  └────┬────┘  └────┬────┘  └────┬────┘  └─────────┘      │   │
│  │       │            │            │                         │   │
│  │       ▼            ▼            ▼                         │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │              Result Channel                          │ │   │
│  │  │  [Extraction1] [Extraction2] [Extraction3] ...      │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 성능 지표

| 지표 | 목표 | 실측 |
|------|------|------|
| 데이터 추출 속도 | 100K rows/sec | 134K rows/sec |
| 메모리 사용량 | < 512MB | ~350MB (100만 rows) |
| GCS 업로드 속도 | 50MB/s | ~60MB/s |
| 압축률 | 60% | 65-70% |

### 메모리 관리

```
┌─────────────────────────────────────────────────────────────────┐
│                     Memory Buffer Strategy                       │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Streaming Pipeline (메모리 버퍼 최소화)                    │ │
│  │                                                             │ │
│  │  Oracle ──> Buffer (1000 rows) ──> JSONL ──> Gzip ──> GCS  │ │
│  │             ▲                                               │ │
│  │             │                                               │ │
│  │       fetch_array_size로 제어                               │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  - 전체 데이터를 메모리에 로드하지 않음                          │
│  - 스트리밍 방식으로 청크 단위 처리                              │
│  - 처리 완료된 청크는 즉시 GC 대상                               │
└─────────────────────────────────────────────────────────────────┘
```

---

## 배포 고려사항

### 환경별 설정

| 환경 | 권장 설정 |
|------|----------|
| 개발 | pool_max: 5, parallel_tables: 2 |
| 스테이징 | pool_max: 10, parallel_tables: 4 |
| 프로덕션 | pool_max: 20, parallel_tables: 8 |

### 컨테이너 배포

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /oracle-etl ./cmd/server

FROM alpine:latest
RUN apk add --no-cache libaio libnsl
COPY --from=builder /oracle-etl /oracle-etl
COPY config.yaml /config.yaml
EXPOSE 8080
CMD ["/oracle-etl"]
```

### 헬스체크

```yaml
# Kubernetes liveness/readiness probe
livenessProbe:
  httpGet:
    path: /api/health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /api/health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

---

## 향후 개선 계획

- [ ] PostgreSQL/MySQL 저장소 구현 (현재 In-Memory)
- [ ] Cron 스케줄러 통합
- [ ] Prometheus 메트릭 내보내기
- [ ] OpenTelemetry 분산 추적
- [ ] 증분 추출 (CDC) 지원
- [ ] 데이터 변환 파이프라인 플러그인
