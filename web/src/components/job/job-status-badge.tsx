/**
 * Job 상태 뱃지 컴포넌트
 * Job 상태에 따른 시각적 표시를 제공
 */

import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { JobStatus } from '@/types/api';
import {
  Clock,
  Loader2,
  CheckCircle,
  XCircle,
  Ban,
} from 'lucide-react';

// ============================================================================
// 상수
// ============================================================================

/**
 * Job 상태별 스타일 및 레이블 설정
 */
const JOB_STATUS_CONFIG: Record<
  JobStatus,
  {
    variant: 'default' | 'secondary' | 'destructive' | 'outline';
    label: string;
    icon: React.ElementType;
    className?: string;
  }
> = {
  pending: {
    variant: 'secondary',
    label: '대기 중',
    icon: Clock,
  },
  running: {
    variant: 'default',
    label: '실행 중',
    icon: Loader2,
    className: 'animate-spin',
  },
  completed: {
    variant: 'outline',
    label: '완료',
    icon: CheckCircle,
    className: 'text-green-600',
  },
  failed: {
    variant: 'destructive',
    label: '실패',
    icon: XCircle,
  },
  cancelled: {
    variant: 'secondary',
    label: '취소됨',
    icon: Ban,
  },
};

// ============================================================================
// 타입
// ============================================================================

interface JobStatusBadgeProps {
  /** Job 상태 */
  status: JobStatus;
  /** 아이콘 표시 여부 */
  showIcon?: boolean;
  /** 추가 클래스명 */
  className?: string;
}

// ============================================================================
// 컴포넌트
// ============================================================================

/**
 * Job 상태 뱃지
 * 상태에 따른 색상과 아이콘을 표시
 *
 * @example
 * <JobStatusBadge status="running" />
 * <JobStatusBadge status="completed" showIcon />
 */
export function JobStatusBadge({
  status,
  showIcon = true,
  className,
}: JobStatusBadgeProps) {
  const config = JOB_STATUS_CONFIG[status] || {
    variant: 'secondary' as const,
    label: status,
    icon: Clock,
  };

  const Icon = config.icon;

  return (
    <Badge variant={config.variant} className={cn('text-xs gap-1', className)}>
      {showIcon && (
        <Icon className={cn('h-3 w-3', config.className)} />
      )}
      {config.label}
    </Badge>
  );
}

/**
 * Job 상태 레이블 반환
 */
export function getJobStatusLabel(status: JobStatus): string {
  return JOB_STATUS_CONFIG[status]?.label || status;
}

/**
 * Job 상태 variant 반환
 */
export function getJobStatusVariant(status: JobStatus): 'default' | 'secondary' | 'destructive' | 'outline' {
  return JOB_STATUS_CONFIG[status]?.variant || 'secondary';
}
