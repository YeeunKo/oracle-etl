# SPEC-ETL-002: Acceptance Criteria

## Overview

Oracle ETL Pipeline Frontend UI의 인수 기준서. 각 기능 요구사항에 대한 Given-When-Then 형식의 테스트 시나리오를 정의합니다.

---

## Test Scenarios

### TC-DASH-001: Dashboard Connection Status

**Requirement**: FR-01.1, FR-01.2

#### Scenario: Oracle 연결 상태 표시

```gherkin
Feature: Dashboard Connection Status Display

  Scenario: Oracle 연결 성공 상태 표시
    Given 사용자가 대시보드 페이지에 접속한다
    And Oracle 데이터베이스가 정상 연결 상태이다
    When 페이지가 로드된다
    Then 연결 상태 표시기가 "Connected" 텍스트를 표시한다
    And 상태 표시기 색상이 녹색이다

  Scenario: Oracle 연결 실패 상태 표시
    Given 사용자가 대시보드 페이지에 접속한다
    And Oracle 데이터베이스 연결이 실패 상태이다
    When 페이지가 로드된다
    Then 연결 상태 표시기가 "Disconnected" 텍스트를 표시한다
    And 상태 표시기 색상이 적색이다

  Scenario: 연결 상태 자동 갱신
    Given 사용자가 대시보드 페이지에 있다
    And 연결 상태가 "Connected"로 표시되어 있다
    When 30초가 경과한다
    Then 시스템이 /api/health 엔드포인트를 호출한다
    And 연결 상태가 최신 정보로 갱신된다
```

**Acceptance Criteria (AC-01)**:
- [ ] 대시보드 상단에 연결 상태 표시기가 존재한다
- [ ] Connected/Disconnected/Checking 3가지 상태를 표시한다
- [ ] 30초 간격으로 자동 갱신된다
- [ ] 상태별 적절한 색상이 적용된다 (녹색/적색/노란색)

---

### TC-DASH-002: Recent Jobs Widget

**Requirement**: FR-01.3, FR-01.4

#### Scenario: 최근 Job 이력 표시

```gherkin
Feature: Dashboard Recent Jobs Widget

  Scenario: 최근 Job 5개 표시
    Given 시스템에 10개의 Job 실행 이력이 있다
    When 사용자가 대시보드 페이지에 접속한다
    Then 최근 5개의 Job이 테이블에 표시된다
    And 각 Job에 대해 ID, Transport 이름, 상태, 시작 시간이 표시된다

  Scenario: Job 이력이 없는 경우
    Given 시스템에 Job 실행 이력이 없다
    When 사용자가 대시보드 페이지에 접속한다
    Then "최근 실행된 Job이 없습니다" 메시지가 표시된다

  Scenario: 통계 정보 표시
    Given 시스템에 다음 데이터가 있다:
      | 총 Transport | 5 |
      | 오늘 실행 Job | 3 |
      | 실행 중 Job | 1 |
    When 사용자가 대시보드 페이지에 접속한다
    Then Quick Stats 위젯에 해당 통계가 표시된다
```

**Acceptance Criteria (AC-02)**:
- [ ] 최근 5개 Job이 최신순으로 표시된다
- [ ] Job ID, Transport 이름, 상태, 시작/종료 시간, 처리 row 수가 표시된다
- [ ] Quick Stats에 총 Transport 수, 오늘 Job 수, 실행 중 Job 수가 표시된다
- [ ] 빈 상태에 적절한 메시지가 표시된다

---

### TC-TRP-001: Transport List

**Requirement**: FR-02.1, FR-02.2, FR-02.3

#### Scenario: Transport 목록 조회

```gherkin
Feature: Transport List View

  Scenario: Transport 목록 페이지네이션
    Given 시스템에 25개의 Transport가 등록되어 있다
    When 사용자가 Transport 관리 페이지에 접속한다
    Then 첫 페이지에 20개의 Transport가 표시된다
    And 페이지네이션 컨트롤이 표시된다
    And 총 2페이지가 있음을 표시한다

  Scenario: Transport 이름으로 검색
    Given Transport 목록에 다음 항목이 있다:
      | ID | Name |
      | TRP-001 | Daily Sales Export |
      | TRP-002 | Monthly Report |
      | TRP-003 | Sales Summary |
    When 사용자가 검색창에 "Sales"를 입력한다
    Then "Daily Sales Export"와 "Sales Summary"가 표시된다
    And "Monthly Report"는 표시되지 않는다

  Scenario: Transport 상태 필터링
    Given Transport 목록에 idle 상태 3개, running 상태 1개가 있다
    When 사용자가 상태 필터에서 "running"을 선택한다
    Then running 상태의 Transport만 표시된다
```

