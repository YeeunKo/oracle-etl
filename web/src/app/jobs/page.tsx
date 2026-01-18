/**
 * Job 히스토리 페이지 (FR-04)
 * Job 목록, 필터링, 페이지네이션 제공
 */

import { Metadata } from 'next';
import { JobList } from '@/components/job/job-list';

// ============================================================================
// 메타데이터
// ============================================================================

export const metadata: Metadata = {
  title: '작업 히스토리 | Oracle ETL Pipeline',
  description: 'ETL 작업 실행 히스토리를 조회하고 관리합니다.',
};

// ============================================================================
// 페이지 컴포넌트
// ============================================================================

/**
 * Job 히스토리 페이지
 * FR-04.1: Job 목록 테이블 (상태, 실행 시간, 처리 행 수)
 * FR-04.2: 페이지네이션 (20개 단위)
 * FR-04.3: 상태별 필터링
 * FR-04.4: Transport별 필터링
 * FR-04.5: 상세 페이지 링크
 */
export default function JobsPage() {
  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* 페이지 헤더 */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight">작업 히스토리</h1>
        <p className="text-muted-foreground mt-1">
          ETL 작업 실행 기록을 조회하고 상세 정보를 확인할 수 있습니다.
        </p>
      </div>

      {/* Job 목록 */}
      <JobList />
    </div>
  );
}
