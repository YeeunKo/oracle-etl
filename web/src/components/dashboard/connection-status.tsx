/**
 * Oracle 연결 상태 카드 컴포넌트
 * 서버 헬스와 Oracle DB 연결 상태를 표시
 */

'use client';

import { useConnectionStatus } from '@/hooks/useConnectionStatus';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { formatRelativeTime } from '@/lib/format';

// ============================================================================
// 상수
// ============================================================================

/**
 * 연결 상태 타입
 */
type ConnectionState = 'connected' | 'disconnected' | 'checking';

/**
 * 상태별 스타일 설정
 */
const STATUS_STYLES: Record<ConnectionState, {
  badge: 'default' | 'secondary' | 'destructive' | 'outline';
  dot: string;
  label: string;
}> = {
  connected: {
    badge: 'default',
    dot: 'bg-green-500',
    label: '연결됨',
  },
  disconnected: {
    badge: 'destructive',
    dot: 'bg-red-500',
    label: '연결 끊김',
  },
  checking: {
    badge: 'secondary',
    dot: 'bg-yellow-500',
    label: '확인 중...',
  },
};

// ============================================================================
// 서브 컴포넌트
// ============================================================================

/**
 * 상태 표시 뱃지
 */
function StatusBadge({ state }: { state: ConnectionState }) {
  const style = STATUS_STYLES[state];

  return (
    <Badge variant={style.badge} className="gap-1.5">
      <span className={`h-2 w-2 rounded-full ${style.dot} animate-pulse`} />
      {style.label}
    </Badge>
  );
}

/**
 * 연결 정보 행
 */
function InfoRow({ label, value }: { label: string; value: string | undefined }) {
  return (
    <div className="flex items-center justify-between text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value || '-'}</span>
    </div>
  );
}

/**
 * 로딩 스켈레톤
 */
function ConnectionStatusSkeleton() {
  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <Skeleton className="h-5 w-32" />
          <Skeleton className="h-5 w-20" />
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-2/3" />
      </CardContent>
    </Card>
  );
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * Oracle 연결 상태 카드
 * 서버 헬스 및 Oracle DB 연결 상태를 표시하고 자동 갱신
 *
 * @example
 * <ConnectionStatus />
 */
export function ConnectionStatus() {
  const {
    isConnected,
    isLoading,
    isError,
    health,
    oracle,
    error,
    lastChecked,
  } = useConnectionStatus();

  // 로딩 상태
  if (isLoading && !oracle) {
    return <ConnectionStatusSkeleton />;
  }

  // 현재 상태 계산
  const getConnectionState = (): ConnectionState => {
    if (isLoading) return 'checking';
    if (isError || !isConnected) return 'disconnected';
    return 'connected';
  };

  const connectionState = getConnectionState();

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-semibold">
            Oracle 연결 상태
          </CardTitle>
          <StatusBadge state={connectionState} />
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {/* 서버 헬스 상태 */}
        <InfoRow
          label="서버 상태"
          value={health?.status === 'ok' ? '정상' : '오류'}
        />

        {/* Oracle 연결 정보 */}
        <InfoRow
          label="Oracle 버전"
          value={oracle?.version}
        />

        <InfoRow
          label="호스트"
          value={oracle?.host}
        />

        <InfoRow
          label="데이터베이스"
          value={oracle?.database}
        />

        {/* 마지막 확인 시간 */}
        <div className="flex items-center justify-between text-sm pt-2 border-t">
          <span className="text-muted-foreground">마지막 확인</span>
          <span className="text-xs text-muted-foreground">
            {lastChecked ? formatRelativeTime(lastChecked.toISOString()) : '-'}
          </span>
        </div>

        {/* 에러 메시지 */}
        {isError && error && (
          <div className="mt-3 p-3 rounded-md bg-destructive/10 text-destructive text-sm">
            <p className="font-medium">연결 오류</p>
            <p className="text-xs mt-1">{error}</p>
          </div>
        )}

        {/* Oracle 에러 메시지 */}
        {oracle?.error && (
          <div className="mt-3 p-3 rounded-md bg-destructive/10 text-destructive text-sm">
            <p className="font-medium">Oracle 오류</p>
            <p className="text-xs mt-1">{oracle.error}</p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