**Acceptance Criteria (AC-03)**:
- [ ] Transport 목록이 테이블 형태로 표시된다
- [ ] ID, 이름, 테이블 수, 상태, 생성일, 최근 실행일이 표시된다
- [ ] 20개 이상일 때 페이지네이션이 작동한다
- [ ] 이름 검색과 상태 필터가 작동한다

---

### TC-TRP-002: Transport Creation

**Requirement**: FR-02.4, FR-02.5

#### Scenario: 새 Transport 생성

```gherkin
Feature: Transport Creation

  Scenario: Transport 생성 성공
    Given 사용자가 Transport 관리 페이지에 있다
    When "새 Transport 생성" 버튼을 클릭한다
    Then Transport 생성 폼이 표시된다
    When 다음 정보를 입력한다:
      | Field | Value |
      | 이름 | Daily Export |
      | 설명 | 일일 데이터 추출 |
    And 테이블 목록에서 "VBRP", "VBRK"를 선택한다
    And "생성" 버튼을 클릭한다
    Then "Transport가 생성되었습니다" 메시지가 표시된다
    And Transport 목록에 새 항목이 추가된다

  Scenario: 필수 필드 누락 시 유효성 검사
    Given Transport 생성 폼이 열려있다
    When 이름을 입력하지 않고 "생성" 버튼을 클릭한다
    Then "이름은 필수 입력 항목입니다" 에러 메시지가 표시된다
    And Transport가 생성되지 않는다

  Scenario: 테이블 선택 필수
    Given Transport 생성 폼이 열려있다
    And 이름이 입력되어 있다
    When 테이블을 선택하지 않고 "생성" 버튼을 클릭한다
    Then "최소 1개 이상의 테이블을 선택해야 합니다" 에러가 표시된다
```

**Acceptance Criteria (AC-04)**:
- [ ] 생성 폼에서 이름, 설명, 테이블 목록을 입력할 수 있다
- [ ] 사용 가능한 테이블 목록이 체크박스로 표시된다
- [ ] 필수 필드 유효성 검사가 작동한다
- [ ] 생성 성공 시 목록이 갱신된다

---

### TC-TRP-003: Transport Execution

**Requirement**: FR-02.6, FR-02.7, FR-02.8

#### Scenario: Transport 실행

```gherkin
Feature: Transport Execution

  Scenario: Transport 실행 시작
    Given Transport "TRP-001"이 idle 상태이다
    When 사용자가 "실행" 버튼을 클릭한다
    Then 실행 확인 다이얼로그가 표시된다
    When "확인" 버튼을 클릭한다
    Then "Job이 시작되었습니다" 메시지가 표시된다
    And 모니터링 페이지로 자동 이동된다

  Scenario: 실행 중 Transport 삭제 방지
    Given Transport "TRP-001"이 running 상태이다
    When 사용자가 Transport 목록을 조회한다
    Then "삭제" 버튼이 비활성화되어 있다
    And 버튼에 마우스를 올리면 "실행 중에는 삭제할 수 없습니다" 툴팁이 표시된다

  Scenario: Transport 삭제
    Given Transport "TRP-001"이 idle 상태이다
    When 사용자가 "삭제" 버튼을 클릭한다
    Then 삭제 확인 다이얼로그가 표시된다
    When "삭제" 버튼을 클릭한다
    Then "Transport가 삭제되었습니다" 메시지가 표시된다
    And Transport 목록에서 해당 항목이 제거된다
```

**Acceptance Criteria (AC-05)**:
- [ ] 실행 버튼 클릭 시 확인 다이얼로그가 표시된다
- [ ] 실행 시작 후 모니터링 페이지로 이동한다
- [ ] 실행 중인 Transport는 삭제할 수 없다
- [ ] 삭제 시 확인 다이얼로그가 표시된다

