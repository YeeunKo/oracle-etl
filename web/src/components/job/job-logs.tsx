/**
 * Job 로그 뷰어 컴포넌트
 * 실시간 로그를 자동 스크롤과 함께 표시
 */

'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { formatDateCompact } from '@/lib/format';
import type { SSELogEntry } from '@/hooks/useJobSSE';
import {
  ScrollText,
  ArrowDown,
  Trash2,
  Pause,
  Play,
  Info,
  AlertTriangle,
  XCircle,
} from 'lucide-react';

// ============================================================================
// 상수
// ============================================================================

/**
 * 로그 레벨별 설정
 */
const LOG_LEVEL_CONFIG: Record<
  SSELogEntry['level'],
  {
    icon: React.ElementType;
    color: string;
    bgColor: string;
  }
> = {
  info: {
    icon: Info,
    color: 'text-blue-500',
    bgColor: 'bg-blue-500/10',
  },
  warn: {
    icon: AlertTriangle,
    color: 'text-yellow-500',
    bgColor: 'bg-yellow-500/10',
  },
  error: {
    icon: XCircle,
    color: 'text-red-500',
    bgColor: 'bg-red-500/10',
  },
};

// ============================================================================
// 타입
// ============================================================================

interface JobLogsProps {
  /** 로그 항목 목록 */
  logs: SSELogEntry[];
  /** 로그 초기화 함수 */
  onClear?: () => void;
  /** 최대 높이 (기본값: 400px) */
  maxHeight?: number;
  /** 추가 클래스명 */
  className?: string;
}

// ============================================================================
// 서브 컴포넌트
// ============================================================================

/**
 * 단일 로그 항목
 */
function LogEntry({ log }: { log: SSELogEntry }) {
  const config = LOG_LEVEL_CONFIG[log.level];
  const Icon = config.icon;

  return (
    <div
      className={cn(
        'flex items-start gap-3 py-2 px-3 rounded-md transition-colors',
        config.bgColor
      )}
    >
      <Icon className={cn('h-4 w-4 mt-0.5 shrink-0', config.color)} />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-mono break-all">{log.message}</p>
      </div>
      <span className="text-xs text-muted-foreground shrink-0">
        {formatDateCompact(log.timestamp).split(' ')[1] || formatDateCompact(log.timestamp)}
      </span>
    </div>
  );
}

/**
 * 빈 상태
 */
function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <ScrollText className="h-10 w-10 text-muted-foreground/50" />
      <p className="mt-3 text-sm text-muted-foreground">
        로그가 없습니다.
      </p>
      <p className="text-xs text-muted-foreground mt-1">
        작업이 시작되면 로그가 여기에 표시됩니다.
      </p>
    </div>
  );
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * Job 로그 뷰어 컴포넌트
 * 실시간 로그를 자동 스크롤과 함께 표시
 *
 * @example
 * <JobLogs logs={logs} onClear={() => clearLogs()} />
 */
export function JobLogs({
  logs,
  onClear,
  maxHeight = 400,
  className,
}: JobLogsProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [isAtBottom, setIsAtBottom] = useState(true);

  /**
   * 스크롤 위치 확인
   */
  const checkScrollPosition = useCallback(() => {
    if (!scrollRef.current) return;

    const { scrollTop, scrollHeight, clientHeight } = scrollRef.current;
    const threshold = 50; // 하단으로부터 50px 이내면 하단으로 간주
    const atBottom = scrollHeight - scrollTop - clientHeight < threshold;
    setIsAtBottom(atBottom);
  }, []);

  /**
   * 하단으로 스크롤
   */
  const scrollToBottom = useCallback(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, []);

  // 새 로그가 추가되면 자동 스크롤
  useEffect(() => {
    if (autoScroll && scrollRef.current) {
      scrollToBottom();
    }
  }, [logs, autoScroll, scrollToBottom]);

  // 스크롤 이벤트 핸들러
  useEffect(() => {
    const scrollElement = scrollRef.current;
    if (!scrollElement) return;

    const handleScroll = () => {
      checkScrollPosition();
      // 사용자가 위로 스크롤하면 자동 스크롤 중지
      if (!isAtBottom) {
        setAutoScroll(false);
      }
    };

    scrollElement.addEventListener('scroll', handleScroll);
    return () => scrollElement.removeEventListener('scroll', handleScroll);
  }, [isAtBottom, checkScrollPosition]);

  // 로그 레벨별 카운트
  const logCounts = logs.reduce(
    (acc, log) => {
      acc[log.level] = (acc[log.level] || 0) + 1;
      return acc;
    },
    {} as Record<SSELogEntry['level'], number>
  );

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <CardTitle className="text-base font-semibold">로그</CardTitle>
            {/* 로그 카운트 뱃지 */}
            <div className="flex items-center gap-1">
              {logCounts.info > 0 && (
                <Badge variant="secondary" className="text-xs">
                  정보 {logCounts.info}
                </Badge>
              )}
              {logCounts.warn > 0 && (
                <Badge variant="outline" className="text-xs text-yellow-600 border-yellow-300">
                  경고 {logCounts.warn}
                </Badge>
              )}
              {logCounts.error > 0 && (
                <Badge variant="destructive" className="text-xs">
                  에러 {logCounts.error}
                </Badge>
              )}
            </div>
          </div>

          {/* 제어 버튼 */}
          <div className="flex items-center gap-2">
            {/* 자동 스크롤 토글 */}
            <Button
              variant={autoScroll ? 'secondary' : 'outline'}
              size="sm"
              onClick={() => {
                setAutoScroll(!autoScroll);
                if (!autoScroll) {
                  scrollToBottom();
                }
              }}
              title={autoScroll ? '자동 스크롤 일시정지' : '자동 스크롤 재개'}
            >
              {autoScroll ? (
                <Pause className="h-3.5 w-3.5" />
              ) : (
                <Play className="h-3.5 w-3.5" />
              )}
            </Button>

            {/* 하단으로 스크롤 */}
            {!isAtBottom && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  scrollToBottom();
                  setAutoScroll(true);
                }}
                title="하단으로 스크롤"
              >
                <ArrowDown className="h-3.5 w-3.5" />
              </Button>
            )}

            {/* 로그 초기화 */}
            {onClear && logs.length > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onClear}
                title="로그 초기화"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            )}
          </div>
        </div>
      </CardHeader>

      <CardContent>
        {logs.length === 0 ? (
          <EmptyState />
        ) : (
          <div
            ref={scrollRef}
            className="space-y-1 overflow-y-auto"
            style={{ maxHeight }}
          >
            {logs.map((log) => (
              <LogEntry key={log.id} log={log} />
            ))}
          </div>
        )}

        {/* 총 로그 수 */}
        {logs.length > 0 && (
          <div className="mt-3 pt-3 border-t">
            <p className="text-xs text-muted-foreground text-center">
              총 {logs.length}개의 로그 항목
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
