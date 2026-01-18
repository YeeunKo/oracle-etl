# SPEC-ETL-001: Oracle ETL Pipeline Backend API

## Metadata

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-ETL-001 |
| Title | Oracle DB to Google Cloud Storage ETL Pipeline Backend API |
| Created | 2026-01-18 |
| Status | Planned |
| Priority | High |
| Lifecycle | spec-anchored |
| Version | 1.0.0 |

## Executive Summary

고성능 Oracle DB에서 Google Cloud Storage로의 ETL 파이프라인 백엔드 API 개발. 목표 처리량 100,000 rows/second를 달성하며, 실시간 모니터링과 Transport 기반 테이블 그룹 관리를 지원합니다.

---

## Environment

### Source System
- **Database**: Oracle EBS 12.1.2 on Oracle 11g (On-Premise)
- **Connection**: Oracle ATP via mTLS wallet 또는 On-Premise 직접 연결
- **Data Volume**: 약 4.7 TB, 월 20GB 증가
- **Concurrent Users**: 피크 타임 약 300명 동시 접속

### Target System
- **Storage**: Google Cloud Storage (Standard)
- **Format**: JSON Lines (NDJSON) with gzip compression
- **Upload Mode**: Streaming resumable upload

### Runtime Environment
- **Platform**: GCP Compute Engine (e2-standard-4: 4 vCPU, 16GB RAM)
- **Runtime**: Go 1.23+
- **Network**: VPC with Oracle ATP/On-Premise connectivity

### Technology Stack
- **Web Framework**: Go Fiber v2.52+
- **Oracle Driver**: godror (Oracle ODPI-C driver for Go)
- **GCS SDK**: cloud.google.com/go/storage v1.40+
- **Compression**: gzip (compress/gzip)
- **Real-time**: Server-Sent Events (SSE)

---

## Assumptions

### Technical Assumptions

| ID | Assumption | Confidence | Evidence | Risk if Wrong | Validation Method |
|----|------------|------------|----------|---------------|-------------------|
| A-01 | Oracle ATP wallet 파일 접근 가능 | High | Oracle Cloud 표준 기능 | 연결 불가 | wallet 파일 테스트 연결 |
| A-02 | GCS 서비스 계정 인증 설정 완료 | High | GCP 표준 설정 | 업로드 실패 | 테스트 업로드 수행 |
| A-03 | 네트워크 대역폭 100Mbps+ 확보 | Medium | e2-standard-4 기본 사양 | 전송 병목 | iperf 테스트 |
| A-04 | 테이블당 평균 250,000 rows | Medium | 사용자 제공 정보 | 예상 시간 오차 | COUNT(*) 검증 |
| A-05 | 컬럼 수 평균 50-300개 | High | VBRP 샘플 274컬럼 | 메모리 사용량 변동 | 스키마 분석 |

### Business Assumptions

| ID | Assumption | Confidence | Risk if Wrong |
|----|------------|------------|---------------|
| B-01 | ETL 작업은 비즈니스 시간 외 수행 | Medium | DB 부하로 인한 서비스 영향 |
| B-02 | 데이터 무결성 > 처리 속도 우선 | High | 재처리 필요 시 혼란 |
| B-03 | 일 1회 이상 full extract 수행 | Medium | 증분 추출 로직 필요 |

---

## Requirements

### Functional Requirements

#### FR-01: Oracle Connection Management

**FR-01.1: mTLS Wallet Connection** (Ubiquitous)
- 시스템은 **항상** Oracle ATP mTLS wallet을 사용하여 보안 연결을 수립해야 한다
- wallet 경로, TNS 이름, 사용자 자격 증명은 환경 변수로 설정한다

**FR-01.2: Connection Pooling** (Ubiquitous)
- 시스템은 **항상** 커넥션 풀을 유지하여 다중 동시 연결을 지원해야 한다
- 기본 풀 크기: 최소 5, 최대 20 커넥션

**FR-01.3: Connection Health Check** (Event-Driven)
- **WHEN** `/api/health` 또는 `/api/oracle/status` 엔드포인트 호출 시
- **THEN** Oracle 연결 상태를 확인하고 JSON 응답을 반환한다

#### FR-02: Data Extraction

**FR-02.1: Batch Fetch** (Ubiquitous)
- 시스템은 **항상** FetchArraySize=1000으로 배치 페치를 수행해야 한다
- 메모리 효율성과 처리량 균형을 유지한다

**FR-02.2: SQL Query Execution** (Event-Driven)
- **WHEN** 사용자가 테이블 추출 요청 시
- **THEN** 해당 테이블의 모든 데이터를 SELECT * 쿼리로 추출한다

**FR-02.3: Chunked Streaming** (Ubiquitous)
- 시스템은 **항상** 10,000 rows 단위로 데이터를 청크하여 스트리밍해야 한다
- 각 청크는 독립적으로 처리 및 업로드된다

