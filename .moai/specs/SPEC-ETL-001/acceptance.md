# SPEC-ETL-001: Acceptance Criteria

## Metadata

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-ETL-001 |
| Document Type | Acceptance Criteria |
| Created | 2026-01-18 |
| Status | Planned |

---

## Acceptance Criteria Summary

| ID | Criteria | Priority | Status |
|----|----------|----------|--------|
| AC-01 | Oracle mTLS 연결 수립 및 풀 관리 | High | Pending |
| AC-02 | 배치 데이터 추출 및 청크 스트리밍 | High | Pending |
| AC-03 | GCS 스트리밍 업로드 (NDJSON + Gzip) | High | Pending |
| AC-04 | Transport/Job 생성 및 실행 | High | Pending |
| AC-05 | SSE 실시간 진행률 모니터링 | Medium | Pending |
| AC-06 | 100,000 rows/second 성능 달성 | High | Pending |
| AC-07 | 자동 재시도 및 에러 복구 | Medium | Pending |
| AC-08 | API 인증 (API Key / Bearer Token) | Medium | Pending |

---

## Detailed Acceptance Criteria

### AC-01: Oracle Connection Management

#### Criteria
Oracle ATP 또는 On-Premise 데이터베이스에 mTLS wallet을 사용하여 보안 연결을 수립하고, 커넥션 풀을 효율적으로 관리한다.

#### Test Scenarios

**TC-CONN-001: mTLS Wallet 연결 성공**
```gherkin
Given wallet 파일이 /opt/wallet 경로에 존재하고
  And 환경 변수에 ORACLE_USER, ORACLE_PASSWORD가 설정되어 있을 때
When 서버가 시작되면
Then Oracle 커넥션 풀이 성공적으로 초기화되고
  And 최소 5개의 연결이 풀에 유지된다
```

**TC-CONN-002: 연결 상태 확인 엔드포인트**
```gherkin
Given 서버가 Oracle에 연결된 상태에서
When GET /api/oracle/status 요청을 보내면
Then 응답 코드는 200이고
  And 응답 본문에 connection_status: "connected"가 포함되고
  And pool_size, active_connections 정보가 포함된다
```

**TC-CONN-003: 연결 실패 시 에러 처리**
```gherkin
Given 잘못된 wallet 경로가 설정되어 있을 때
When 서버가 시작되면
Then 명확한 에러 메시지가 로그에 기록되고
  And 서버는 정상적으로 종료된다
```

**TC-CONN-004: 커넥션 풀 자동 복구**
```gherkin
Given Oracle 연결이 일시적으로 끊어졌을 때
When 새로운 요청이 들어오면
Then 커넥션 풀이 자동으로 재연결을 시도하고
  And 최대 3회 재시도 후 성공 또는 에러를 반환한다
```

#### Verification Method
- 단위 테스트: 커넥션 풀 모킹 테스트
- 통합 테스트: 실제 Oracle ATP 연결 테스트
- 수동 테스트: wallet 경로 변경 후 에러 확인

---

### AC-02: Data Extraction

#### Criteria
Oracle 데이터베이스에서 FetchArraySize=1000으로 배치 페치를 수행하고, 10,000 rows 단위로 청크 스트리밍을 수행한다.

#### Test Scenarios

**TC-EXTRACT-001: 배치 페치 설정 확인**
```gherkin
Given Oracle 연결이 수립된 상태에서
When 테이블 데이터 추출을 시작하면
Then FetchArraySize=1000 설정이 적용되고
  And 1000개 단위로 rows가 페치된다
```

**TC-EXTRACT-002: 테이블 목록 조회**
```gherkin
Given Oracle 연결이 수립된 상태에서
When GET /api/tables 요청을 보내면
Then 접근 가능한 모든 테이블 목록이 반환되고
  And 각 테이블의 row count가 포함된다
```

