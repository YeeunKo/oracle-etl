/**
 * 실행 중인 작업 알림 배너
 * 현재 실행 중인 작업이 있으면 상단에 배너 표시
 */

'use client';

import { useQuery } from '@tanstack/react-query';
import { getJobs } from '@/lib/api';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { Progress } from '@/components/ui/progress';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import Link from 'next/link';
import type { Job } from '@/types/api';

// ============================================================================
// Query Keys
// ============================================================================

export const runningJobQueryKeys = {
  running: ['dashboard', 'runningJobs'] as const,
};

// ============================================================================
// 유틸리티 함수
// ============================================================================

/**
 * 진행 상태 phase 레이블
 */
function getPhaseLabel(phase: string): string {
  const labels: Record<string, string> = {
    initializing: '초기화 중',
    extracting: '추출 중',
    transforming: '변환 중',
    loading: '로딩 중',
    completed: '완료',
  };
  return labels[phase] || phase;
}

// ============================================================================
// 서브 컴포넌트
// ============================================================================

/**
 * 로딩 스켈레톤
 */
function BannerSkeleton() {
  return (
    <div className="rounded-lg border p-4 bg-card">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <Skeleton className="h-5 w-48 mb-2" />
          <Skeleton className="h-4 w-64 mb-3" />
          <Skeleton className="h-2 w-full" />
        </div>
        <Skeleton className="h-8 w-24 ml-4" />
      </div>
    </div>
  );
}

/**
 * 단일 작업 배너
 */
function SingleJobBanner({ job }: { job: Job }) {
  const { progress } = job;

  return (
    <Alert className="border-primary/50 bg-primary/5">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <AlertTitle className="flex items-center gap-2">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-primary" />
            </span>
            작업 실행 중
            <Badge variant="outline" className="ml-2 text-xs">
              {getPhaseLabel(progress.phase)}
            </Badge>
          </AlertTitle>

          <AlertDescription className="mt-2">
            <div className="space-y-2">
              <p className="text-sm">
                <span className="font-medium">{job.transportName}</span>
                {progress.currentTable && (
                  <span className="text-muted-foreground">
                    {' '}/ {progress.currentTable}
                  </span>
                )}
              </p>

              {/* 진행률 바 */}
              <div className="flex items-center gap-3">
                <Progress value={progress.percentage} className="flex-1 h-2" />
                <span className="text-sm font-medium min-w-12 text-right">
                  {progress.percentage.toFixed(0)}%
                </span>
              </div>

              {/* 처리 통계 */}
              <p className="text-xs text-muted-foreground">
                {progress.rowsProcessed.toLocaleString()} / {progress.totalRows.toLocaleString()} 행 처리됨
                {progress.estimatedTimeRemaining !== undefined && progress.estimatedTimeRemaining > 0 && (
                  <span className="ml-2">
                    (약 {Math.ceil(progress.estimatedTimeRemaining / 1000)}초 남음)
                  </span>
                )}
              </p>
            </div>
          </AlertDescription>
        </div>

        {/* 모니터링 페이지 링크 */}
        <Link href={`/jobs/${job.id}`} passHref>
          <Button variant="outline" size="sm">
            상세 보기
          </Button>
        </Link>
      </div>
    </Alert>
  );
}

/**
 * 다중 작업 배너
 */
function MultipleJobsBanner({ jobs }: { jobs: Job[] }) {
  return (
    <Alert className="border-primary/50 bg-primary/5">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <AlertTitle className="flex items-center gap-2">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-primary" />
            </span>
            {jobs.length}개 작업 실행 중
          </AlertTitle>

          <AlertDescription className="mt-2">
            <div className="space-y-1">
              {jobs.slice(0, 3).map((job) => (
                <div key={job.id} className="flex items-center gap-2 text-sm">
                  <Badge variant="outline" className="text-xs">
                    {getPhaseLabel(job.progress.phase)}
                  </Badge>
                  <span className="font-medium truncate">{job.transportName}</span>
                  <span className="text-muted-foreground">
                    {job.progress.percentage.toFixed(0)}%
                  </span>
                </div>
              ))}
              {jobs.length > 3 && (
                <p className="text-xs text-muted-foreground mt-1">
                  외 {jobs.length - 3}개 작업 더 있음
                </p>
              )}
            </div>
          </AlertDescription>
        </div>

        <Link href="/jobs?status=running" passHref>
          <Button variant="outline" size="sm">
            전체 보기
          </Button>
        </Link>
      </div>
    </Alert>
  );
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * 실행 중인 작업 알림 배너
 * 현재 실행 중인 작업이 있으면 상단에 표시
 * 진행률 바와 모니터링 페이지 링크 포함
 *
 * @example
 * <RunningJobBanner />
 */
export function RunningJobBanner() {
  const {
    data: jobsData,
    isLoading,
  } = useQuery({
    queryKey: runningJobQueryKeys.running,
    queryFn: () => getJobs({ status: 'running', limit: 10 }),
    staleTime: 5 * 1000, // 5초
    refetchInterval: 5 * 1000, // 5초마다 갱신 (실시간 진행률 표시)
  });

  // 실행 중인 작업 필터링 (pending 포함)
  const runningJobs = jobsData?.data?.filter(
    job => job.status === 'running' || job.status === 'pending'
  ) ?? [];

  // 로딩 중이면서 첫 로드인 경우에만 스켈레톤 표시
  if (isLoading && !jobsData) {
    return <BannerSkeleton />;
  }

  // 실행 중인 작업이 없으면 아무것도 표시하지 않음
  if (runningJobs.length === 0) {
    return null;
  }

  // 단일 작업 vs 다중 작업
  if (runningJobs.length === 1) {
    return <SingleJobBanner job={runningJobs[0]} />;
  }

  return <MultipleJobsBanner jobs={runningJobs} />;
}