---

### TC-MON-001: SSE Connection

**Requirement**: FR-03.1, FR-03.6

#### Scenario: SSE 연결 및 재연결

```gherkin
Feature: SSE Real-time Connection

  Scenario: SSE 연결 성공
    Given Transport "TRP-001"이 실행 중이다
    When 사용자가 모니터링 페이지에 접속한다
    Then SSE 연결이 수립된다
    And 연결 상태가 "Connected"로 표시된다

  Scenario: SSE 연결 끊김 시 재연결
    Given SSE 연결이 활성화되어 있다
    When 네트워크 연결이 일시적으로 끊어진다
    Then "연결이 끊어졌습니다. 재연결 중..." 메시지가 표시된다
    And 5초 후 재연결을 시도한다
    When 네트워크가 복구된다
    Then SSE 연결이 재수립된다
    And 연결 상태가 "Connected"로 변경된다

  Scenario: SSE 재연결 3회 실패
    Given SSE 연결이 끊어진 상태이다
    When 재연결을 3회 시도하고 모두 실패한다
    Then "연결할 수 없습니다. 페이지를 새로고침 해주세요." 메시지가 표시된다
    And "새로고침" 버튼이 표시된다
```

**Acceptance Criteria (AC-06)**:
- [ ] 모니터링 페이지 진입 시 SSE 연결이 자동으로 수립된다
- [ ] 연결 상태가 UI에 표시된다
- [ ] 연결 끊김 시 5초 간격으로 최대 3회 재연결을 시도한다
- [ ] 재연결 실패 시 사용자에게 안내 메시지가 표시된다

---

### TC-MON-002: Real-time Progress

**Requirement**: FR-03.2, FR-03.3, FR-03.4, FR-03.5, FR-03.7

#### Scenario: 실시간 진행률 표시

```gherkin
Feature: Real-time Monitoring

  Scenario: 전체 진행률 표시
    Given Job이 실행 중이고 50% 완료되었다
    When 모니터링 페이지를 조회한다
    Then 전체 진행률 바가 50%로 표시된다
    And "50%" 텍스트가 표시된다

  Scenario: 테이블별 진행률 표시
    Given 다음 테이블이 추출 대상이다:
      | Table | Status | Progress |
      | VBRP | completed | 100% |
      | VBRK | running | 45% |
      | LIKP | pending | 0% |
    When 모니터링 페이지를 조회한다
    Then 각 테이블에 대해 개별 프로그레스 바가 표시된다
    And VBRP는 완료(녹색) 상태로 표시된다
    And VBRK는 진행 중(파란색) 상태로 45% 표시된다
    And LIKP는 대기(회색) 상태로 표시된다

  Scenario: 실시간 메트릭 표시
    Given Job이 실행 중이다
    When progress 이벤트가 수신된다:
      | rows_processed | 50000 |
      | rows_per_second | 10500 |
      | bytes_transferred | 15728640 |
    Then 다음 메트릭이 화면에 표시된다:
      | 처리된 rows | 50,000 |
      | 처리 속도 | 10,500 rows/s |
      | 전송 데이터 | 15 MB |

  Scenario: 에러 토스트 알림
    Given Job이 실행 중이다
    When error 이벤트가 수신된다:
      | table | VBRK |
      | message | Connection timeout |
    Then 화면 우상단에 에러 토스트가 표시된다
    And 토스트 내용에 "VBRK: Connection timeout"이 표시된다
    And 5초 후 토스트가 자동으로 닫힌다

  Scenario: Job 완료 요약
    Given Job이 실행 중이다
    When complete 이벤트가 수신된다:
      | total_rows | 750000 |
      | total_bytes | 157286400 |
      | duration_seconds | 7.5 |
    Then 완료 요약 모달이 표시된다
    And 다음 정보가 표시된다:
      | 총 처리 rows | 750,000 |
      | 전송 데이터 | 150 MB |
      | 소요 시간 | 7.5초 |
      | 평균 속도 | 100,000 rows/s |
```

