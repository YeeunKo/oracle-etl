/**
 * 최근 작업 히스토리 위젯
 * 최근 5개의 Job을 테이블로 표시
 */

'use client';

import { useQuery } from '@tanstack/react-query';
import { getJobs } from '@/lib/api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { formatDateCompact, formatNumber } from '@/lib/format';
import type { JobStatus } from '@/types/api';
import Link from 'next/link';

// ============================================================================
// 상수
// ============================================================================

/**
 * 표시할 최근 작업 수
 */
const RECENT_JOBS_COUNT = 5;

/**
 * Job 상태별 스타일 설정
 */
const JOB_STATUS_STYLES: Record<JobStatus, {
  variant: 'default' | 'secondary' | 'destructive' | 'outline';
  label: string;
}> = {
  pending: { variant: 'secondary', label: '대기 중' },
  running: { variant: 'default', label: '실행 중' },
  completed: { variant: 'outline', label: '완료' },
  failed: { variant: 'destructive', label: '실패' },
  cancelled: { variant: 'secondary', label: '취소됨' },
};

// ============================================================================
// Query Keys
// ============================================================================

export const recentJobsQueryKeys = {
  recent: ['dashboard', 'recentJobs'] as const,
};

// ============================================================================
// 서브 컴포넌트
// ============================================================================

/**
 * Job 상태 뱃지
 */
function JobStatusBadge({ status }: { status: JobStatus }) {
  const style = JOB_STATUS_STYLES[status] || {
    variant: 'secondary' as const,
    label: status,
  };

  return (
    <Badge variant={style.variant} className="text-xs">
      {style.label}
    </Badge>
  );
}

/**
 * 테이블 로딩 스켈레톤
 */
function TableSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: RECENT_JOBS_COUNT }).map((_, i) => (
        <div key={i} className="flex items-center gap-4">
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-4 w-32 flex-1" />
          <Skeleton className="h-5 w-16" />
          <Skeleton className="h-4 w-28" />
          <Skeleton className="h-4 w-20" />
        </div>
      ))}
    </div>
  );
}

/**
 * 빈 상태 표시
 */
function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-8 text-center">
      <p className="text-muted-foreground text-sm">
        아직 실행된 작업이 없습니다.
      </p>
      <p className="text-muted-foreground text-xs mt-1">
        Transport를 실행하면 여기에 표시됩니다.
      </p>
    </div>
  );
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * 최근 작업 히스토리 위젯
 * 최근 5개의 Job을 테이블 형식으로 표시
 *
 * @example
 * <RecentJobs />
 */
export function RecentJobs() {
  const {
    data: jobsData,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: recentJobsQueryKeys.recent,
    queryFn: () => getJobs({ limit: RECENT_JOBS_COUNT }),
    staleTime: 10 * 1000, // 10초
    refetchInterval: 30 * 1000, // 30초마다 갱신
  });

  const jobs = jobsData?.data ?? [];

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <CardTitle className="text-base font-semibold">
          최근 작업 히스토리
        </CardTitle>
        <Link href="/jobs" passHref>
          <Button variant="ghost" size="sm" className="text-xs">
            전체 보기
          </Button>
        </Link>
      </CardHeader>
      <CardContent>
        {/* 로딩 상태 */}
        {isLoading && <TableSkeleton />}

        {/* 에러 상태 */}
        {isError && (
          <div className="flex flex-col items-center justify-center py-6 text-center">
            <p className="text-destructive text-sm">작업 목록을 불러오는데 실패했습니다.</p>
            <p className="text-muted-foreground text-xs mt-1">
              {error instanceof Error ? error.message : '알 수 없는 오류'}
            </p>
            <Button
              variant="outline"
              size="sm"
              className="mt-3"
              onClick={() => refetch()}
            >
              다시 시도
            </Button>
          </div>
        )}

        {/* 빈 상태 */}
        {!isLoading && !isError && jobs.length === 0 && <EmptyState />}

        {/* 작업 테이블 */}
        {!isLoading && !isError && jobs.length > 0 && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-24">Job ID</TableHead>
                <TableHead>Transport</TableHead>
                <TableHead className="w-20">상태</TableHead>
                <TableHead className="w-36">시작 시간</TableHead>
                <TableHead className="w-28 text-right">처리 행 수</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {jobs.map((job) => (
                <TableRow key={job.id}>
                  <TableCell className="font-mono text-xs">
                    {job.id.slice(0, 8)}
                  </TableCell>
                  <TableCell className="font-medium">
                    {job.transportName}
                  </TableCell>
                  <TableCell>
                    <JobStatusBadge status={job.status} />
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDateCompact(job.startedAt)}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {formatNumber(job.progress.rowsProcessed)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
