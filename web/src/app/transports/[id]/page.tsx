/**
 * Transport 상세 페이지
 * FR-02.9: Transport 상세 정보 표시
 * FR-02.8: 실행 버튼
 * FR-02.6, FR-02.7: 삭제 기능
 */

'use client';

import { use, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ChevronLeft,
  Play,
  Trash2,
  Settings,
  Database,
  FolderOutput,
  Clock,
  Calendar,
} from 'lucide-react';
import { getTransport, executeTransport, deleteTransport } from '@/lib/api';
import { MainLayout } from '@/components/layout/main-layout';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { DeleteConfirmDialog } from '@/components/transport/delete-confirm-dialog';
import { transportQueryKeys } from '@/components/transport/transport-list';
import { formatDate, formatNumber } from '@/lib/format';
import type { TransportStatus } from '@/types/api';

// ============================================================================
// 상수
// ============================================================================

/**
 * Transport 상태별 스타일 설정
 */
const TRANSPORT_STATUS_STYLES: Record<
  TransportStatus,
  {
    variant: 'default' | 'secondary' | 'destructive' | 'outline';
    label: string;
  }
> = {
  idle: { variant: 'secondary', label: '대기 중' },
  running: { variant: 'default', label: '실행 중' },
  completed: { variant: 'outline', label: '완료' },
  failed: { variant: 'destructive', label: '실패' },
  cancelled: { variant: 'secondary', label: '취소됨' },
};

/**
 * 출력 형식 레이블
 */
const FORMAT_LABELS: Record<string, string> = {
  csv: 'CSV',
  json: 'JSON',
  parquet: 'Parquet',
};

// ============================================================================
// Props 타입
// ============================================================================

interface TransportDetailPageProps {
  params: Promise<{ id: string }>;
}

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
    <Badge variant={style.variant} className="text-sm">
      {style.label}
    </Badge>
  );
}

/**
 * 정보 카드 아이템
 */
function InfoItem({
  icon: Icon,
  label,
  value,
  mono = false,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: React.ReactNode;
  mono?: boolean;
}) {
  return (
    <div className="flex items-start gap-3">
      <div className="flex h-8 w-8 items-center justify-center rounded-md bg-muted">
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>
      <div className="flex-1">
        <p className="text-sm text-muted-foreground">{label}</p>
        <p className={mono ? 'font-mono text-sm' : 'text-sm font-medium'}>
          {value}
        </p>
      </div>
    </div>
  );
}

/**
 * 로딩 스켈레톤
 */
function DetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Skeleton className="h-8 w-8" />
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-6 w-16" />
      </div>
      <div className="grid gap-6 md:grid-cols-2">
        <Skeleton className="h-48" />
        <Skeleton className="h-48" />
      </div>
      <Skeleton className="h-32" />
    </div>
  );
}

// ============================================================================
// 메인 페이지 컴포넌트
// ============================================================================

/**
 * Transport 상세 페이지
 * - Transport 상세 정보 표시
 * - 실행/삭제 액션
 */