**TC-EXTRACT-003: 샘플 데이터 조회**
```gherkin
Given VBRP 테이블이 존재할 때
When GET /api/tables/VBRP/sample 요청을 보내면
Then 응답 코드는 200이고
  And 최대 100개의 row가 JSON 배열로 반환되고
  And 각 row에 모든 컬럼이 포함된다
```

**TC-CHUNK-001: 청크 스트리밍**
```gherkin
Given 250,000 rows가 있는 테이블에서
When 데이터 추출을 시작하면
Then 10,000 rows 단위로 25개의 청크가 생성되고
  And 각 청크는 독립적으로 처리된다
```

**TC-EXTRACT-004: 대용량 테이블 추출 (274 컬럼)**
```gherkin
Given VBRP 테이블 (274 컬럼, 250,000 rows)이 존재할 때
When 전체 테이블 추출을 실행하면
Then 모든 컬럼 데이터가 손실 없이 추출되고
  And 메모리 사용량이 2GB를 초과하지 않는다
```

#### Verification Method
- 단위 테스트: 청크 분할 로직 테스트
- 통합 테스트: 실제 테이블 데이터 추출 테스트
- 성능 테스트: 대용량 테이블 추출 시간 측정

---

### AC-03: GCS Upload

#### Criteria
추출된 데이터를 JSON Lines 형식으로 직렬화하고, gzip 압축을 적용하여 Google Cloud Storage에 스트리밍 업로드한다.

#### Test Scenarios

**TC-UPLOAD-001: NDJSON 형식 검증**
```gherkin
Given 데이터 청크가 준비되었을 때
When GCS에 업로드하면
Then 파일 형식이 JSON Lines (한 줄에 하나의 JSON 객체)이고
  And 각 줄이 유효한 JSON 객체이다
```

**TC-UPLOAD-002: Gzip 압축 적용**
```gherkin
Given NDJSON 데이터가 준비되었을 때
When GCS에 업로드하면
Then 파일이 gzip으로 압축되고
  And Content-Encoding이 gzip으로 설정된다
```

**TC-UPLOAD-003: 스트리밍 업로드**
```gherkin
Given 250,000 rows 테이블 추출 중일 때
When GCS 업로드를 시작하면
Then 전체 데이터를 메모리에 적재하지 않고 스트리밍하고
  And 업로드 중 메모리 사용량이 500MB를 초과하지 않는다
```

**TC-UPLOAD-004: 파일 명명 규칙**
```gherkin
Given Transport ID가 TRP-001이고 Job Version이 v001일 때
When VBRP 테이블을 업로드하면
Then GCS 경로는 "etl-bucket/TRP-001/v001/VBRP.jsonl.gz"이다
```

**TC-UPLOAD-005: Resumable Upload 복구**
```gherkin
Given 업로드 중 네트워크 오류가 발생했을 때
When 연결이 복구되면
Then 업로드가 중단된 지점부터 재개되고
  And 데이터 손실 없이 완료된다
```

**TC-UPLOAD-006: 서비스 계정 인증**
```gherkin
Given 유효한 서비스 계정 JSON 키가 설정되어 있을 때
When GCS 클라이언트가 초기화되면
Then 서비스 계정으로 인증이 성공하고
  And 버킷에 쓰기 권한이 확인된다
```

#### Verification Method
- 단위 테스트: JSONL 인코더, Gzip 압축 테스트
- 통합 테스트: 실제 GCS 버킷 업로드 테스트
- 수동 테스트: 업로드된 파일 gunzip 및 내용 확인

---

### AC-04: Transport Management

#### Criteria
Transport를 생성하고, 포함된 테이블들에 대해 ETL 작업을 실행하며, Job 버전을 관리한다.

#### Test Scenarios

**TC-TRANSPORT-001: Transport 생성**
```gherkin
Given 인증된 사용자가
When POST /api/transport 요청을 다음 본문으로 보내면
  | name | SAP Sales Export |
  | tables | ["VBRP", "VBRK"] |
  | target_bucket | oracle-etl-data |
Then 응답 코드는 201이고
  And Transport ID가 "TRP-YYYYMMDD-NNN" 형식으로 생성되고
  And status는 "created"이다
```

