/**
 * Transport 목록 테이블 컴포넌트
 * FR-02.1: Transport 목록 표시
 * FR-02.2: 페이지네이션
 * FR-02.3: 이름 검색
 */

'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Plus, Play, Trash2, Search, ChevronLeft, ChevronRight, Eye } from 'lucide-react';
import { getTransports, executeTransport, deleteTransport } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { DeleteConfirmDialog } from './delete-confirm-dialog';
import { formatDateCompact } from '@/lib/format';
import type { Transport, TransportStatus } from '@/types/api';

// ============================================================================
// 상수
// ============================================================================

/**
 * 페이지당 표시할 Transport 수
 */
const PAGE_SIZE = 20;

/**
 * Transport 상태별 스타일 설정
 */
const TRANSPORT_STATUS_STYLES: Record<TransportStatus, {
  variant: 'default' | 'secondary' | 'destructive' | 'outline';
  label: string;
}> = {
  idle: { variant: 'secondary', label: '대기 중' },
  running: { variant: 'default', label: '실행 중' },
  completed: { variant: 'outline', label: '완료' },
  failed: { variant: 'destructive', label: '실패' },
  cancelled: { variant: 'secondary', label: '취소됨' },
};

// ============================================================================
// Query Keys
// ============================================================================

export const transportQueryKeys = {
  all: ['transports'] as const,
  list: (params: { offset?: number; limit?: number; search?: string }) =>
    ['transports', 'list', params] as const,
  detail: (id: string) => ['transports', 'detail', id] as const,
};

// ============================================================================
// 서브 컴포넌트
// ============================================================================

/**
 * Transport 상태 뱃지
 */
function TransportStatusBadge({ status }: { status: TransportStatus }) {
  const style = TRANSPORT_STATUS_STYLES[status] || {
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
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex items-center gap-4">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-4 w-40 flex-1" />
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-5 w-16" />
          <Skeleton className="h-4 w-28" />
          <Skeleton className="h-8 w-24" />
        </div>
      ))}
    </div>
  );
}

/**
 * 빈 상태 표시
 */
function EmptyState({ hasSearch }: { hasSearch: boolean }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <p className="text-muted-foreground text-sm">
        {hasSearch
          ? '검색 결과가 없습니다.'
          : '등록된 Transport가 없습니다.'}
      </p>
      {!hasSearch && (
        <p className="text-muted-foreground text-xs mt-1">
          새 Transport를 생성하여 데이터 추출을 시작하세요.
        </p>
      )}
      {!hasSearch && (
        <Link href="/transports/new" className="mt-4">
          <Button size="sm">
            <Plus className="mr-2 h-4 w-4" />
            새 Transport 생성
          </Button>
        </Link>
      )}
    </div>
  );
}

/**
 * 페이지네이션 컴포넌트
 */
