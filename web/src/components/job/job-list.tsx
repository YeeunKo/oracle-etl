/**
 * Job 목록 컴포넌트
 * 필터링, 페이지네이션, 정렬 기능을 포함한 Job 히스토리 테이블
 */

'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { getJobs, getTransports } from '@/lib/api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { JobStatusBadge } from './job-status-badge';
import { formatDateCompact, formatNumber, formatDuration, formatBytes } from '@/lib/format';
import type { Job, JobStatus } from '@/types/api';
import {
  ChevronLeft,
  ChevronRight,
  RefreshCw,
  ExternalLink,
  Filter,
  X,
} from 'lucide-react';

// ============================================================================
// 상수
// ============================================================================

/**
 * 페이지당 항목 수
 */
const PAGE_SIZE = 20;

/**
 * 상태 필터 옵션
 */
const STATUS_OPTIONS: { value: JobStatus | 'all'; label: string }[] = [
  { value: 'all', label: '전체 상태' },
  { value: 'running', label: '실행 중' },
  { value: 'completed', label: '완료' },
  { value: 'failed', label: '실패' },
  { value: 'pending', label: '대기 중' },
  { value: 'cancelled', label: '취소됨' },
];

// ============================================================================
// Query Keys
// ============================================================================

export const jobListQueryKeys = {
  all: ['jobs'] as const,
  list: (params: { status?: JobStatus; transportId?: string; page: number }) =>
    ['jobs', 'list', params] as const,
  transports: ['transports', 'all'] as const,
};

// ============================================================================
// 타입
// ============================================================================

interface JobListProps {
  /** 초기 상태 필터 */
  initialStatus?: JobStatus;
  /** 초기 Transport 필터 */
  initialTransportId?: string;
  /** 추가 클래스명 */
  className?: string;
}

// ============================================================================
// 헬퍼 함수
// ============================================================================

/**
 * Job 실행 시간 계산
 */
function calculateDuration(job: Job): number {
  const start = new Date(job.startedAt).getTime();
  const end = job.completedAt
    ? new Date(job.completedAt).getTime()
    : Date.now();
  return end - start;
}

// ============================================================================
// 서브 컴포넌트
// ============================================================================

/**
 * 테이블 로딩 스켈레톤
 */
function TableSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex items-center gap-4 py-2">
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-4 w-32 flex-1" />
          <Skeleton className="h-5 w-16" />
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-8 w-8" />
        </div>
      ))}
    </div>
  );
}

/**
 * 빈 상태
 */
function EmptyState({ hasFilters }: { hasFilters: boolean }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <p className="text-muted-foreground text-sm">
        {hasFilters
          ? '검색 조건에 맞는 작업이 없습니다.'
          : '아직 실행된 작업이 없습니다.'}
      </p>
      <p className="text-muted-foreground text-xs mt-1">
        {hasFilters
          ? '필터를 변경해 보세요.'
          : 'Transport를 실행하면 여기에 표시됩니다.'}
      </p>
    </div>
  );
}

/**
 * 페이지네이션 컨트롤
 */
function Pagination({
  currentPage,
  totalPages,
  totalItems,
  onPageChange,
}: {
  currentPage: number;
  totalPages: number;
  totalItems: number;
  onPageChange: (page: number) => void;
}) {
  const startItem = (currentPage - 1) * PAGE_SIZE + 1;
  const endItem = Math.min(currentPage * PAGE_SIZE, totalItems);

  return (
    <div className="flex items-center justify-between pt-4">
      <p className="text-sm text-muted-foreground">
        총 {formatNumber(totalItems)}개 중 {startItem}-{endItem} 표시
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage <= 1}
        >
          <ChevronLeft className="h-4 w-4" />
          이전
        </Button>
        <span className="text-sm text-muted-foreground px-2">
          {currentPage} / {totalPages}
        </span>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage >= totalPages}
        >
          다음
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * Job 목록 컴포넌트
 *
 * @example
 * <JobList />
 * <JobList initialStatus="running" />
 */