**FR-02.4: Sample Data Preview** (Event-Driven)
- **WHEN** `/api/tables/:name/sample` 엔드포인트 호출 시
- **THEN** 해당 테이블의 상위 100개 rows를 JSON으로 반환한다

#### FR-03: GCS Upload

**FR-03.1: Streaming Upload** (Ubiquitous)
- 시스템은 **항상** GCS로 스트리밍 업로드를 수행해야 한다
- 전체 데이터를 메모리에 적재하지 않고 청크 단위로 전송한다

**FR-03.2: NDJSON with Gzip** (Ubiquitous)
- 시스템은 **항상** JSON Lines 포맷으로 데이터를 직렬화해야 한다
- gzip 압축을 적용하여 전송 대역폭을 최적화한다

**FR-03.3: Resumable Upload** (State-Driven)
- **IF** 업로드 중 네트워크 오류 발생 시
- **THEN** resumable upload 세션을 사용하여 중단 지점부터 재개한다

**FR-03.4: Object Naming Convention** (Ubiquitous)
- 시스템은 **항상** 다음 명명 규칙을 따라야 한다
- 패턴: `{bucket}/{trpid}/{jobver}/{table_name}.jsonl.gz`
- 예시: `etl-bucket/TRP-001/v001/VBRP.jsonl.gz`

#### FR-04: Transport Management

**FR-04.1: Transport Creation** (Event-Driven)
- **WHEN** `POST /api/transport` 요청 시
- **THEN** 새로운 Transport ID (TRPID)를 생성하고 테이블 목록을 연결한다

**FR-04.2: Transport-Table Relationship** (Ubiquitous)
- 시스템은 **항상** Table : Transport = 1:N 관계를 유지해야 한다
- 하나의 테이블은 여러 Transport에 포함될 수 있다

**FR-04.3: Job Versioning** (Event-Driven)
- **WHEN** Transport 실행 요청 시
- **THEN** JOBVER를 자동 증가시키고 새로운 Job을 생성한다
- JOBVER 형식: `v001`, `v002`, ...

**FR-04.4: Transport Execution** (Event-Driven)
- **WHEN** `POST /api/transport/:id/execute` 요청 시
- **THEN** 해당 Transport의 모든 테이블에 대해 ETL 작업을 순차/병렬 실행한다

#### FR-05: API Endpoints

**FR-05.1: Health Check Endpoint** (Event-Driven)
- **WHEN** `GET /api/health` 요청 시
- **THEN** 시스템 전반적인 상태를 JSON으로 반환한다
- 응답: `{ "status": "ok", "timestamp": "...", "version": "..." }`

**FR-05.2: Oracle Status Endpoint** (Event-Driven)
- **WHEN** `GET /api/oracle/status` 요청 시
- **THEN** Oracle 연결 상태, 풀 상태, 연결 테스트 결과를 반환한다

**FR-05.3: Table List Endpoint** (Event-Driven)
- **WHEN** `GET /api/tables` 요청 시
- **THEN** 접근 가능한 모든 테이블 목록과 각 테이블의 row count를 반환한다

**FR-05.4: Sample Data Endpoint** (Event-Driven)
- **WHEN** `GET /api/tables/:name/sample` 요청 시
- **THEN** 해당 테이블의 샘플 데이터 100개 rows를 반환한다

**FR-05.5: Transport Management Endpoints** (Event-Driven)
- `POST /api/transport`: Transport 생성
- `GET /api/transport/:id`: Transport 상세 조회
- `POST /api/transport/:id/execute`: Transport 실행
- `GET /api/transport/:id/status`: Transport 실행 상태 (SSE)

**FR-05.6: Job History Endpoint** (Event-Driven)
- **WHEN** `GET /api/jobs` 요청 시
- **THEN** 과거 Job 실행 이력을 페이지네이션과 함께 반환한다

#### FR-06: Real-time Monitoring

**FR-06.1: SSE Progress Stream** (State-Driven)
- **IF** Transport 실행 중이면
- **THEN** SSE를 통해 실시간 진행 상황을 스트리밍한다
- 이벤트: `progress`, `table_start`, `table_complete`, `error`, `complete`

**FR-06.2: Progress Metrics** (Ubiquitous)
- 시스템은 **항상** 다음 메트릭을 제공해야 한다
  - `rows_processed`: 처리된 총 row 수
  - `rows_per_second`: 현재 처리 속도
  - `bytes_transferred`: 전송된 총 바이트
  - `current_table`: 현재 처리 중인 테이블명
  - `progress_percent`: 전체 진행률 (%)

**FR-06.3: Error Notification** (Event-Driven)
- **WHEN** ETL 작업 중 오류 발생 시
- **THEN** SSE 채널을 통해 에러 이벤트를 즉시 발송한다