**TC-TRANSPORT-002: Transport 조회**
```gherkin
Given Transport TRP-001이 존재할 때
When GET /api/transport/TRP-001 요청을 보내면
Then 응답 코드는 200이고
  And Transport 상세 정보가 반환된다
```

**TC-TRANSPORT-003: Transport 실행**
```gherkin
Given Transport TRP-001이 생성된 상태에서
When POST /api/transport/TRP-001/execute 요청을 보내면
Then 응답 코드는 202 (Accepted)이고
  And 새로운 Job이 생성되고 (JOBVER v001)
  And 모든 테이블에 대해 ETL 작업이 시작된다
```

**TC-TRANSPORT-004: Job 버전 증가**
```gherkin
Given Transport TRP-001이 v001로 실행 완료된 상태에서
When 다시 POST /api/transport/TRP-001/execute 요청을 보내면
Then 새로운 Job이 JOBVER v002로 생성되고
  And 이전 Job 기록이 유지된다
```

**TC-TRANSPORT-005: Job 이력 조회**
```gherkin
Given 여러 Job이 실행된 상태에서
When GET /api/jobs 요청을 보내면
Then 최신순으로 Job 목록이 반환되고
  And 각 Job에 transport_id, version, status, total_rows가 포함된다
```

**TC-TRANSPORT-006: 존재하지 않는 Transport**
```gherkin
Given Transport TRP-999가 존재하지 않을 때
When GET /api/transport/TRP-999 요청을 보내면
Then 응답 코드는 404이고
  And 에러 메시지가 반환된다
```

#### Verification Method
- 단위 테스트: Transport/Job 도메인 로직 테스트
- 통합 테스트: 전체 Transport 실행 플로우 테스트
- E2E 테스트: API를 통한 전체 워크플로우 테스트

---

### AC-05: Real-time Monitoring

#### Criteria
Transport 실행 중 SSE를 통해 실시간 진행률, 메트릭, 에러 이벤트를 스트리밍한다.

#### Test Scenarios

**TC-SSE-001: SSE 연결 수립**
```gherkin
Given Transport TRP-001이 실행 중일 때
When GET /api/transport/TRP-001/status 요청을 보내면
Then Content-Type이 "text/event-stream"이고
  And SSE 연결이 수립된다
```

**TC-SSE-002: 진행률 이벤트 수신**
```gherkin
Given SSE 연결이 수립된 상태에서
When 테이블 추출이 진행되면
Then "progress" 이벤트가 주기적으로 발송되고
  And 이벤트에 rows_processed, rows_per_second, progress_percent가 포함된다
```

**TC-SSE-003: 테이블 완료 이벤트**
```gherkin
Given VBRP 테이블 추출이 완료되었을 때
When SSE 채널에서
Then "table_complete" 이벤트가 발송되고
  And 이벤트에 table, total_rows, bytes_transferred, duration_seconds가 포함된다
```

**TC-SSE-004: 전체 완료 이벤트**
```gherkin
Given 모든 테이블 추출이 완료되었을 때
When SSE 채널에서
Then "complete" 이벤트가 발송되고
  And 이벤트에 transport_id, job_version, total_rows, total_duration_seconds가 포함되고
  And SSE 연결이 종료된다
```

**TC-SSE-005: 에러 이벤트 발송**
```gherkin
Given 테이블 추출 중 오류가 발생했을 때
When SSE 채널에서
Then "error" 이벤트가 즉시 발송되고
  And 이벤트에 error_code, message, table, details가 포함된다
```

**TC-SSE-006: 다중 클라이언트 지원**
```gherkin
Given 2개의 SSE 클라이언트가 동일한 Transport를 구독할 때
When 진행률 이벤트가 발생하면
Then 두 클라이언트 모두 동일한 이벤트를 수신한다
```

#### Verification Method
- 단위 테스트: SSE Broadcaster 로직 테스트
- 통합 테스트: 실제 SSE 이벤트 수신 테스트
- 수동 테스트: 브라우저 EventSource로 실시간 확인