export default function TransportDetailPage({
  params,
}: TransportDetailPageProps) {
  const { id } = use(params);
  const router = useRouter();
  const queryClient = useQueryClient();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  // Transport 상세 조회
  const {
    data: transport,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: transportQueryKeys.detail(id),
    queryFn: () => getTransport(id),
    staleTime: 30 * 1000,
  });

  // Transport 실행 mutation
  const executeMutation = useMutation({
    mutationFn: () => executeTransport(id),
    onSuccess: (data) => {
      // Job 모니터링 페이지로 이동 (FR-02.8)
      router.push(`/jobs/${data.jobId}`);
    },
  });

  // Transport 삭제 mutation
  const deleteMutation = useMutation({
    mutationFn: () => deleteTransport(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: transportQueryKeys.all });
      router.push('/transports');
    },
  });

  // 삭제 확인 핸들러
  const handleDeleteConfirm = () => {
    deleteMutation.mutate();
  };

  return (
    <MainLayout>
      <div className="space-y-6">
        {/* 뒤로가기 버튼 */}
        <div className="flex items-center gap-4">
          <Link href="/transports">
            <Button variant="ghost" size="sm">
              <ChevronLeft className="mr-1 h-4 w-4" />
              목록으로
            </Button>
          </Link>
        </div>

        {/* 로딩 상태 */}
        {isLoading && <DetailSkeleton />}

        {/* 에러 상태 */}
        {isError && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-destructive text-sm">
              Transport를 불러오는데 실패했습니다.
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

        {/* Transport 상세 정보 */}
        {transport && (
          <>
            {/* 헤더 */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div>
                  <h1 className="text-2xl font-bold tracking-tight">
                    {transport.name}
                  </h1>
                  {transport.description && (
                    <p className="text-muted-foreground mt-1">
                      {transport.description}
                    </p>
                  )}
                </div>
                <TransportStatusBadge status={transport.status} />
              </div>

              {/* 액션 버튼 */}
              <div className="flex items-center gap-2">
                {/* 실행 버튼 (FR-02.8) */}
                <Button
                  onClick={() => executeMutation.mutate()}
                  disabled={
                    transport.status === 'running' || executeMutation.isPending
                  }
                >
                  <Play className="mr-2 h-4 w-4" />
                  {executeMutation.isPending ? '실행 중...' : '실행'}
                </Button>

                {/* 삭제 버튼 (FR-02.6, FR-02.7) */}
                <Button
                  variant="outline"
                  onClick={() => setShowDeleteDialog(true)}
                  disabled={transport.status === 'running'}
                  className="text-destructive hover:text-destructive"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  삭제
                </Button>
              </div>
            </div>

            {/* 정보 카드 그리드 */}
            <div className="grid gap-6 md:grid-cols-2">
              {/* 소스 정보 */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-base">
                    <Database className="h-4 w-4" />
                    소스 정보
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <InfoItem
                    icon={Database}
                    label="스키마"
                    value={transport.sourceSchema}
                    mono
                  />
                  <InfoItem
                    icon={Database}
                    label="테이블"
                    value={transport.sourceTable}
                    mono
                  />
                </CardContent>
              </Card>

              {/* 대상 정보 */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-base">
                    <FolderOutput className="h-4 w-4" />
                    대상 정보
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <InfoItem
                    icon={FolderOutput}
                    label="경로"
                    value={transport.targetPath}
                    mono
                  />
                  <InfoItem
                    icon={FolderOutput}
                    label="형식"
                    value={
                      <Badge variant="outline">
                        {FORMAT_LABELS[transport.targetFormat] ||
                          transport.targetFormat.toUpperCase()}
                      </Badge>
                    }
                  />
                </CardContent>
              </Card>
            </div>

            {/* 설정 정보 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <Settings className="h-4 w-4" />
                  설정
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-4">
                  <div className="rounded-md bg-muted p-3">
                    <p className="text-xs text-muted-foreground">배치 크기</p>
                    <p className="font-mono text-sm font-medium">
                      {formatNumber(transport.config.batchSize)}
                    </p>
                  </div>
                  <div className="rounded-md bg-muted p-3">
                    <p className="text-xs text-muted-foreground">병렬 처리</p>
                    <p className="font-mono text-sm font-medium">
                      {transport.config.parallelism}
                    </p>
                  </div>
                  <div className="rounded-md bg-muted p-3">
                    <p className="text-xs text-muted-foreground">재시도 횟수</p>
                    <p className="font-mono text-sm font-medium">
                      {transport.config.retryAttempts}
                    </p>
                  </div>
                  <div className="rounded-md bg-muted p-3">
                    <p className="text-xs text-muted-foreground">
                      타임아웃 (초)
                    </p>
                    <p className="font-mono text-sm font-medium">
                      {transport.config.timeout / 1000}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* 시간 정보 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <Clock className="h-4 w-4" />
                  실행 정보
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 sm:grid-cols-3">
                  <InfoItem
                    icon={Calendar}
                    label="생성일"
                    value={formatDate(transport.createdAt)}
                  />
                  <InfoItem
                    icon={Calendar}
                    label="수정일"
                    value={formatDate(transport.updatedAt)}
                  />
                  <InfoItem
                    icon={Clock}
                    label="마지막 실행"
                    value={
                      transport.lastRunAt
                        ? formatDate(transport.lastRunAt)
                        : '실행 기록 없음'
                    }
                  />
                </div>
              </CardContent>
            </Card>

            {/* 삭제 확인 다이얼로그 */}
            <DeleteConfirmDialog
              open={showDeleteDialog}
              onOpenChange={setShowDeleteDialog}
              transportName={transport.name}
              onConfirm={handleDeleteConfirm}
              isDeleting={deleteMutation.isPending}
            />
          </>
        )}
      </div>
    </MainLayout>
  );
}