**Acceptance Criteria (AC-07)**:
- [ ] 전체 진행률이 프로그레스 바로 표시된다
- [ ] 테이블별 개별 진행률이 표시된다
- [ ] 실시간 메트릭(rows, speed, bytes)이 표시된다
- [ ] 에러 발생 시 토스트 알림이 표시된다
- [ ] 완료 시 요약 모달이 표시된다

---

### TC-JOB-001: Job List

**Requirement**: FR-04.1, FR-04.2, FR-04.3

#### Scenario: Job 이력 조회

```gherkin
Feature: Job History List

  Scenario: Job 목록 조회
    Given 시스템에 50개의 Job 이력이 있다
    When 사용자가 Job 이력 페이지에 접속한다
    Then 첫 페이지에 20개의 Job이 최신순으로 표시된다
    And 각 Job에 대해 ID, Transport, 버전, 상태, 시간, row 수가 표시된다

  Scenario: Job 상태별 필터링
    Given Job 목록에 completed 30개, failed 5개가 있다
    When 사용자가 상태 필터에서 "failed"를 선택한다
    Then failed 상태의 Job 5개만 표시된다

  Scenario: Transport별 필터링
    Given Job 목록에 다양한 Transport의 Job이 있다
    When 사용자가 Transport 드롭다운에서 "TRP-001"을 선택한다
    Then TRP-001 Transport의 Job만 표시된다

  Scenario: 날짜 범위 필터링
    Given Job 목록에 최근 30일 간의 Job이 있다
    When 사용자가 시작일을 7일 전으로 설정한다
    Then 최근 7일 내 실행된 Job만 표시된다
```

**Acceptance Criteria (AC-08)**:
- [ ] Job 목록이 최신순으로 정렬되어 표시된다
- [ ] 페이지당 20개씩 페이지네이션된다
- [ ] 상태, Transport, 날짜 범위로 필터링할 수 있다
- [ ] 필터 조합이 올바르게 작동한다

---

### TC-JOB-002: Job Detail

**Requirement**: FR-04.4, FR-04.5, FR-04.6

#### Scenario: Job 상세 조회

```gherkin
Feature: Job Detail View

  Scenario: 성공한 Job 상세 보기
    Given Job "JOB-001"이 completed 상태이다
    When 사용자가 Job 행을 클릭한다
    Then Job 상세 페이지가 표시된다
    And Job 메타데이터(ID, Transport, 시간, 상태)가 표시된다
    And 테이블별 추출 결과가 아코디언으로 표시된다

  Scenario: 테이블 추출 상세 펼치기
    Given Job 상세 페이지에서 테이블 목록이 있다
    When 사용자가 "VBRP" 아코디언을 클릭한다
    Then 다음 정보가 펼쳐진다:
      | Row Count | 250,000 |
      | Bytes | 45 MB |
      | GCS Path | gs://bucket/TRP-001/v001/VBRP.jsonl.gz |
      | Duration | 2.4s |
    And GCS 경로 옆에 "복사" 버튼이 있다

  Scenario: 실패한 Job 에러 표시
    Given Job "JOB-002"가 failed 상태이다
    And 테이블 "VBRK" 추출 시 에러가 발생했다
    When 사용자가 Job 상세 페이지를 조회한다
    Then 에러 섹션이 강조 표시된다
    And 에러 메시지 "Connection timeout after 30s"가 표시된다
    And 실패한 테이블 "VBRK"가 적색으로 표시된다
```

**Acceptance Criteria (AC-09)**:
- [ ] Job 메타데이터가 표시된다
- [ ] 테이블별 추출 결과가 아코디언으로 표시된다
- [ ] GCS 경로 복사 기능이 작동한다
- [ ] 실패한 Job의 에러 정보가 강조 표시된다

---

### TC-PERF-001: Performance

**Requirement**: NFR-01.1, NFR-01.2

#### Scenario: 페이지 로드 성능

```gherkin
Feature: Page Load Performance

  Scenario: 대시보드 초기 로드
    Given 사용자가 인증된 상태이다
    When 대시보드 페이지에 접속한다
    Then 페이지가 2초 이내에 로드된다
    And LCP(Largest Contentful Paint)가 2.5초 이하이다

  Scenario: 로딩 상태 표시
    Given API 응답이 지연되고 있다
    When 페이지가 데이터를 로딩 중이다
    Then 스켈레톤 UI 또는 스피너가 표시된다
    And 사용자가 로딩 중임을 인지할 수 있다
```