---

### AC-06: Performance Target

#### Criteria
100,000 rows/second 이상의 처리량을 달성하고, API 응답 시간을 100ms 이하로 유지한다.

#### Test Scenarios

**TC-PERF-001: 처리량 벤치마크**
```gherkin
Given 250,000 rows 테이블 (274 컬럼)이 준비되었을 때
When ETL 작업을 실행하면
Then 평균 처리량이 100,000 rows/second 이상이고
  And 전체 소요 시간이 2.5초 이하이다
```

**TC-PERF-002: API 응답 시간**
```gherkin
Given 서버가 idle 상태에서
When GET /api/health 요청을 보내면
Then 응답 시간이 100ms 이하이다
```

**TC-PERF-003: 대용량 처리 (10M rows)**
```gherkin
Given 40개 테이블, 각 250,000 rows (총 10M rows)가 준비되었을 때
When 전체 Transport를 실행하면
Then 총 소요 시간이 100초 이하이고
  And 메모리 사용량이 2GB를 초과하지 않는다
```

**TC-PERF-004: 병렬 테이블 처리**
```gherkin
Given 4개의 테이블이 Transport에 포함되어 있을 때
When parallel_tables=4 설정으로 실행하면
Then 4개 테이블이 동시에 처리되고
  And 순차 처리 대비 3배 이상 빠르다
```

**TC-PERF-005: 메모리 효율성**
```gherkin
Given ETL 작업이 진행 중일 때
When 메모리 사용량을 모니터링하면
Then 피크 메모리가 2GB를 초과하지 않고
  And 메모리 누수가 발생하지 않는다
```

#### Verification Method
- 벤치마크 테스트: Go benchmark 도구 사용
- 성능 프로파일링: pprof 메모리/CPU 분석
- 부하 테스트: 다양한 데이터 크기에서 측정

---

### AC-07: Error Handling & Retry

#### Criteria
일시적 오류에 대해 자동 재시도를 수행하고, 모든 오류에 대해 구조화된 응답을 반환한다.

#### Test Scenarios

**TC-RETRY-001: 네트워크 오류 재시도**
```gherkin
Given GCS 업로드 중 일시적 네트워크 오류가 발생했을 때
When 재시도 로직이 동작하면
Then 최대 3회까지 exponential backoff로 재시도하고
  And 성공 시 작업이 정상 완료된다
```

**TC-RETRY-002: 재시도 실패 시 에러 반환**
```gherkin
Given 3회 재시도 후에도 오류가 지속될 때
When 재시도가 종료되면
Then 구조화된 에러 응답이 반환되고
  And 에러 로그에 모든 재시도 기록이 포함된다
```

**TC-RETRY-003: 구조화된 에러 응답**
```gherkin
Given 잘못된 테이블 이름으로 요청할 때
When GET /api/tables/INVALID/sample 요청을 보내면
Then 응답 코드는 404이고
  And 응답에 code, message, trace_id가 포함된다
```

**TC-RETRY-004: Graceful Shutdown**
```gherkin
Given ETL 작업이 진행 중일 때
When SIGTERM 시그널을 보내면
Then 현재 청크 처리가 완료되고
  And 리소스가 정리된 후 서버가 종료된다
```

**TC-RETRY-005: Oracle 연결 복구**
```gherkin
Given Oracle 연결이 끊어진 상태에서
When 새로운 요청이 들어오면
Then 커넥션 풀이 재연결을 시도하고
  And 성공 시 요청이 정상 처리된다
```

#### Verification Method
- 단위 테스트: 재시도 로직 모킹 테스트
- 통합 테스트: 네트워크 오류 시뮬레이션
- 수동 테스트: SIGTERM 시그널 테스트

---

### AC-08: API Authentication

#### Criteria
모든 보호된 API 엔드포인트에 API Key 또는 Bearer Token 인증을 요구한다.

#### Test Scenarios

