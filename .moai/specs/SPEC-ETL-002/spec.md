# SPEC-ETL-002: Oracle ETL Pipeline Frontend UI

## Metadata

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-ETL-002 |
| Title | Oracle ETL Pipeline Frontend Dashboard and Management UI |
| Created | 2026-01-18 |
| Status | Planned |
| Priority | High |
| Lifecycle | spec-anchored |
| Version | 1.0.0 |
| Related SPEC | SPEC-ETL-001 (Backend API) |

## Executive Summary

Oracle ETL Pipeline의 프론트엔드 대시보드 및 관리 UI 개발. SPEC-ETL-001에서 정의된 백엔드 API를 활용하여 Transport 관리, 실시간 모니터링, Job 이력 조회 기능을 제공하는 웹 기반 사용자 인터페이스를 구현합니다.

---

## Environment

### Target Platform
- **Browser**: Chrome 120+, Firefox 120+, Safari 17+, Edge 120+
- **Viewport**: Desktop (1280px+), Tablet (768px+) 반응형 지원
- **Language**: 한국어 기본, 영어 지원 (i18n)

### Technology Stack
- **Framework**: React 19 with TypeScript 5.9+
- **Build Tool**: Vite 6.0+
- **Styling**: Tailwind CSS 4.0+ with shadcn/ui components
- **State Management**: TanStack Query v5 (서버 상태), Zustand (클라이언트 상태)
- **Routing**: React Router v7
- **Real-time**: EventSource API (SSE)
- **Form Handling**: React Hook Form + Zod validation
- **Charts**: Recharts 또는 Chart.js for metrics visualization

### Backend Integration
- **API Base URL**: `/api` (SPEC-ETL-001 참조)
- **Authentication**: API Key 또는 Bearer Token
- **Real-time**: SSE endpoint `/api/transports/:id/status`

---

## Assumptions

### Technical Assumptions

| ID | Assumption | Confidence | Evidence | Risk if Wrong | Validation Method |
|----|------------|------------|----------|---------------|-------------------|
| A-01 | SPEC-ETL-001 백엔드 API 구현 완료 | High | SPEC 완료 상태 | 프론트엔드 개발 블록 | API 엔드포인트 테스트 |
| A-02 | SSE 연결 안정성 확보 | Medium | 브라우저 표준 지원 | 실시간 업데이트 불가 | 장시간 연결 테스트 |
| A-03 | 동시 접속자 10명 이하 | High | 관리자 전용 도구 | UI 성능 저하 | 부하 테스트 |
| A-04 | CORS 설정 완료 | High | 개발 서버 설정 | API 호출 실패 | 크로스 오리진 테스트 |

### Business Assumptions

| ID | Assumption | Confidence | Risk if Wrong |
|----|------------|------------|---------------|
| B-01 | 관리자 전용 인터페이스 | High | 권한 관리 복잡성 증가 |
| B-02 | 단일 Transport 동시 실행 제한 | High | UI 상태 관리 복잡성 |
| B-03 | 실시간 모니터링 필수 기능 | High | 사용자 경험 저하 |

---

## Requirements

### Functional Requirements

#### FR-01: Dashboard

**FR-01.1: Oracle Connection Status Display** (Ubiquitous)
- 시스템은 **항상** 대시보드 상단에 Oracle 연결 상태를 표시해야 한다
- 상태 표시: Connected (녹색), Disconnected (적색), Checking (노란색)
- `/api/health` 및 `/api/oracle/status` 엔드포인트 활용

**FR-01.2: Auto Status Refresh** (Event-Driven)
- **WHEN** 대시보드 페이지 로드 시
- **THEN** 30초 간격으로 연결 상태를 자동 갱신한다

**FR-01.3: Recent Job History Widget** (Ubiquitous)
- 시스템은 **항상** 대시보드에 최근 5개의 Job 실행 이력을 표시해야 한다
- 표시 정보: Job ID, Transport 이름, 상태, 시작/종료 시간, 처리 row 수

**FR-01.4: Quick Stats Summary** (Ubiquitous)
- 시스템은 **항상** 다음 통계 정보를 대시보드에 표시해야 한다
  - 총 Transport 수
  - 오늘 실행된 Job 수
  - 현재 실행 중인 Job 수
  - 마지막 성공 Job 시간