**Acceptance Criteria (AC-10)**:
- [ ] 모든 페이지가 2초 이내에 로드된다
- [ ] LCP가 2.5초 이하이다
- [ ] 번들 크기가 500KB(gzip) 이하이다
- [ ] 로딩 중 적절한 피드백이 제공된다

---

### TC-UX-001: Responsive Design

**Requirement**: NFR-02.1, NFR-02.4

#### Scenario: 반응형 레이아웃

```gherkin
Feature: Responsive Design

  Scenario: 데스크톱 레이아웃
    Given 뷰포트 너비가 1440px이다
    When 대시보드 페이지를 조회한다
    Then 사이드바와 메인 콘텐츠가 나란히 표시된다
    And 테이블이 모든 컬럼을 표시한다

  Scenario: 태블릿 레이아웃
    Given 뷰포트 너비가 768px이다
    When 대시보드 페이지를 조회한다
    Then 사이드바가 접힌 상태이다
    And 햄버거 메뉴로 네비게이션에 접근한다

  Scenario: 빈 상태 표시
    Given Transport 목록이 비어있다
    When Transport 관리 페이지를 조회한다
    Then "등록된 Transport가 없습니다" 메시지가 표시된다
    And "새 Transport 생성" 버튼이 표시된다
```

**Acceptance Criteria (AC-11)**:
- [ ] 1280px 이상에서 최적화된 레이아웃이 제공된다
- [ ] 768px 이상에서 기본 기능이 사용 가능하다
- [ ] 빈 상태에 적절한 메시지와 액션이 제공된다
- [ ] 모든 비동기 작업에 로딩 상태가 표시된다

---

## Definition of Done

### Functional Completeness
- [ ] 모든 FR(Functional Requirements)이 구현됨
- [ ] 모든 TC(Test Case)가 통과함
- [ ] 에러 케이스가 처리됨

### Code Quality
- [ ] TypeScript strict 모드 통과
- [ ] ESLint/Prettier 규칙 준수
- [ ] 테스트 커버리지 85% 이상

### Documentation
- [ ] 컴포넌트 Props 문서화
- [ ] API 연동 가이드 작성
- [ ] README 업데이트

### Performance
- [ ] Lighthouse Performance 점수 90+
- [ ] 번들 크기 500KB 이하 (gzip)
- [ ] Core Web Vitals 기준 충족

### Accessibility
- [ ] WCAG 2.1 AA 준수
- [ ] 키보드 네비게이션 가능
- [ ] 스크린 리더 테스트 통과

---

## Verification Methods

| Method | Tool | Target |
|--------|------|--------|
| Unit Test | Vitest | 개별 컴포넌트, 훅, 유틸리티 |
| Integration Test | React Testing Library | 페이지 단위 기능 |
| E2E Test | Playwright | 사용자 시나리오 |
| Performance Test | Lighthouse | LCP, 번들 크기 |
| Accessibility Test | axe-core | WCAG 준수 |
| Visual Test | Storybook | 컴포넌트 스냅샷 |

---

## Traceability Matrix

| Requirement | Test Case | Acceptance Criteria | Status |
|-------------|-----------|---------------------|--------|
| FR-01.1 | TC-DASH-001 | AC-01 | Pending |
| FR-01.2 | TC-DASH-001 | AC-01 | Pending |
| FR-01.3 | TC-DASH-002 | AC-02 | Pending |
| FR-01.4 | TC-DASH-002 | AC-02 | Pending |
| FR-02.1 | TC-TRP-001 | AC-03 | Pending |
| FR-02.4 | TC-TRP-002 | AC-04 | Pending |
| FR-02.8 | TC-TRP-003 | AC-05 | Pending |
| FR-03.1 | TC-MON-001 | AC-06 | Pending |
| FR-03.4 | TC-MON-002 | AC-07 | Pending |
| FR-04.1 | TC-JOB-001 | AC-08 | Pending |
| FR-04.4 | TC-JOB-002 | AC-09 | Pending |
| NFR-01.1 | TC-PERF-001 | AC-10 | Pending |
| NFR-02.1 | TC-UX-001 | AC-11 | Pending |
