/**
 * 빠른 통계 요약 컴포넌트
 * 전체 Transport 수, 오늘의 작업 수, 실행 중인 작업, 마지막 성공 시간
 */

'use client';

import { useQuery } from '@tanstack/react-query';
import { getTransports, getJobs } from '@/lib/api';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { formatRelativeTime } from '@/lib/format';
import type { Job, Transport } from '@/types/api';

// ============================================================================
// Query Keys
// ============================================================================

export const dashboardQueryKeys = {
  stats: ['dashboard', 'stats'] as const,
  transports: ['dashboard', 'transports'] as const,
  jobs: ['dashboard', 'jobs'] as const,
};

// ============================================================================
// 타입 정의
// ============================================================================

interface StatCardProps {
  title: string;
  value: string | number;
  description?: string;
  isLoading?: boolean;
}

interface QuickStatsData {
  totalTransports: number;
  todayJobs: number;
  runningJobs: number;
  lastSuccessTime: string | null;
}

// ============================================================================
// 유틸리티 함수
// ============================================================================

/**
 * 오늘 날짜인지 확인
 */
function isToday(dateString: string): boolean {
  const date = new Date(dateString);
  const today = new Date();

  return (
    date.getFullYear() === today.getFullYear() &&
    date.getMonth() === today.getMonth() &&
    date.getDate() === today.getDate()
  );
}

/**
 * 통계 데이터 계산
 */
function calculateStats(
  transports: Transport[] | undefined,
  jobs: Job[] | undefined
): QuickStatsData {
  const totalTransports = transports?.length ?? 0;

  // 오늘의 작업 수
  const todayJobs = jobs?.filter(job => isToday(job.startedAt)).length ?? 0;

  // 실행 중인 작업 수
  const runningJobs = jobs?.filter(
    job => job.status === 'running' || job.status === 'pending'
  ).length ?? 0;

  // 마지막 성공 작업 시간
  const completedJobs = jobs?.filter(job => job.status === 'completed') ?? [];
  const lastSuccessJob = completedJobs.sort((a, b) => {
    const dateA = a.completedAt ? new Date(a.completedAt).getTime() : 0;
    const dateB = b.completedAt ? new Date(b.completedAt).getTime() : 0;
    return dateB - dateA;
  })[0];

  return {
    totalTransports,
    todayJobs,
    runningJobs,
    lastSuccessTime: lastSuccessJob?.completedAt ?? null,
  };
}

// ============================================================================
// 서브 컴포넌트
// ============================================================================

/**
 * 개별 통계 카드
 */
function StatCard({ title, value, description, isLoading }: StatCardProps) {
  if (isLoading) {
    return (
      <Card>
        <CardContent className="pt-6">
          <Skeleton className="h-4 w-24 mb-2" />
          <Skeleton className="h-8 w-16 mb-1" />
          <Skeleton className="h-3 w-20" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <p className="text-sm font-medium text-muted-foreground">{title}</p>
        <p className="text-2xl font-bold mt-1">{value}</p>
        {description && (
          <p className="text-xs text-muted-foreground mt-1">{description}</p>
        )}
      </CardContent>
    </Card>
  );
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * 빠른 통계 요약 컴포넌트
 * Transport 수, 오늘의 작업, 실행 중인 작업, 마지막 성공 시간 표시
 *
 * @example
 * <QuickStats />
 */
export function QuickStats() {
  // Transport 목록 조회
  const {
    data: transportsData,
    isLoading: isTransportsLoading,
  } = useQuery({
    queryKey: dashboardQueryKeys.transports,
    queryFn: () => getTransports({ limit: 1000 }),
    staleTime: 30 * 1000, // 30초
    refetchInterval: 60 * 1000, // 1분마다 갱신
  });

  // Job 목록 조회
  const {
    data: jobsData,
    isLoading: isJobsLoading,
  } = useQuery({
    queryKey: dashboardQueryKeys.jobs,
    queryFn: () => getJobs({ limit: 100 }),
    staleTime: 10 * 1000, // 10초
    refetchInterval: 30 * 1000, // 30초마다 갱신
  });

  const isLoading = isTransportsLoading || isJobsLoading;

  const stats = calculateStats(transportsData?.data, jobsData?.data);

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="전체 Transport"
        value={stats.totalTransports}
        description="등록된 전송 규칙"
        isLoading={isLoading}
      />

      <StatCard
        title="오늘의 작업"
        value={stats.todayJobs}
        description="오늘 실행된 작업 수"
        isLoading={isLoading}
      />

      <StatCard
        title="실행 중"
        value={stats.runningJobs}
        description="현재 진행 중인 작업"
        isLoading={isLoading}
      />

      <StatCard
        title="마지막 성공"
        value={stats.lastSuccessTime
          ? formatRelativeTime(stats.lastSuccessTime)
          : '-'
        }
        description="마지막 성공 작업 시간"
        isLoading={isLoading}
      />
    </div>
  );
}

// ============================================================================
// 내보내기
// ============================================================================

export { dashboardQueryKeys as queryKeys };