**FR-01.5: Running Job Alert** (State-Driven)
- **IF** 현재 실행 중인 Job이 있으면
- **THEN** 대시보드 상단에 실행 중 배너를 표시하고 실시간 진행률을 보여준다

#### FR-02: Transport Management

**FR-02.1: Transport List View** (Event-Driven)
- **WHEN** Transport 관리 페이지 접근 시
- **THEN** 모든 Transport 목록을 테이블 형태로 표시한다
- 표시 컬럼: ID, 이름, 테이블 수, 상태, 생성일, 최근 실행일, 액션 버튼

**FR-02.2: Transport List Pagination** (Event-Driven)
- **WHEN** Transport가 20개 이상일 때
- **THEN** 페이지당 20개씩 페이지네이션을 제공한다

**FR-02.3: Transport Search and Filter** (Event-Driven)
- **WHEN** 사용자가 검색어 입력 시
- **THEN** Transport 이름으로 필터링한다
- **WHEN** 사용자가 상태 필터 선택 시
- **THEN** 해당 상태의 Transport만 표시한다

**FR-02.4: Transport Creation Form** (Event-Driven)
- **WHEN** "새 Transport 생성" 버튼 클릭 시
- **THEN** Transport 생성 모달/페이지를 표시한다
- 필수 입력: 이름, 대상 테이블 목록
- 선택 입력: 설명

**FR-02.5: Table Selection UI** (Event-Driven)
- **WHEN** Transport 생성/수정 시 테이블 선택 필요할 때
- **THEN** 사용 가능한 테이블 목록을 체크박스 리스트로 표시한다
- `/api/tables` 엔드포인트에서 테이블 목록 조회

**FR-02.6: Transport Deletion** (Event-Driven)
- **WHEN** 삭제 버튼 클릭 시
- **THEN** 확인 다이얼로그를 표시하고 확인 후 삭제를 실행한다
- 삭제 실패 시 에러 메시지 표시

**FR-02.7: Transport Deletion Guard** (State-Driven)
- **IF** Transport가 현재 실행 중이면
- **THEN** 삭제 버튼을 비활성화하고 "실행 중" 툴팁을 표시한다

**FR-02.8: Transport Execution Trigger** (Event-Driven)
- **WHEN** "실행" 버튼 클릭 시
- **THEN** 확인 다이얼로그 표시 후 Transport 실행을 시작한다
- 실행 시작 후 자동으로 모니터링 페이지로 이동

**FR-02.9: Transport Detail View** (Event-Driven)
- **WHEN** Transport 행 클릭 또는 상세 버튼 클릭 시
- **THEN** Transport 상세 정보 페이지를 표시한다
- 표시 정보: 기본 정보, 테이블 목록, 실행 이력, 통계

#### FR-03: Real-time Monitoring

**FR-03.1: SSE Connection Establishment** (Event-Driven)
- **WHEN** Transport 실행이 시작되거나 모니터링 페이지 접근 시
- **THEN** SSE 연결을 수립하고 실시간 이벤트를 수신한다

**FR-03.2: Overall Progress Display** (State-Driven)
- **IF** Job이 실행 중이면
- **THEN** 전체 진행률을 프로그레스 바와 퍼센트로 표시한다

**FR-03.3: Table-by-Table Progress** (State-Driven)
- **IF** Transport 실행 중이면
- **THEN** 각 테이블별 추출 상태를 개별 프로그레스 바로 표시한다
- 상태 표시: Pending (대기), Running (진행 중), Completed (완료), Failed (실패)

**FR-03.4: Live Metrics Display** (Ubiquitous)
- 시스템은 **항상** 실행 중 다음 메트릭을 실시간으로 표시해야 한다
  - `rows_processed`: 처리된 총 row 수
  - `rows_per_second`: 현재 처리 속도
  - `bytes_transferred`: 전송된 바이트 (MB/GB 단위)
  - `current_table`: 현재 처리 중인 테이블명
  - `elapsed_time`: 경과 시간

