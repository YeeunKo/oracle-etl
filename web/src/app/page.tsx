/**
 * 대시보드 메인 페이지
 * Oracle ETL Pipeline 현황 및 상태 표시
 */

'use client';

import { MainLayout } from '@/components/layout/main-layout';
import {
  ConnectionStatus,
  QuickStats,
  RecentJobs,
  RunningJobBanner,
} from '@/components/dashboard';

// ============================================================================
// 메인 페이지 컴포넌트
// ============================================================================

/**
 * 대시보드 페이지
 * - Oracle 연결 상태 표시
 * - 빠른 통계 요약
 * - 실행 중인 작업 알림
 * - 최근 작업 히스토리
 */
export default function DashboardPage() {
  return (
    <MainLayout>
      <div className="space-y-6">
        {/* 페이지 헤더 */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">대시보드</h1>
            <p className="text-muted-foreground">
              Oracle ETL Pipeline 현황을 확인하세요.
            </p>
          </div>
        </div>

        {/* 실행 중인 작업 알림 배너 (FR-01.5) */}
        <RunningJobBanner />

        {/* 빠른 통계 요약 (FR-01.4) */}
        <QuickStats />

        {/* 메인 컨텐츠 그리드 */}
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Oracle 연결 상태 (FR-01.1, FR-01.2) */}
          <div className="lg:col-span-1">
            <ConnectionStatus />
          </div>

          {/* 최근 작업 히스토리 (FR-01.3) */}
          <div className="lg:col-span-2">
            <RecentJobs />
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
