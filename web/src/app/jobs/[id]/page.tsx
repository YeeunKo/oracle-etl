/**
 * Job 상세/모니터링 페이지 (FR-03)
 * 실시간 진행 상황, 로그, 에러 표시
 */

'use client';

import { useState, use } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { getJob, cancelJob, retryJob } from '@/lib/api';
import { useJobSSE } from '@/hooks/useJobSSE';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { JobStatusBadge } from '@/components/job/job-status-badge';
import { JobProgress } from '@/components/job/job-progress';
import { JobLogs } from '@/components/job/job-logs';
import { formatDateCompact, formatDuration, formatBytes, formatNumber } from '@/lib/format';
import type { Job, JobStatus } from '@/types/api';
import {
  ArrowLeft,
  RefreshCw,
  XCircle,
  RotateCcw,
  AlertTriangle,
  Info,
  Clock,
  Database,
  HardDrive,
  Rows3,
  Wifi,
  WifiOff,
} from 'lucide-react';
import { toast } from 'sonner';

// ============================================================================
// Query Keys
// ============================================================================

const jobDetailQueryKeys = {
  detail: (id: string) => ['jobs', 'detail', id] as const,
};

// ============================================================================
// 타입
// ============================================================================

interface JobDetailPageProps {
  params: Promise<{ id: string }>;
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

/**
 * Job이 활성 상태인지 확인
 */
function isJobActive(status: JobStatus): boolean {
  return status === 'pending' || status === 'running';
}

// ============================================================================
// 서브 컴포넌트
// ============================================================================

/**
 * 로딩 스켈레톤
 */
function LoadingSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Skeleton className="h-8 w-8" />
        <Skeleton className="h-8 w-64" />
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Skeleton className="h-96" />
        <Skeleton className="h-96" />
      </div>
    </div>
  );
}

/**
 * SSE 연결 상태 표시
 */
function ConnectionStatus({
  state,
}: {
  state: 'connecting' | 'connected' | 'disconnected' | 'error';
}) {
  const config = {
    connecting: {
      icon: Wifi,
      label: '연결 중...',
      className: 'text-yellow-500',
    },
    connected: {
      icon: Wifi,
      label: '실시간 연결됨',
      className: 'text-green-500',
    },
    disconnected: {
      icon: WifiOff,
      label: '연결 끊김',
      className: 'text-muted-foreground',
    },
    error: {
      icon: WifiOff,
      label: '연결 오류',
      className: 'text-red-500',
    },
  };

  const { icon: Icon, label, className } = config[state];

  return (
    <div className={`flex items-center gap-1.5 text-xs ${className}`}>
      <Icon className="h-3.5 w-3.5" />
      <span>{label}</span>
    </div>
  );
}

/**
 * Job 기본 정보 카드
 */