**FR-03.5: Error Notification Toast** (Event-Driven)
- **WHEN** SSE를 통해 에러 이벤트 수신 시
- **THEN** 화면 우상단에 에러 토스트 알림을 표시한다
- 토스트는 5초 후 자동 닫힘, 수동 닫기 가능

**FR-03.6: SSE Reconnection** (State-Driven)
- **IF** SSE 연결이 끊어지면
- **THEN** 5초 후 자동 재연결을 시도하고 최대 3회 재시도한다
- 재연결 시도 중 상태 표시

**FR-03.7: Completion Summary** (Event-Driven)
- **WHEN** Job 완료 이벤트 수신 시
- **THEN** 완료 요약 모달을 표시한다
- 표시 정보: 총 처리 row 수, 총 전송 바이트, 소요 시간, 처리 속도

**FR-03.8: Monitoring Page Persistence** (State-Driven)
- **IF** 사용자가 모니터링 페이지를 벗어났다가 돌아오면
- **THEN** 현재 실행 중인 Job의 상태를 복원하고 SSE 연결을 재수립한다

#### FR-04: Job History

**FR-04.1: Job List View** (Event-Driven)
- **WHEN** Job 이력 페이지 접근 시
- **THEN** 모든 Job 목록을 테이블 형태로 최신순 정렬하여 표시한다
- 표시 컬럼: Job ID, Transport 이름, 버전, 상태, 시작/종료 시간, 처리 row 수, 소요 시간

**FR-04.2: Job List Pagination** (Ubiquitous)
- 시스템은 **항상** Job 목록을 페이지당 20개씩 페이지네이션하여 제공해야 한다
- Offset-based 페이지네이션 사용

**FR-04.3: Job List Filtering** (Event-Driven)
- **WHEN** 사용자가 필터 조건 선택 시
- **THEN** 해당 조건으로 Job 목록을 필터링한다
- 필터 옵션: Transport ID, 상태 (pending/running/completed/failed), 날짜 범위

**FR-04.4: Job Detail View** (Event-Driven)
- **WHEN** Job 행 클릭 시
- **THEN** Job 상세 정보를 표시한다
- 표시 정보: Job 메타데이터, 테이블별 추출 결과, 메트릭, 에러 로그 (있는 경우)

**FR-04.5: Table Extraction Details** (Event-Driven)
- **WHEN** Job 상세 페이지에서 테이블 목록 표시 시
- **THEN** 각 테이블별 추출 결과를 확장 가능한 아코디언으로 표시한다
- 표시 정보: 테이블명, row 수, 바이트 수, GCS 경로, 시작/종료 시간, 상태

**FR-04.6: Job Error Display** (State-Driven)
- **IF** Job이 실패 상태이면
- **THEN** 에러 메시지와 실패한 테이블 정보를 강조 표시한다

---

### Non-Functional Requirements

#### NFR-01: Performance

**NFR-01.1: Initial Load Time** (Ubiquitous)
- 시스템은 **항상** 페이지 초기 로드 시간 2초 이하를 유지해야 한다
- LCP (Largest Contentful Paint) 2.5초 이하

**NFR-01.2: API Response Handling** (Ubiquitous)
- 시스템은 **항상** API 호출 중 로딩 상태를 표시해야 한다
- Skeleton UI 또는 Spinner 사용

**NFR-01.3: SSE Memory Management** (Ubiquitous)
- 시스템은 **항상** SSE 연결 시 메모리 누수를 방지해야 한다
- 컴포넌트 언마운트 시 EventSource 정리

**NFR-01.4: List Virtualization** (State-Driven)
- **IF** 목록 아이템이 100개 이상이면
- **THEN** 가상 스크롤링을 적용하여 렌더링 성능을 최적화한다

#### NFR-02: Usability

**NFR-02.1: Responsive Design** (Ubiquitous)
- 시스템은 **항상** 1280px 이상 데스크톱에서 최적화된 레이아웃을 제공해야 한다
- 768px 이상 태블릿에서 기본 기능 사용 가능

**NFR-02.2: Loading States** (Ubiquitous)
- 시스템은 **항상** 모든 비동기 작업에 대해 로딩 상태를 표시해야 한다