function Pagination({
  currentPage,
  totalPages,
  onPageChange,
}: {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}) {
  if (totalPages <= 1) return null;

  return (
    <div className="flex items-center justify-between px-2 py-4">
      <p className="text-sm text-muted-foreground">
        페이지 {currentPage} / {totalPages}
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
 * Transport 목록 테이블
 * 검색, 페이지네이션, 실행/삭제 기능 포함
 */
export function TransportList() {
  const router = useRouter();
  const queryClient = useQueryClient();

  // 상태 관리
  const [currentPage, setCurrentPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [deleteTarget, setDeleteTarget] = useState<Transport | null>(null);

  // 목록 조회
  const {
    data: transportsData,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: transportQueryKeys.list({
      offset: (currentPage - 1) * PAGE_SIZE,
      limit: PAGE_SIZE,
      search: searchQuery || undefined,
    }),
    queryFn: () =>
      getTransports({
        offset: (currentPage - 1) * PAGE_SIZE,
        limit: PAGE_SIZE,
        search: searchQuery || undefined,
      }),
    staleTime: 30 * 1000, // 30초
  });

  // Transport 실행 mutation
  const executeMutation = useMutation({
    mutationFn: (id: string) => executeTransport(id),
    onSuccess: (data) => {
      // Job 모니터링 페이지로 이동 (FR-02.8)
      router.push(`/jobs/${data.jobId}`);
    },
    onError: (error) => {
      console.error('Transport 실행 실패:', error);
    },
  });

  // Transport 삭제 mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteTransport(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: transportQueryKeys.all });
      setDeleteTarget(null);
    },
    onError: (error) => {
      console.error('Transport 삭제 실패:', error);
    },
  });

  // 검색 핸들러
  const handleSearch = () => {
    setSearchQuery(searchInput);
    setCurrentPage(1);
  };

  // 검색 키 입력 핸들러
  const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  // 페이지 변경 핸들러
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  // 실행 핸들러
  const handleExecute = (transport: Transport) => {
    if (transport.status === 'running') return;
    executeMutation.mutate(transport.id);
  };

  // 삭제 확인 핸들러
  const handleDeleteConfirm = () => {
    if (deleteTarget) {
      deleteMutation.mutate(deleteTarget.id);
    }
  };

  const transports = transportsData?.data ?? [];
  const total = transportsData?.total ?? 0;
  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="space-y-4">
      {/* 검색 및 생성 버튼 */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2 flex-1 max-w-md">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Transport 이름으로 검색..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              onKeyDown={handleSearchKeyDown}
              className="pl-9"
            />
          </div>
          <Button variant="secondary" size="sm" onClick={handleSearch}>
            검색
          </Button>
        </div>
        <Link href="/transports/new">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            새 Transport
          </Button>
        </Link>
      </div>

      {/* 로딩 상태 */}
      {isLoading && <TableSkeleton />}

      {/* 에러 상태 */}
      {isError && (
        <div className="flex flex-col items-center justify-center py-8 text-center">
          <p className="text-destructive text-sm">
            Transport 목록을 불러오는데 실패했습니다.
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
      {!isLoading && !isError && transports.length === 0 && (
        <EmptyState hasSearch={!!searchQuery} />
      )}

      {/* Transport 테이블 */}
      {!isLoading && !isError && transports.length > 0 && (
        <>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-48">이름</TableHead>
                  <TableHead>소스 테이블</TableHead>
                  <TableHead>대상 경로</TableHead>
                  <TableHead className="w-24">상태</TableHead>
                  <TableHead className="w-36">마지막 실행</TableHead>
                  <TableHead className="w-32 text-right">작업</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {transports.map((transport) => (
                  <TableRow key={transport.id}>
                    <TableCell className="font-medium">
                      <Link
                        href={`/transports/${transport.id}`}
                        className="hover:underline"
                      >
                        {transport.name}
                      </Link>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {transport.sourceSchema}.{transport.sourceTable}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground truncate max-w-[200px]">
                      {transport.targetPath}
                    </TableCell>
                    <TableCell>
                      <TransportStatusBadge status={transport.status} />
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {transport.lastRunAt
                        ? formatDateCompact(transport.lastRunAt)
                        : '-'}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        {/* 상세 보기 버튼 */}
                        <Link href={`/transports/${transport.id}`}>
                          <Button variant="ghost" size="sm" title="상세 보기">
                            <Eye className="h-4 w-4" />
                          </Button>
                        </Link>
                        {/* 실행 버튼 (FR-02.8) */}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleExecute(transport)}
                          disabled={
                            transport.status === 'running' ||
                            executeMutation.isPending
                          }
                          title="실행"
                        >
                          <Play className="h-4 w-4" />
                        </Button>
                        {/* 삭제 버튼 (FR-02.6, FR-02.7) */}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setDeleteTarget(transport)}
                          disabled={transport.status === 'running'}
                          title={
                            transport.status === 'running'
                              ? '실행 중인 Transport는 삭제할 수 없습니다'
                              : '삭제'
                          }
                          className="text-destructive hover:text-destructive"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* 페이지네이션 (FR-02.2) */}
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={handlePageChange}
          />
        </>
      )}

      {/* 삭제 확인 다이얼로그 (FR-02.6) */}
      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        transportName={deleteTarget?.name ?? ''}
        onConfirm={handleDeleteConfirm}
        isDeleting={deleteMutation.isPending}
      />
    </div>
  );
}