function JobInfoCard({ job }: { job: Job }) {
  const duration = calculateDuration(job);

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-semibold">작업 정보</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground">Job ID</p>
            <p className="text-sm font-mono">{job.id}</p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground">Transport</p>
            <Link
              href={`/transports/${job.transportId}`}
              className="text-sm font-medium text-primary hover:underline"
            >
              {job.transportName}
            </Link>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground">시작 시간</p>
            <p className="text-sm">{formatDateCompact(job.startedAt)}</p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground">완료 시간</p>
            <p className="text-sm">
              {job.completedAt ? formatDateCompact(job.completedAt) : '-'}
            </p>
          </div>
        </div>

        {/* 통계 */}
        <div className="grid grid-cols-2 gap-3 pt-3 border-t">
          <div className="flex items-center gap-2 p-2 rounded-lg bg-muted/50">
            <Clock className="h-4 w-4 text-muted-foreground" />
            <div>
              <p className="text-xs text-muted-foreground">실행 시간</p>
              <p className="text-sm font-semibold">{formatDuration(duration)}</p>
            </div>
          </div>
          <div className="flex items-center gap-2 p-2 rounded-lg bg-muted/50">
            <Rows3 className="h-4 w-4 text-muted-foreground" />
            <div>
              <p className="text-xs text-muted-foreground">처리된 행</p>
              <p className="text-sm font-semibold">
                {formatNumber(job.progress.rowsProcessed)}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2 p-2 rounded-lg bg-muted/50">
            <HardDrive className="h-4 w-4 text-muted-foreground" />
            <div>
              <p className="text-xs text-muted-foreground">데이터 크기</p>
              <p className="text-sm font-semibold">
                {formatBytes(job.progress.bytesProcessed)}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2 p-2 rounded-lg bg-muted/50">
            <Database className="h-4 w-4 text-muted-foreground" />
            <div>
              <p className="text-xs text-muted-foreground">추출 테이블</p>
              <p className="text-sm font-semibold">{job.extractions?.length || 0}개</p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * 에러 표시 컴포넌트
 */
function ErrorDisplay({
  error,
  onRetry,
  isRetrying,
}: {
  error: string;
  onRetry: () => void;
  isRetrying: boolean;
}) {
  return (
    <Alert variant="destructive">
      <AlertTriangle className="h-4 w-4" />
      <AlertTitle>작업 실패</AlertTitle>
      <AlertDescription className="mt-2">
        <p className="mb-3">{error}</p>
        <Button
          variant="outline"
          size="sm"
          onClick={onRetry}
          disabled={isRetrying}
        >
          {isRetrying ? (
            <>
              <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
              재시도 중...
            </>
          ) : (
            <>
              <RotateCcw className="h-4 w-4 mr-2" />
              재시도
            </>
          )}
        </Button>
      </AlertDescription>
    </Alert>
  );
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * Job 상세/모니터링 페이지
 * FR-03.1: 진행 상황 (0-100%)
 * FR-03.2: SSE 기반 실시간 업데이트
 * FR-03.3: 진행 단계 표시
 * FR-03.4: 처리 행/바이트 실시간 표시
 * FR-03.5: 실시간 로그 뷰어
 * FR-03.6: 에러 표시 및 재시도
 * FR-03.7: 작업 취소 버튼
 */
export default function JobDetailPage({ params }: JobDetailPageProps) {
  const { id: jobId } = use(params);
  const router = useRouter();
  const queryClient = useQueryClient();

  // Job 상태
  const [activeTab, setActiveTab] = useState('progress');

  // Job 데이터 조회
  const {
    data: job,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: jobDetailQueryKeys.detail(jobId),
    queryFn: () => getJob(jobId),
    staleTime: 5 * 1000, // 5초
    refetchInterval: (data) => {
      // 활성 상태일 때만 자동 갱신
      if (data.state.data && isJobActive(data.state.data.status)) {
        return 10 * 1000; // 10초
      }
      return false;
    },
  });

  // SSE 연결 (활성 상태일 때만)
  const shouldConnectSSE = job && isJobActive(job.status);

  const {
    status: sseStatus,
    progress: sseProgress,
    logs,
    connectionState,
    error: sseError,
    clearLogs,
  } = useJobSSE(shouldConnectSSE ? jobId : null, {
    onComplete: () => {
      // 완료 시 데이터 새로고침
      queryClient.invalidateQueries({ queryKey: jobDetailQueryKeys.detail(jobId) });
      toast.success('작업이 완료되었습니다.');
    },
    onError: (err) => {
      queryClient.invalidateQueries({ queryKey: jobDetailQueryKeys.detail(jobId) });
      toast.error(`작업 실패: ${err}`);
    },
  });

  // 실시간 업데이트된 진행 상황 사용
  const currentStatus = sseStatus || job?.status;
  const currentProgress = sseProgress || job?.progress;

  // Job 취소 mutation
  const cancelMutation = useMutation({
    mutationFn: () => cancelJob(jobId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: jobDetailQueryKeys.detail(jobId) });
      toast.success('작업이 취소되었습니다.');
    },
    onError: (err) => {
      toast.error(`취소 실패: ${err instanceof Error ? err.message : '알 수 없는 오류'}`);
    },
  });

  // Job 재시도 mutation
  const retryMutation = useMutation({
    mutationFn: () => retryJob(jobId),
    onSuccess: (data) => {
      toast.success('작업이 다시 시작되었습니다.');
      router.push(`/jobs/${data.jobId}`);
    },
    onError: (err) => {
      toast.error(`재시도 실패: ${err instanceof Error ? err.message : '알 수 없는 오류'}`);
    },
  });

  // 로딩 상태
  if (isLoading) {
    return (
      <div className="container mx-auto py-6">
        <LoadingSkeleton />
      </div>
    );
  }

  // 에러 상태
  if (isError) {
    return (
      <div className="container mx-auto py-6">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>작업을 불러올 수 없습니다</AlertTitle>
          <AlertDescription>
            <p className="mb-3">
              {error instanceof Error ? error.message : '알 수 없는 오류가 발생했습니다.'}
            </p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => refetch()}>
                다시 시도
              </Button>
              <Button variant="outline" size="sm" onClick={() => router.push('/jobs')}>
                목록으로
              </Button>
            </div>
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  if (!job) {
    return null;
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* 페이지 헤더 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="sm" onClick={() => router.push('/jobs')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            목록으로
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-bold tracking-tight">
                작업 상세
              </h1>
              <JobStatusBadge status={currentStatus || job.status} />
              {shouldConnectSSE && (
                <ConnectionStatus state={connectionState} />
              )}
            </div>
            <p className="text-sm text-muted-foreground mt-1">
              {job.transportName} - {job.id.slice(0, 8)}
            </p>
          </div>
        </div>

        {/* 액션 버튼 */}
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            새로고침
          </Button>

          {/* 취소 버튼 (활성 상태일 때만) */}
          {isJobActive(job.status) && (
            <Button
              variant="destructive"
              size="sm"
              onClick={() => cancelMutation.mutate()}
              disabled={cancelMutation.isPending}
            >
              {cancelMutation.isPending ? (
                <>
                  <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                  취소 중...
                </>
              ) : (
                <>
                  <XCircle className="h-4 w-4 mr-2" />
                  작업 취소
                </>
              )}
            </Button>
          )}

          {/* 재시도 버튼 (실패 상태일 때만) */}
          {job.status === 'failed' && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => retryMutation.mutate()}
              disabled={retryMutation.isPending}
            >
              {retryMutation.isPending ? (
                <>
                  <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                  재시도 중...
                </>
              ) : (
                <>
                  <RotateCcw className="h-4 w-4 mr-2" />
                  재시도
                </>
              )}
            </Button>
          )}
        </div>
      </div>

      {/* 에러 표시 */}
      {job.status === 'failed' && job.error && (
        <ErrorDisplay
          error={job.error}
          onRetry={() => retryMutation.mutate()}
          isRetrying={retryMutation.isPending}
        />
      )}

      {/* SSE 연결 에러 */}
      {sseError && (
        <Alert>
          <Info className="h-4 w-4" />
          <AlertTitle>실시간 연결 문제</AlertTitle>
          <AlertDescription>
            {sseError}. 페이지를 새로고침하거나 잠시 후 다시 시도해 주세요.
          </AlertDescription>
        </Alert>
      )}

      {/* 탭 콘텐츠 */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="progress">진행 상황</TabsTrigger>
          <TabsTrigger value="logs">
            로그
            {logs.length > 0 && (
              <span className="ml-1.5 px-1.5 py-0.5 text-xs bg-muted rounded-full">
                {logs.length}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="details">상세 정보</TabsTrigger>
        </TabsList>

        <TabsContent value="progress" className="mt-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <JobProgress progress={currentProgress || null} />
            <JobInfoCard job={job} />
          </div>
        </TabsContent>

        <TabsContent value="logs" className="mt-4">
          <JobLogs
            logs={logs.length > 0 ? logs : (job.logs || []).map((msg, i) => ({
              id: `static-${i}`,
              level: 'info' as const,
              message: msg,
              timestamp: job.startedAt,
            }))}
            onClear={logs.length > 0 ? clearLogs : undefined}
            maxHeight={500}
          />
        </TabsContent>

        <TabsContent value="details" className="mt-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <JobInfoCard job={job} />

            {/* Extraction 정보 */}
            {job.extractions && job.extractions.length > 0 && (
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base font-semibold">
                    추출 테이블 ({job.extractions.length})
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {job.extractions.map((extraction) => (
                      <div
                        key={extraction.id}
                        className="flex items-center justify-between p-3 rounded-lg bg-muted/50"
                      >
                        <div className="flex items-center gap-3">
                          <Database className="h-4 w-4 text-muted-foreground" />
                          <div>
                            <p className="text-sm font-medium">
                              {extraction.tableName}
                            </p>
                            <p className="text-xs text-muted-foreground">
                              {formatNumber(extraction.rowsExtracted)} 행 |{' '}
                              {formatBytes(extraction.bytesExtracted)}
                            </p>
                          </div>
                        </div>
                        <JobStatusBadge status={extraction.status} showIcon={false} />
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
