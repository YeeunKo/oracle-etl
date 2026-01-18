# API 문서

Oracle ETL Pipeline API 엔드포인트 상세 문서

## 목차

- [개요](#개요)
- [인증](#인증)
- [에러 응답](#에러-응답)
- [엔드포인트](#엔드포인트)
  - [Health](#health)
  - [Transport](#transport)
  - [Job](#job)
  - [실시간 상태 (SSE)](#실시간-상태-sse)

---

## 개요

### 기본 정보

| 항목 | 값 |
|------|-----|
| Base URL | `http://localhost:8080/api` |
| Content-Type | `application/json` |
| 문자 인코딩 | UTF-8 |

### 공통 응답 형식

모든 성공 응답은 JSON 형식으로 반환됩니다.

```json
{
  "id": "TRPID-abc12345",
  "name": "Daily Export",
  "status": "idle",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

## 인증

API 인증은 두 가지 방법을 지원합니다.

### API Key 인증

요청 헤더에 `X-API-Key`를 포함합니다.

```bash
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/transports
```

### Bearer Token 인증 (JWT)

요청 헤더에 `Authorization: Bearer <token>`을 포함합니다.

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." http://localhost:8080/api/transports
```

#### JWT 토큰 구조

| 클레임 | 설명 |
|--------|------|
| `sub` | 사용자 ID |
| `exp` | 만료 시간 (Unix timestamp) |
| `iat` | 발급 시간 (Unix timestamp) |

### 인증 제외 경로

다음 경로는 인증 없이 접근 가능합니다:

- `/api/health`
- `/health`

---

## 에러 응답

### 에러 응답 형식

```json
{
  "code": "ERROR_CODE",
  "message": "사용자 친화적 에러 메시지",
  "details": {
    "field": "추가 정보"
  },
  "trace_id": "abc123"
}
```

### 에러 코드

| 코드 | HTTP 상태 | 설명 |
|------|----------|------|
| `VALIDATION_ERROR` | 400 | 요청 유효성 검사 실패 |
| `AUTHENTICATION_ERROR` | 401 | 인증 실패 |
| `TRANSPORT_NOT_FOUND` | 404 | Transport를 찾을 수 없음 |
| `JOB_NOT_FOUND` | 404 | Job을 찾을 수 없음 |
| `TRANSPORT_NOT_EXECUTABLE` | 409 | Transport가 실행 불가 상태 |
| `RATE_LIMIT_EXCEEDED` | 429 | 요청 제한 초과 |
| `ORACLE_CONNECTION_ERROR` | 503 | Oracle 연결 오류 |
| `GCS_UPLOAD_ERROR` | 502 | GCS 업로드 오류 |
| `INTERNAL_ERROR` | 500 | 내부 서버 오류 |

---

## 엔드포인트

---

### Health

서버 상태를 확인합니다.

#### GET /api/health

**인증**: 불필요

**응답 예시**

```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

**응답 필드**

| 필드 | 타입 | 설명 |
|------|------|------|
| `status` | string | 서버 상태 (`ok`) |
| `timestamp` | string | 응답 시간 (RFC3339) |
| `version` | string | 애플리케이션 버전 |

---

### Transport

ETL 전송 구성을 관리합니다.

#### POST /api/transports

새로운 Transport를 생성합니다.

**요청 본문**

```json
{
  "name": "Daily Sales Export",
  "description": "일일 매출 데이터 추출",
  "tables": ["SALES_ORDER", "SALES_LINE_ITEM", "CUSTOMER"]
}
```

**요청 필드**

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `name` | string | O | Transport 이름 |
| `description` | string | X | 설명 |
| `tables` | string[] | O | 추출할 테이블 목록 (최소 1개) |

**응답** (201 Created)

```json
{
  "id": "TRPID-abc12345",
  "name": "Daily Sales Export",
  "description": "일일 매출 데이터 추출",
  "tables": ["SALES_ORDER", "SALES_LINE_ITEM", "CUSTOMER"],
  "enabled": true,
  "status": "idle",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**에러 응답**

| 상태 | 코드 | 설명 |
|------|------|------|
| 400 | `INVALID_REQUEST` | 요청 본문 파싱 실패 |
| 400 | `VALIDATION_ERROR` | 필수 필드 누락 또는 잘못된 값 |

---

#### GET /api/transports

Transport 목록을 조회합니다.

**쿼리 파라미터**

| 파라미터 | 타입 | 기본값 | 설명 |
|----------|------|--------|------|
| `offset` | integer | 0 | 시작 위치 |
| `limit` | integer | 20 | 조회 개수 (최대 100) |

**요청 예시**

```bash
curl "http://localhost:8080/api/transports?offset=0&limit=10" \
  -H "X-API-Key: your-api-key"
```

**응답** (200 OK)

```json
{
  "transports": [
    {
      "id": "TRPID-abc12345",
      "name": "Daily Sales Export",
      "description": "일일 매출 데이터 추출",
      "tables": ["SALES_ORDER", "SALES_LINE_ITEM"],
      "enabled": true,
      "status": "idle",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 20
}
```

---

#### GET /api/transports/:id

특정 Transport를 조회합니다.

**경로 파라미터**

| 파라미터 | 타입 | 설명 |
|----------|------|------|
| `id` | string | Transport ID (TRPID-xxxxxxxx) |

**응답** (200 OK)

```json
{
  "id": "TRPID-abc12345",
  "name": "Daily Sales Export",
  "description": "일일 매출 데이터 추출",
  "tables": ["SALES_ORDER", "SALES_LINE_ITEM", "CUSTOMER"],
  "enabled": true,
  "schedule": {
    "expression": "0 2 * * *",
    "timezone": "Asia/Seoul"
  },
  "status": "idle",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**에러 응답**

| 상태 | 코드 | 설명 |
|------|------|------|
| 404 | `TRANSPORT_NOT_FOUND` | Transport를 찾을 수 없음 |

---

#### DELETE /api/transports/:id

Transport를 삭제합니다.

**경로 파라미터**

| 파라미터 | 타입 | 설명 |
|----------|------|------|
| `id` | string | Transport ID |

**응답** (204 No Content)

응답 본문 없음

**에러 응답**

| 상태 | 코드 | 설명 |
|------|------|------|
| 404 | `TRANSPORT_NOT_FOUND` | Transport를 찾을 수 없음 |

---

#### POST /api/transports/:id/execute

Transport를 실행하고 새 Job을 생성합니다.

**경로 파라미터**

| 파라미터 | 타입 | 설명 |
|----------|------|------|
| `id` | string | Transport ID |

**응답** (202 Accepted)

```json
{
  "job_id": "JOB-20240115-103000-a1b2",
  "transport_id": "TRPID-abc12345",
  "version": 1,
  "status": "pending"
}
```

**응답 필드**

| 필드 | 타입 | 설명 |
|------|------|------|
| `job_id` | string | 생성된 Job ID |
| `transport_id` | string | Transport ID |
| `version` | integer | Job 버전 (Transport별 증가) |
| `status` | string | 초기 상태 (`pending`) |

**에러 응답**

| 상태 | 코드 | 설명 |
|------|------|------|
| 404 | `TRANSPORT_NOT_FOUND` | Transport를 찾을 수 없음 |
| 409 | `TRANSPORT_NOT_EXECUTABLE` | 이미 실행 중이거나 비활성화 상태 |
| 500 | `JOB_CREATION_FAILED` | Job 생성 실패 |

---

### Job

ETL 실행 이력을 관리합니다.

#### GET /api/jobs

Job 목록을 조회합니다.

**쿼리 파라미터**

| 파라미터 | 타입 | 기본값 | 설명 |
|----------|------|--------|------|
| `transport_id` | string | - | Transport ID로 필터링 |
| `status` | string | - | 상태로 필터링 |
| `offset` | integer | 0 | 시작 위치 |
| `limit` | integer | 20 | 조회 개수 |

**Job 상태 값**

| 상태 | 설명 |
|------|------|
| `pending` | 대기 중 |
| `running` | 실행 중 |
| `completed` | 완료 |
| `failed` | 실패 |
| `cancelled` | 취소됨 |

**요청 예시**

```bash
curl "http://localhost:8080/api/jobs?transport_id=TRPID-abc12345&status=completed" \
  -H "X-API-Key: your-api-key"
```

**응답** (200 OK)

```json
{
  "jobs": [
    {
      "id": "JOB-20240115-103000-a1b2",
      "transport_id": "TRPID-abc12345",
      "version": 1,
      "status": "completed",
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:35:00Z",
      "extractions": [
        {
          "id": "EXT-001",
          "job_id": "JOB-20240115-103000-a1b2",
          "table_name": "SALES_ORDER",
          "status": "completed",
          "row_count": 150000,
          "byte_count": 45000000,
          "gcs_path": "gs://bucket/TRPID-abc12345/v001/SALES_ORDER.jsonl.gz"
        }
      ],
      "metrics": {
        "total_rows": 150000,
        "total_bytes": 45000000,
        "duration": 300000000000,
        "rows_per_second": 500.0
      },
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 20
}
```

---

#### GET /api/jobs/:id

특정 Job을 조회합니다.

**경로 파라미터**

| 파라미터 | 타입 | 설명 |
|----------|------|------|
| `id` | string | Job ID |

**응답** (200 OK)

```json
{
  "id": "JOB-20240115-103000-a1b2",
  "transport_id": "TRPID-abc12345",
  "version": 1,
  "status": "completed",
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:35:00Z",
  "extractions": [
    {
      "id": "EXT-001",
      "job_id": "JOB-20240115-103000-a1b2",
      "table_name": "SALES_ORDER",
      "status": "completed",
      "row_count": 150000,
      "byte_count": 45000000,
      "gcs_path": "gs://bucket/TRPID-abc12345/v001/SALES_ORDER.jsonl.gz",
      "started_at": "2024-01-15T10:30:05Z",
      "completed_at": "2024-01-15T10:32:00Z"
    }
  ],
  "metrics": {
    "total_rows": 150000,
    "total_bytes": 45000000,
    "duration": 300000000000,
    "rows_per_second": 500.0
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

**에러 응답**

| 상태 | 코드 | 설명 |
|------|------|------|
| 404 | `JOB_NOT_FOUND` | Job을 찾을 수 없음 |

---

### 실시간 상태 (SSE)

Server-Sent Events를 통해 Transport 실행 상태를 실시간으로 모니터링합니다.

#### GET /api/transports/:id/status

Transport의 실시간 상태를 SSE 스트림으로 받습니다.

**경로 파라미터**

| 파라미터 | 타입 | 설명 |
|----------|------|------|
| `id` | string | Transport ID |

**응답 헤더**

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Accel-Buffering: no
```

**이벤트 타입**

| 이벤트 | 설명 |
|--------|------|
| `connected` | SSE 연결 설정됨 |
| `job_started` | Job 실행 시작 |
| `extraction_started` | 테이블 추출 시작 |
| `extraction_progress` | 추출 진행률 업데이트 |
| `extraction_completed` | 테이블 추출 완료 |
| `extraction_failed` | 테이블 추출 실패 |
| `job_completed` | Job 완료 |
| `job_failed` | Job 실패 |

**이벤트 형식**

```
event: connected
data: {"transport_id":"TRPID-abc12345","client_id":"client-123","message":"SSE 연결이 설정되었습니다"}

event: extraction_progress
data: {"table":"SALES_ORDER","rows_processed":50000,"total_rows":150000,"percentage":33.3}

event: job_completed
data: {"job_id":"JOB-20240115-103000-a1b2","status":"completed","metrics":{"total_rows":150000,"duration_seconds":300}}
```

**사용 예시**

JavaScript:
```javascript
const eventSource = new EventSource('/api/transports/TRPID-abc12345/status');

eventSource.addEventListener('connected', (event) => {
  console.log('Connected:', JSON.parse(event.data));
});

eventSource.addEventListener('extraction_progress', (event) => {
  const progress = JSON.parse(event.data);
  console.log(`${progress.table}: ${progress.percentage}%`);
});

eventSource.addEventListener('job_completed', (event) => {
  console.log('Job completed:', JSON.parse(event.data));
  eventSource.close();
});

eventSource.onerror = (error) => {
  console.error('SSE Error:', error);
  eventSource.close();
};
```

cURL:
```bash
curl -N http://localhost:8080/api/transports/TRPID-abc12345/status \
  -H "X-API-Key: your-api-key"
```

---

## 데이터 모델

### Transport

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | string | 고유 ID (TRPID-xxxxxxxx) |
| `name` | string | 이름 |
| `description` | string | 설명 |
| `tables` | string[] | 대상 테이블 목록 |
| `enabled` | boolean | 활성화 여부 |
| `schedule` | object | Cron 스케줄 설정 |
| `status` | string | 현재 상태 (idle/running/failed) |
| `created_at` | string | 생성 시간 (RFC3339) |
| `updated_at` | string | 수정 시간 (RFC3339) |

### Job

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | string | 고유 ID (JOB-YYYYMMDD-HHMMSS-xxxx) |
| `transport_id` | string | 연결된 Transport ID |
| `version` | integer | Transport별 버전 번호 |
| `status` | string | 상태 (pending/running/completed/failed/cancelled) |
| `started_at` | string | 시작 시간 |
| `completed_at` | string | 완료 시간 |
| `extractions` | array | 테이블별 추출 결과 |
| `error` | string | 에러 메시지 |
| `metrics` | object | 실행 메트릭 |
| `created_at` | string | 생성 시간 |

### Extraction

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | string | 추출 ID |
| `job_id` | string | 연결된 Job ID |
| `table_name` | string | 테이블 이름 |
| `status` | string | 상태 (pending/running/completed/failed) |
| `row_count` | integer | 처리된 row 수 |
| `byte_count` | integer | 전송된 바이트 수 |
| `gcs_path` | string | GCS 객체 경로 |
| `started_at` | string | 시작 시간 |
| `completed_at` | string | 완료 시간 |
| `error` | string | 에러 메시지 |

### JobMetrics

| 필드 | 타입 | 설명 |
|------|------|------|
| `total_rows` | integer | 총 처리 row 수 |
| `total_bytes` | integer | 총 바이트 수 |
| `duration` | integer | 실행 시간 (nanoseconds) |
| `rows_per_second` | float | 초당 처리 row 수 |

---

## Rate Limiting

Rate limiting이 활성화된 경우 분당 요청 수가 제한됩니다.

**제한 초과 응답** (429 Too Many Requests)

```json
{
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "요청 제한 초과. 잠시 후 다시 시도해주세요."
}
```

**설정 옵션**

| 항목 | 기본값 | 설명 |
|------|--------|------|
| `requests_per_minute` | 100 | 분당 최대 요청 수 |
| `burst_size` | 20 | 버스트 허용 크기 |