export function JobList({
  initialStatus,
  initialTransportId,
  className,
}: JobListProps) {
  // 필터 상태
  const [statusFilter, setStatusFilter] = useState<JobStatus | 'all'>(
    initialStatus || 'all'
  );
  const [transportFilter, setTransportFilter] = useState<string>(
    initialTransportId || 'all'
  );
  const [currentPage, setCurrentPage] = useState(1);

  // 필터가 활성화되어 있는지 확인
  const hasFilters = statusFilter !== 'all' || transportFilter !== 'all';

  // Transport 목록 조회 (필터 드롭다운용)
  const { data: transportsData } = useQuery({
    queryKey: jobListQueryKeys.transports,
    queryFn: () => getTransports({ limit: 100 }),
    staleTime: 60 * 1000, // 1분
  });

  const transports = transportsData?.data ?? [];

  // Job 목록 조회
  const {
    data: jobsData,
    isLoading,
    isError,
    error,
    refetch,
    isFetching,
  } = useQuery({
    queryKey: jobListQueryKeys.list({
      status: statusFilter !== 'all' ? statusFilter : undefined,
      transportId: transportFilter !== 'all' ? transportFilter : undefined,
      page: currentPage,
    }),
    queryFn: () =>
      getJobs({
        status: statusFilter !== 'all' ? statusFilter : undefined,
        transportId: transportFilter !== 'all' ? transportFilter : undefined,
        offset: (currentPage - 1) * PAGE_SIZE,
        limit: PAGE_SIZE,
      }),
    staleTime: 10 * 1000, // 10초
    refetchInterval: 30 * 1000, // 30초마다 갱신
  });

  const jobs = jobsData?.data ?? [];
  const totalItems = jobsData?.total ?? 0;
  const totalPages = Math.ceil(totalItems / PAGE_SIZE);

  // 필터 초기화
  const clearFilters = () => {
    setStatusFilter('all');
    setTransportFilter('all');
    setCurrentPage(1);
  };

  // 필터 변경 시 페이지 리셋
  const handleStatusChange = (value: string) => {
    setStatusFilter(value as JobStatus | 'all');
    setCurrentPage(1);
  };

  const handleTransportChange = (value: string) => {
    setTransportFilter(value);
    setCurrentPage(1);
  };

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold">작업 히스토리</CardTitle>
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            disabled={isFetching}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${isFetching ? 'animate-spin' : ''}`} />
            새로고침
          </Button>
        </div>

        {/* 필터 */}
        <div className="flex items-center gap-3 pt-3">
          <Filter className="h-4 w-4 text-muted-foreground" />

          {/* 상태 필터 */}
          <Select value={statusFilter} onValueChange={handleStatusChange}>
            <SelectTrigger className="w-36">
              <SelectValue placeholder="상태 선택" />
            </SelectTrigger>
            <SelectContent>
              {STATUS_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {/* Transport 필터 */}
          <Select value={transportFilter} onValueChange={handleTransportChange}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder="Transport 선택" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">전체 Transport</SelectItem>
              {transports.map((transport) => (
                <SelectItem key={transport.id} value={transport.id}>
                  {transport.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {/* 필터 초기화 */}
          {hasFilters && (
            <Button variant="ghost" size="sm" onClick={clearFilters}>
              <X className="h-4 w-4 mr-1" />
              초기화
            </Button>
          )}
        </div>
      </CardHeader>

      <CardContent>
        {/* 로딩 상태 */}
        {isLoading && <TableSkeleton />}

        {/* 에러 상태 */}
        {isError && (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <p className="text-destructive text-sm">
              작업 목록을 불러오는데 실패했습니다.
            </p>
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
        {!isLoading && !isError && jobs.length === 0 && (
          <EmptyState hasFilters={hasFilters} />
        )}

        {/* 작업 테이블 */}
        {!isLoading && !isError && jobs.length > 0 && (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-24">Job ID</TableHead>
                  <TableHead>Transport</TableHead>
                  <TableHead className="w-24">상태</TableHead>
                  <TableHead className="w-40">시작 시간</TableHead>
                  <TableHead className="w-24 text-right">실행 시간</TableHead>
                  <TableHead className="w-28 text-right">처리 행 수</TableHead>
                  <TableHead className="w-24 text-right">데이터 크기</TableHead>
                  <TableHead className="w-16"></TableHead>
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
                    <TableCell className="text-right text-sm font-mono">
                      {formatDuration(calculateDuration(job))}
                    </TableCell>
                    <TableCell className="text-right font-mono text-sm">
                      {formatNumber(job.progress.rowsProcessed)}
                    </TableCell>
                    <TableCell className="text-right text-sm">
                      {formatBytes(job.progress.bytesProcessed)}
                    </TableCell>
                    <TableCell>
                      <Link href={`/jobs/${job.id}`} passHref>
                        <Button variant="ghost" size="sm" title="상세 보기">
                          <ExternalLink className="h-4 w-4" />
                        </Button>
                      </Link>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>

            {/* 페이지네이션 */}
            {totalPages > 1 && (
              <Pagination
                currentPage={currentPage}
                totalPages={totalPages}
                totalItems={totalItems}
                onPageChange={setCurrentPage}
              />
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