**NFR-02.3: Error States** (Ubiquitous)
- 시스템은 **항상** 에러 발생 시 사용자 친화적 메시지를 표시해야 한다
- 기술적 에러 코드는 상세 보기에서만 표시

**NFR-02.4: Empty States** (Ubiquitous)
- 시스템은 **항상** 데이터가 없는 경우 빈 상태 UI를 표시해야 한다
- 예: "등록된 Transport가 없습니다. 새 Transport를 생성해보세요."

**NFR-02.5: Keyboard Navigation** (Optional)
- **가능하면** 주요 기능에 대해 키보드 단축키를 제공한다
- 예: `Ctrl+N` 새 Transport 생성, `Ctrl+R` 새로고침

#### NFR-03: Reliability

**NFR-03.1: Error Boundary** (Ubiquitous)
- 시스템은 **항상** React Error Boundary를 사용하여 컴포넌트 오류를 격리해야 한다
- 오류 발생 시 대체 UI 표시 및 복구 옵션 제공

**NFR-03.2: Offline Detection** (Event-Driven)
- **WHEN** 네트워크 연결이 끊어지면
- **THEN** 오프라인 상태 배너를 표시한다

**NFR-03.3: Session Persistence** (State-Driven)
- **IF** 브라우저가 새로고침되면
- **THEN** 인증 상태와 마지막 페이지 위치를 복원한다

#### NFR-04: Security

**NFR-04.1: Token Storage** (Ubiquitous)
- 시스템은 **항상** 인증 토큰을 httpOnly 쿠키 또는 메모리에 저장해야 한다
- localStorage에 민감 정보 저장 금지

**NFR-04.2: XSS Prevention** (Unwanted)
- 시스템은 사용자 입력을 직접 HTML에 삽입**하지 않아야 한다**
- React의 기본 이스케이프 활용

**NFR-04.3: API Key Protection** (Unwanted)
- 시스템은 API Key를 클라이언트 코드에 하드코딩**하지 않아야 한다**
- 환경 변수 또는 서버 사이드 프록시 사용

#### NFR-05: Accessibility

**NFR-05.1: ARIA Labels** (Ubiquitous)
- 시스템은 **항상** 인터랙티브 요소에 적절한 ARIA 레이블을 제공해야 한다

**NFR-05.2: Color Contrast** (Ubiquitous)
- 시스템은 **항상** WCAG 2.1 AA 수준의 색상 대비를 유지해야 한다

**NFR-05.3: Focus Management** (Ubiquitous)
- 시스템은 **항상** 모달 열림 시 포커스를 모달 내부로 트랩해야 한다

---

## Specifications

### Page Structure

```
/                          → Dashboard (FR-01)
/transports                → Transport List (FR-02)
/transports/new            → Transport Creation Form (FR-02.4)
/transports/:id            → Transport Detail (FR-02.9)
/transports/:id/monitor    → Real-time Monitoring (FR-03)
/jobs                      → Job History List (FR-04)
/jobs/:id                  → Job Detail (FR-04.4)
```

### Component Architecture

```
src/
├── components/
│   ├── ui/                    # shadcn/ui 기본 컴포넌트
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── dialog.tsx
│   │   ├── progress.tsx
│   │   ├── table.tsx
│   │   └── toast.tsx
│   ├── dashboard/
│   │   ├── ConnectionStatus.tsx
│   │   ├── QuickStats.tsx
│   │   ├── RecentJobsWidget.tsx
│   │   └── RunningJobBanner.tsx
│   ├── transport/
│   │   ├── TransportList.tsx
│   │   ├── TransportForm.tsx
│   │   ├── TransportDetail.tsx
│   │   ├── TableSelector.tsx
│   │   └── DeleteConfirmDialog.tsx
│   ├── monitoring/
│   │   ├── OverallProgress.tsx
│   │   ├── TableProgressList.tsx
│   │   ├── LiveMetrics.tsx
│   │   ├── CompletionSummary.tsx
│   │   └── SSEProvider.tsx
│   └── job/
│       ├── JobList.tsx
│       ├── JobDetail.tsx
│       ├── JobFilters.tsx
│       └── ExtractionAccordion.tsx
├── hooks/
│   ├── useSSE.ts              # SSE 연결 관리 훅
│   ├── useTransports.ts       # Transport CRUD 훅
│   ├── useJobs.ts             # Job 조회 훅
│   └── useConnectionStatus.ts # 연결 상태 폴링 훅
├── lib/
│   ├── api.ts                 # API 클라이언트 (fetch wrapper)
│   ├── sse.ts                 # SSE 유틸리티
│   └── format.ts              # 포맷팅 유틸리티 (bytes, duration)
└── store/
    └── authStore.ts           # Zustand 인증 상태
```