---

### Non-Functional Requirements

#### NFR-01: Performance

**NFR-01.1: Throughput Target** (Ubiquitous)
- 시스템은 **항상** 최소 100,000 rows/second 처리량을 달성해야 한다

**NFR-01.2: API Response Time** (Ubiquitous)
- 시스템은 **항상** 일반 API 응답 시간 100ms 이하를 유지해야 한다
- 제외: 대량 데이터 조회 엔드포인트 (sample, tables with counts)

**NFR-01.3: Memory Efficiency** (Ubiquitous)
- 시스템은 **항상** 청크 기반 스트리밍으로 메모리 사용량 2GB 이하를 유지해야 한다

**NFR-01.4: Scalability** (State-Driven)
- **IF** 테이블 수가 100개 이상이면
- **THEN** 병렬 처리를 통해 선형적 성능 확장을 지원한다

#### NFR-02: Reliability

**NFR-02.1: Graceful Error Handling** (Ubiquitous)
- 시스템은 **항상** 모든 오류를 캡처하고 구조화된 에러 응답을 반환해야 한다

**NFR-02.2: Automatic Retry** (State-Driven)
- **IF** 일시적 네트워크 오류 발생 시
- **THEN** 최대 3회 exponential backoff로 재시도한다

**NFR-02.3: Transaction Safety** (Ubiquitous)
- 시스템은 **항상** 읽기 전용 트랜잭션으로 데이터를 추출해야 한다
- Oracle 스냅샷 격리를 사용하여 일관된 읽기를 보장한다

**NFR-02.4: Graceful Shutdown** (Event-Driven)
- **WHEN** SIGTERM/SIGINT 시그널 수신 시
- **THEN** 진행 중인 작업을 완료하고 리소스를 정리한 후 종료한다

#### NFR-03: Security

**NFR-03.1: mTLS Connection** (Ubiquitous)
- 시스템은 **항상** Oracle ATP 연결에 mTLS를 사용해야 한다

**NFR-03.2: Service Account Auth** (Ubiquitous)
- 시스템은 **항상** GCS 접근에 서비스 계정 인증을 사용해야 한다

**NFR-03.3: API Authentication** (State-Driven)
- **IF** API 엔드포인트 접근 시
- **THEN** API Key 또는 Bearer Token 인증을 요구한다
- 헤더: `Authorization: Bearer <token>` 또는 `X-API-Key: <key>`

**NFR-03.4: Sensitive Data Protection** (Unwanted)
- 시스템은 로그에 비밀번호, wallet 자격 증명, API 키를 **기록하지 않아야 한다**

#### NFR-04: Observability

**NFR-04.1: Structured Logging** (Ubiquitous)
- 시스템은 **항상** JSON 형식의 구조화된 로그를 출력해야 한다
- 필수 필드: timestamp, level, message, request_id, transport_id

**NFR-04.2: Metrics Endpoint** (Optional)
- **가능하면** `/metrics` Prometheus 엔드포인트를 제공한다

---

## Specifications

### API Specification

#### Base URL
```
/api/v1
```

#### Endpoints

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| GET | /health | System health check | No |
| GET | /oracle/status | Oracle connection status | Yes |
| GET | /tables | List available tables | Yes |
| GET | /tables/:name/sample | Get sample data (100 rows) | Yes |
| POST | /transport | Create new transport | Yes |
| GET | /transport/:id | Get transport details | Yes |
| POST | /transport/:id/execute | Execute ETL for transport | Yes |
| GET | /transport/:id/status | Get transport status (SSE) | Yes |
| GET | /jobs | List job history | Yes |

#### Request/Response Examples

**POST /api/transport**
```json
// Request
{
  "name": "SAP Sales Data Export",
  "tables": ["VBRP", "VBRK", "LIKP"],
  "target_bucket": "oracle-etl-data",
  "description": "Monthly sales data extraction"
}

// Response (201 Created)
{
  "id": "TRP-20260118-001",
  "name": "SAP Sales Data Export",
  "tables": ["VBRP", "VBRK", "LIKP"],
  "target_bucket": "oracle-etl-data",
  "created_at": "2026-01-18T10:30:00Z",
  "status": "created"
}
```

**GET /api/transport/:id/status (SSE)**
```
event: progress
data: {"table": "VBRP", "rows_processed": 50000, "rows_per_second": 105234, "progress_percent": 20}

event: table_complete
data: {"table": "VBRP", "total_rows": 250000, "bytes_transferred": 45678901, "duration_seconds": 2.4}

event: complete
data: {"transport_id": "TRP-20260118-001", "job_version": "v001", "total_rows": 750000, "total_duration_seconds": 7.5}
```

### Data Model