**TC-AUTH-001: API Key 인증 성공**
```gherkin
Given 유효한 API Key가 설정되어 있을 때
When X-API-Key 헤더와 함께 GET /api/tables 요청을 보내면
Then 응답 코드는 200이고
  And 테이블 목록이 반환된다
```

**TC-AUTH-002: Bearer Token 인증 성공**
```gherkin
Given 유효한 Bearer Token이 설정되어 있을 때
When Authorization: Bearer <token> 헤더와 함께 요청을 보내면
Then 응답 코드는 200이다
```

**TC-AUTH-003: 인증 없이 요청**
```gherkin
Given 인증 헤더 없이
When GET /api/tables 요청을 보내면
Then 응답 코드는 401이고
  And 에러 메시지가 반환된다
```

**TC-AUTH-004: 잘못된 API Key**
```gherkin
Given 잘못된 API Key로
When GET /api/tables 요청을 보내면
Then 응답 코드는 401이고
  And 에러 메시지가 반환된다
```

**TC-AUTH-005: Health 엔드포인트 인증 불필요**
```gherkin
Given 인증 헤더 없이
When GET /api/health 요청을 보내면
Then 응답 코드는 200이고
  And 상태 정보가 반환된다
```

**TC-AUTH-006: 민감 정보 로깅 방지**
```gherkin
Given 로깅이 활성화된 상태에서
When API Key가 포함된 요청을 처리하면
Then 로그에 API Key 값이 마스킹되거나 제외된다
```

#### Verification Method
- 단위 테스트: 인증 미들웨어 테스트
- 통합 테스트: 다양한 인증 시나리오 테스트
- 보안 테스트: 로그 파일 민감 정보 검사

---

## Quality Gates

### Code Quality
- [ ] golangci-lint 경고 0개
- [ ] gofmt 적용 완료
- [ ] 코드 커버리지 85% 이상

### Security
- [ ] gosec 스캔 통과
- [ ] 민감 정보 로깅 방지 확인
- [ ] mTLS/인증 정상 동작

### Performance
- [ ] 100,000 rows/second 처리량 달성
- [ ] API 응답 시간 100ms 이하
- [ ] 메모리 사용량 2GB 이하

### Documentation
- [ ] API 문서 완성
- [ ] README 업데이트
- [ ] 배포 가이드 작성

---

## Definition of Done

SPEC-ETL-001은 다음 모든 조건이 충족될 때 완료로 간주됩니다:

1. **모든 AC 통과**: AC-01 ~ AC-08의 모든 테스트 시나리오가 통과
2. **품질 게이트 충족**: 위 Quality Gates의 모든 항목 체크
3. **코드 리뷰 완료**: 팀 리뷰 및 승인
4. **문서화 완성**: API 문서, 배포 가이드 완성
5. **성능 검증**: 벤치마크 결과 목표 달성 확인

---

## Traceability Matrix

| AC ID | Requirement IDs | Test Scenarios | Milestone |
|-------|-----------------|----------------|-----------|
| AC-01 | FR-01.1, FR-01.2, FR-01.3 | TC-CONN-001~004 | M2 |
| AC-02 | FR-02.1, FR-02.2, FR-02.3, FR-02.4 | TC-EXTRACT-001~004, TC-CHUNK-001 | M2 |
| AC-03 | FR-03.1, FR-03.2, FR-03.3, FR-03.4 | TC-UPLOAD-001~006 | M3 |
| AC-04 | FR-04.1, FR-04.2, FR-04.3, FR-04.4, FR-05.5, FR-05.6 | TC-TRANSPORT-001~006 | M4 |
| AC-05 | FR-06.1, FR-06.2, FR-06.3 | TC-SSE-001~006 | M5 |
| AC-06 | NFR-01.1, NFR-01.2, NFR-01.3, NFR-01.4 | TC-PERF-001~005 | M8 |
| AC-07 | NFR-02.1, NFR-02.2, NFR-02.3, NFR-02.4 | TC-RETRY-001~005 | M7 |
| AC-08 | NFR-03.3, NFR-03.4 | TC-AUTH-001~006 | M6 |