### API Integration

#### API Client Configuration

```typescript
// lib/api.ts
const API_BASE = '/api';

interface ApiConfig {
  baseUrl: string;
  headers: () => Record<string, string>;
}

export const api = {
  get: <T>(path: string) => fetch(`${API_BASE}${path}`, { headers }),
  post: <T>(path: string, body: unknown) => fetch(...),
  delete: <T>(path: string) => fetch(...),
};
```

#### SSE Event Handler

```typescript
// hooks/useSSE.ts
interface SSEEvents {
  onProgress: (data: ProgressData) => void;
  onTableComplete: (data: TableCompleteData) => void;
  onJobComplete: (data: JobCompleteData) => void;
  onError: (data: ErrorData) => void;
}

function useSSE(transportId: string, events: SSEEvents) {
  // EventSource 생성, 이벤트 핸들링, 재연결 로직
}
```

### State Management

#### Server State (TanStack Query)

```typescript
// Transport 목록 조회
const { data, isLoading } = useQuery({
  queryKey: ['transports', { offset, limit }],
  queryFn: () => api.getTransports({ offset, limit }),
});

// Transport 생성
const { mutate } = useMutation({
  mutationFn: api.createTransport,
  onSuccess: () => queryClient.invalidateQueries(['transports']),
});
```

#### Client State (Zustand)

```typescript
// 인증 상태
interface AuthState {
  token: string | null;
  setToken: (token: string) => void;
  clearToken: () => void;
}
```

---

## Constraints

### Technical Constraints

| ID | Constraint | Rationale |
|----|------------|-----------|
| TC-01 | React 19 + TypeScript 5.9+ | 최신 기능 및 타입 안정성 |
| TC-02 | Node.js 22+ (빌드 환경) | Vite 6 호환성 |
| TC-03 | 번들 크기 500KB 이하 (gzip) | 초기 로드 성능 |
| TC-04 | 브라우저 지원: Chrome 120+, Firefox 120+ | ESM 및 최신 API 사용 |

### Design Constraints

| ID | Constraint | Rationale |
|----|------------|-----------|
| DC-01 | shadcn/ui 컴포넌트 우선 사용 | 일관된 디자인 시스템 |
| DC-02 | Tailwind CSS 유틸리티 우선 | 스타일 유지보수성 |
| DC-03 | 다크 모드 미지원 (1차) | 개발 범위 제한 |

---

## Traceability

| Requirement | Test Scenario | Acceptance Criteria |
|-------------|---------------|---------------------|
| FR-01.1 | TC-DASH-001 | AC-01 |
| FR-01.3 | TC-DASH-002 | AC-02 |
| FR-02.1 | TC-TRP-001 | AC-03 |
| FR-02.4 | TC-TRP-002 | AC-04 |
| FR-02.8 | TC-TRP-003 | AC-05 |
| FR-03.1 | TC-MON-001 | AC-06 |
| FR-03.4 | TC-MON-002 | AC-07 |
| FR-04.1 | TC-JOB-001 | AC-08 |
| FR-04.4 | TC-JOB-002 | AC-09 |
| NFR-01.1 | TC-PERF-001 | AC-10 |
| NFR-02.1 | TC-UX-001 | AC-11 |

---

## Related Documents

- SPEC-ETL-001/spec.md - Backend API Specification
- docs/API.md - API Documentation
- SPEC-ETL-002/plan.md - Implementation Plan
- SPEC-ETL-002/acceptance.md - Acceptance Criteria

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-18 | Claude | Initial SPEC creation |
