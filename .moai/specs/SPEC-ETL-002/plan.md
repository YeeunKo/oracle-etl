# SPEC-ETL-002: Implementation Plan

## Overview

Oracle ETL Pipeline Frontend UI 구현 계획서. SPEC-ETL-001 백엔드 API를 기반으로 React 19 + TypeScript 프론트엔드 애플리케이션을 구현합니다.

---

## Implementation Approach

### Architecture Pattern
- **Component-Based Architecture**: React 19 함수형 컴포넌트
- **Server State Management**: TanStack Query v5로 API 상태 관리
- **Client State Management**: Zustand로 인증 및 UI 상태 관리
- **Design System**: shadcn/ui + Tailwind CSS

### Development Strategy
- **Feature-Sliced Development**: 화면별 독립적 개발
- **API-First Integration**: MSW(Mock Service Worker)로 백엔드 병렬 개발
- **TDD Approach**: Vitest + React Testing Library

---

## Milestones

### Primary Goal: Project Setup and Core Infrastructure

**Scope**: 프로젝트 초기 설정, 공통 컴포넌트, API 클라이언트

**Tasks**:

1. **Project Initialization**
   - Vite + React 19 + TypeScript 프로젝트 생성
   - ESLint, Prettier 설정
   - 디렉토리 구조 생성

2. **Design System Setup**
   - Tailwind CSS 4.0 설정
   - shadcn/ui 컴포넌트 설치 (Button, Card, Dialog, Table, Progress, Toast)
   - 기본 레이아웃 컴포넌트 (Header, Sidebar, Layout)

3. **API Client Infrastructure**
   - Fetch 기반 API 클라이언트 구현
   - TanStack Query 설정 및 Provider 구성
   - 에러 핸들링 유틸리티

4. **Routing Setup**
   - React Router v7 설정
   - 라우트 정의 (/, /transports, /jobs 등)
   - 인증 가드 (ProtectedRoute)

5. **Authentication**
   - Zustand 인증 스토어
   - API Key / Bearer Token 헤더 인젝션
   - 로그인 페이지 (API Key 입력)

**Dependencies**: SPEC-ETL-001 API 명세서

**Deliverables**:
- 실행 가능한 Vite 개발 서버
- 기본 페이지 라우팅
- API 클라이언트 구현

---

### Secondary Goal: Dashboard Implementation

**Scope**: 대시보드 페이지 및 위젯 구현 (FR-01)

**Tasks**:

1. **Connection Status Component**
   - Oracle 연결 상태 표시 (FR-01.1)
   - 30초 자동 갱신 (FR-01.2)
   - `/api/health`, `/api/oracle/status` 통합

2. **Quick Stats Widget**
   - 통계 카드 UI (FR-01.4)
   - Transport 수, Job 수, 실행 중 Job 카운트

3. **Recent Jobs Widget**
   - 최근 5개 Job 테이블 (FR-01.3)
   - Job 상태별 배지 스타일링
   - Job 상세 페이지 링크

4. **Running Job Banner**
   - 실행 중 Job 알림 배너 (FR-01.5)
   - 실시간 진행률 미니 프로그레스 바
   - 모니터링 페이지 바로가기

**Dependencies**: Primary Goal 완료

**Deliverables**:
- 대시보드 페이지 완성
- 자동 갱신 연결 상태
- 실행 중 Job 알림

---

### Tertiary Goal: Transport Management

**Scope**: Transport CRUD 기능 구현 (FR-02)

**Tasks**:

1. **Transport List Page**
   - 테이블 뷰 구현 (FR-02.1)
   - 페이지네이션 (FR-02.2)
   - 검색 및 필터 (FR-02.3)

2. **Transport Creation**
   - 생성 폼 모달/페이지 (FR-02.4)
   - 테이블 선택 UI (FR-02.5)
   - React Hook Form + Zod 유효성 검사

3. **Transport Detail Page**
   - 상세 정보 표시 (FR-02.9)
   - 테이블 목록
   - 실행 이력 탭

4. **Transport Actions**
   - 삭제 확인 다이얼로그 (FR-02.6)
   - 실행 중 삭제 방지 (FR-02.7)
   - 실행 트리거 (FR-02.8)

**Dependencies**: Secondary Goal 완료

**Deliverables**:
- Transport CRUD 전체 기능
- 테이블 선택 컴포넌트
- 삭제/실행 확인 다이얼로그

---

### Quaternary Goal: Real-time Monitoring

**Scope**: SSE 기반 실시간 모니터링 구현 (FR-03)

**Tasks**:

1. **SSE Infrastructure**
   - useSSE 커스텀 훅 구현 (FR-03.1)
   - EventSource 연결/해제 관리
   - 자동 재연결 로직 (FR-03.6)

2. **Progress Components**
   - 전체 진행률 프로그레스 바 (FR-03.2)
   - 테이블별 진행률 리스트 (FR-03.3)
   - 상태별 색상 및 아이콘

3. **Live Metrics Display**
   - 실시간 메트릭 카드 (FR-03.4)
   - rows/second, bytes transferred 포맷팅
   - 경과 시간 타이머

4. **Event Handling**
   - 에러 토스트 알림 (FR-03.5)
   - 완료 요약 모달 (FR-03.7)
   - 페이지 이탈/복귀 처리 (FR-03.8)

**Dependencies**: Tertiary Goal 완료, 백엔드 SSE 엔드포인트

**Deliverables**:
- 실시간 모니터링 페이지
- SSE 연결 관리 훅
- 진행률 시각화

---

### Fifth Goal: Job History

**Scope**: Job 이력 조회 기능 구현 (FR-04)

**Tasks**:

1. **Job List Page**
   - Job 목록 테이블 (FR-04.1)
   - 페이지네이션 (FR-04.2)
   - 필터링 (Transport, 상태, 날짜) (FR-04.3)

2. **Job Detail Page**
   - Job 메타데이터 표시 (FR-04.4)
   - 테이블별 추출 결과 아코디언 (FR-04.5)
   - GCS 경로 복사 기능

3. **Error Display**
   - 실패 Job 에러 메시지 (FR-04.6)
   - 실패 테이블 강조 표시

**Dependencies**: Quaternary Goal 완료

**Deliverables**:
- Job 이력 페이지
- Job 상세 페이지
- 필터링 및 검색

---

### Final Goal: Polish and Optimization

**Scope**: 비기능 요구사항 및 품질 개선

**Tasks**:

1. **Performance Optimization**
   - 초기 로드 최적화 (NFR-01.1)
   - 코드 스플리팅 (React.lazy)
   - 목록 가상화 (NFR-01.4)

2. **UX Improvements**
   - 로딩 스켈레톤 (NFR-02.2)
   - 에러 상태 UI (NFR-02.3)
   - 빈 상태 UI (NFR-02.4)

3. **Accessibility**
   - ARIA 레이블 적용 (NFR-05.1)
   - 키보드 네비게이션
   - 포커스 관리 (NFR-05.3)

4. **Error Handling**
   - Error Boundary 구현 (NFR-03.1)
   - 오프라인 감지 (NFR-03.2)
   - 세션 복구 (NFR-03.3)

5. **Testing**
   - 단위 테스트 (Vitest)
   - 통합 테스트 (React Testing Library)
   - E2E 테스트 (Playwright)

**Dependencies**: Fifth Goal 완료

**Deliverables**:
- 최적화된 번들 (500KB 이하)
- 테스트 커버리지 85%+
- 접근성 검수 완료

---

## Technical Approach

### API Integration Strategy

```typescript
// TanStack Query 설정
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30 * 1000, // 30초
      retry: 3,
      refetchOnWindowFocus: false,
    },
  },
});

// API 호출 패턴
const useTransports = (params: TransportListParams) => {
  return useQuery({
    queryKey: ['transports', params],
    queryFn: () => api.getTransports(params),
  });
};
```

### SSE Connection Management

```typescript
// SSE 훅 구조
function useSSE(transportId: string) {
  const [status, setStatus] = useState<'connecting' | 'connected' | 'error'>('connecting');
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    const es = new EventSource(`/api/transports/${transportId}/status`);
    eventSourceRef.current = es;

    es.onopen = () => setStatus('connected');
    es.onerror = () => handleReconnect();

    // 이벤트 리스너 등록
    es.addEventListener('progress', handleProgress);
    es.addEventListener('complete', handleComplete);

    return () => es.close();
  }, [transportId]);

  return { status, ... };
}
```

### Component Structure Pattern

```typescript
// Feature 컴포넌트 구조
// components/transport/TransportList.tsx
export function TransportList() {
  const [params, setParams] = useState({ offset: 0, limit: 20 });
  const { data, isLoading, error } = useTransports(params);

  if (isLoading) return <TransportListSkeleton />;
  if (error) return <ErrorState error={error} />;
  if (!data?.transports.length) return <EmptyState />;

  return (
    <Table>
      {data.transports.map(transport => (
        <TransportRow key={transport.id} transport={transport} />
      ))}
    </Table>
  );
}
```

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| 백엔드 API 지연 | High | Medium | MSW로 Mock API 구현, 병렬 개발 |
| SSE 브라우저 호환성 | Medium | Low | 폴리필 또는 폴백 메커니즘 |
| 번들 크기 초과 | Medium | Medium | 코드 스플리팅, 트리 쉐이킹 검증 |
| 실시간 업데이트 성능 | Medium | Low | 디바운싱, 가상화 적용 |

---

## Dependencies

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| react | ^19.0.0 | UI Framework |
| react-router-dom | ^7.0.0 | Routing |
| @tanstack/react-query | ^5.0.0 | Server State |
| zustand | ^5.0.0 | Client State |
| tailwindcss | ^4.0.0 | Styling |
| zod | ^3.23.0 | Validation |
| react-hook-form | ^7.54.0 | Form Handling |
| recharts | ^2.15.0 | Charts |
| vitest | ^2.0.0 | Testing |
| @testing-library/react | ^16.0.0 | Component Testing |
| playwright | ^1.49.0 | E2E Testing |

### Internal Dependencies

| Dependency | Type | Description |
|------------|------|-------------|
| SPEC-ETL-001 | SPEC | Backend API Specification |
| docs/API.md | Doc | API Endpoint Documentation |

---

## Quality Gates

### Code Quality
- [ ] ESLint 에러 0개
- [ ] TypeScript strict 모드 통과
- [ ] 테스트 커버리지 85%+

### Performance
- [ ] LCP 2.5초 이하
- [ ] 번들 크기 500KB 이하 (gzip)
- [ ] Lighthouse Performance 90+

### Accessibility
- [ ] WCAG 2.1 AA 준수
- [ ] 키보드 네비게이션 가능
- [ ] 스크린 리더 호환

---

## Traceability

| Milestone | Requirements Covered |
|-----------|---------------------|
| Primary Goal | NFR-01, NFR-03, NFR-04 |
| Secondary Goal | FR-01.1 ~ FR-01.5 |
| Tertiary Goal | FR-02.1 ~ FR-02.9 |
| Quaternary Goal | FR-03.1 ~ FR-03.8 |
| Fifth Goal | FR-04.1 ~ FR-04.6 |
| Final Goal | NFR-01 ~ NFR-05 |