#### Transport
```go
type Transport struct {
    ID          string    `json:"id"`           // TRP-YYYYMMDD-NNN
    Name        string    `json:"name"`
    Tables      []string  `json:"tables"`
    TargetBucket string   `json:"target_bucket"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    Status      string    `json:"status"`       // created, running, completed, failed
}
```

#### Job
```go
type Job struct {
    ID           string    `json:"id"`            // JOB-YYYYMMDD-NNN
    TransportID  string    `json:"transport_id"`
    Version      string    `json:"version"`       // v001, v002, ...
    StartedAt    time.Time `json:"started_at"`
    CompletedAt  *time.Time `json:"completed_at,omitempty"`
    Status       string    `json:"status"`        // running, completed, failed
    TotalRows    int64     `json:"total_rows"`
    BytesTransferred int64 `json:"bytes_transferred"`
    Error        *string   `json:"error,omitempty"`
}
```

#### TableExtraction
```go
type TableExtraction struct {
    TableName     string    `json:"table_name"`
    RowCount      int64     `json:"row_count"`
    BytesTransferred int64  `json:"bytes_transferred"`
    StartedAt     time.Time `json:"started_at"`
    CompletedAt   *time.Time `json:"completed_at,omitempty"`
    Status        string    `json:"status"`       // pending, running, completed, failed
    GCSPath       string    `json:"gcs_path"`
}
```

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Go Fiber API Server                         │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐  │
│  │   Handlers  │  │   Services  │  │ Extractors  │  │  Uploaders │  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └─────┬─────┘  │
│         │                │                │                │        │
│         v                v                v                v        │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Domain Layer                             │   │
│  │  Transport | Job | TableExtraction | Progress               │   │
│  └─────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐                           ┌──────────────────┐   │
│  │ Oracle Pool  │◄──── godror ────────────► │ Oracle ATP/      │   │
│  │ (Connection) │                           │ On-Premise DB    │   │
│  └──────────────┘                           └──────────────────┘   │
│                                                                     │
│  ┌──────────────┐                           ┌──────────────────┐   │
│  │ GCS Client   │◄──── cloud.google.com ──► │ Google Cloud     │   │
│  │ (Storage)    │      /go/storage          │ Storage          │   │
│  └──────────────┘                           └──────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### Configuration

```yaml
# config.yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 300s  # Long timeout for ETL operations

oracle:
  wallet_path: /opt/wallet
  tns_name: oracle_atp_high
  username: ${ORACLE_USER}
  password: ${ORACLE_PASSWORD}
  pool_min: 5
  pool_max: 20
  fetch_array_size: 1000

gcs:
  project_id: ${GCP_PROJECT_ID}
  credentials_file: /opt/gcp/service-account.json
  default_bucket: oracle-etl-data

etl:
  chunk_size: 10000
  parallel_tables: 4
  retry_attempts: 3
  retry_backoff: 1s

auth:
  api_key: ${API_KEY}
  enable_bearer: true
```

---

## Constraints

### Technical Constraints

| ID | Constraint | Rationale |
|----|------------|-----------|
| TC-01 | Go 1.23+ 필수 | 최신 제네릭 및 성능 최적화 활용 |
| TC-02 | Oracle Instant Client 21c+ | godror 호환성 |
| TC-03 | GCP 리전: asia-northeast3 (서울) | 레이턴시 최소화 |
| TC-04 | 메모리 16GB 이하 사용 | e2-standard-4 사양 제한 |

### Business Constraints

| ID | Constraint | Rationale |
|----|------------|-----------|
| BC-01 | ETL 작업은 업무 외 시간 권장 | 소스 DB 부하 최소화 |
| BC-02 | 데이터 보관 기간 90일 | GCS 비용 관리 |
| BC-03 | 동시 Transport 실행 1개 | 리소스 경합 방지 |

---

## Traceability

| Requirement | Test Scenario | Acceptance Criteria |
|-------------|---------------|---------------------|
| FR-01.1 | TC-CONN-001 | AC-01 |
| FR-02.1 | TC-EXTRACT-001 | AC-02 |
| FR-02.3 | TC-CHUNK-001 | AC-02 |
| FR-03.1 | TC-UPLOAD-001 | AC-03 |
| FR-04.1 | TC-TRANSPORT-001 | AC-04 |
| FR-06.1 | TC-SSE-001 | AC-05 |
| NFR-01.1 | TC-PERF-001 | AC-06 |
| NFR-02.2 | TC-RETRY-001 | AC-07 |
| NFR-03.3 | TC-AUTH-001 | AC-08 |

---

## Related Documents

- SPEC-ETL-001/plan.md - Implementation Plan
- SPEC-ETL-001/acceptance.md - Acceptance Criteria
- sample/vbrp_schema.json - Sample Table Schema (274 columns)
- sample/vbrp_data.json - Sample Data (JSONL format)

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-18 | Claude | Initial SPEC creation |
