/**
 * 헤더 컴포넌트
 * 연결 상태 표시, 검색, 사용자 메뉴 제공
 */

'use client';

import { Menu, Bell, User, Wifi, WifiOff, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useConnectionStatus } from '@/hooks/useConnectionStatus';
import { formatRelativeTime } from '@/lib/format';

// ============================================================================
// 연결 상태 배지 컴포넌트
// ============================================================================

/**
 * 연결 상태 배지
 */
function ConnectionStatusBadge() {
  const { isConnected, isLoading, oracle, lastChecked } = useConnectionStatus();

  // 로딩 중
  if (isLoading) {
    return (
      <div className="flex items-center gap-2 rounded-full bg-muted px-3 py-1.5 text-sm">
        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        <span className="text-muted-foreground">연결 확인 중...</span>
      </div>
    );
  }

  // 연결됨
  if (isConnected) {
    return (
      <div
        className="flex items-center gap-2 rounded-full bg-green-500/10 px-3 py-1.5 text-sm"
        title={`Oracle ${oracle?.version || ''}\n마지막 확인: ${lastChecked ? formatRelativeTime(lastChecked.toISOString()) : '-'}`}
      >
        <Wifi className="h-4 w-4 text-green-600" />
        <span className="text-green-700 dark:text-green-400">연결됨</span>
        {oracle?.database && (
          <span className="text-green-600/70 dark:text-green-500/70">
            ({oracle.database})
          </span>
        )}
      </div>
    );
  }

  // 연결 안됨
  return (
    <div
      className="flex items-center gap-2 rounded-full bg-destructive/10 px-3 py-1.5 text-sm"
      title={oracle?.error || '서버에 연결할 수 없습니다'}
    >
      <WifiOff className="h-4 w-4 text-destructive" />
      <span className="text-destructive">연결 안됨</span>
    </div>
  );
}

// ============================================================================
// 헤더 컴포넌트
// ============================================================================

/**
 * 헤더 Props
 */
interface HeaderProps {
  /** 추가 클래스명 */
  className?: string;
  /** 모바일 메뉴 토글 콜백 */
  onMenuToggle?: () => void;
}

/**
 * 헤더 컴포넌트
 *
 * @param props - 컴포넌트 Props
 * @returns 헤더 컴포넌트
 */
export function Header({ className, onMenuToggle }: HeaderProps) {
  return (
    <header
      className={cn(
        'flex h-16 items-center justify-between border-b border-border bg-background px-4 md:px-6',
        className
      )}
    >
      {/* 왼쪽: 모바일 메뉴 버튼 + 타이틀 */}
      <div className="flex items-center gap-4">
        {/* 모바일 메뉴 버튼 */}
        <button
          onClick={onMenuToggle}
          className="md:hidden rounded-md p-2 hover:bg-accent"
          aria-label="메뉴 열기"
        >
          <Menu className="h-5 w-5" />
        </button>

        {/* 페이지 타이틀 (필요 시 동적으로 변경) */}
        <h1 className="text-lg font-semibold text-foreground md:hidden">
          Oracle ETL
        </h1>
      </div>

      {/* 오른쪽: 상태 표시 + 액션 */}
      <div className="flex items-center gap-4">
        {/* Oracle 연결 상태 */}
        <ConnectionStatusBadge />

        {/* 알림 버튼 */}
        <button
          className="rounded-md p-2 hover:bg-accent transition-colors"
          aria-label="알림"
        >
          <Bell className="h-5 w-5 text-muted-foreground" />
        </button>

        {/* 사용자 메뉴 */}
        <button
          className="rounded-md p-2 hover:bg-accent transition-colors"
          aria-label="사용자 메뉴"
        >
          <User className="h-5 w-5 text-muted-foreground" />
        </button>
      </div>
    </header>
  );
}
