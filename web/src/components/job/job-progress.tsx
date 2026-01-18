/**
 * Job 진행 상황 컴포넌트
 * 실시간 진행률과 통계를 표시
 */

'use client';

import { Progress } from '@/components/ui/progress';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { formatBytes, formatDuration, formatNumber } from '@/lib/format';
import type { JobProgress as JobProgressType } from '@/types/api';
import {
  Loader2,
  Database,
  ArrowRightLeft,
  Upload,
  CheckCircle,
  Clock,
  HardDrive,
  Rows3,
} from 'lucide-react';

// ============================================================================
// 상수
// ============================================================================

/**
 * 진행 단계별 설정
 */
const PHASE_CONFIG: Record<
  JobProgressType['phase'],
  {
    label: string;
    icon: React.ElementType;
    color: string;
  }
> = {
  initializing: {
    label: '초기화 중',
    icon: Loader2,
    color: 'text-blue-500',
  },
  extracting: {
    label: '데이터 추출 중',
    icon: Database,
    color: 'text-orange-500',
  },
  transforming: {
    label: '데이터 변환 중',
    icon: ArrowRightLeft,
    color: 'text-purple-500',
  },
  loading: {
    label: '데이터 로딩 중',
    icon: Upload,
    color: 'text-green-500',
  },
  completed: {
    label: '완료',
    icon: CheckCircle,
    color: 'text-green-600',
  },
};

// ============================================================================
// 타입
// ============================================================================

interface JobProgressProps {
  /** 진행 상황 데이터 */
  progress: JobProgressType | null;
  /** 추가 클래스명 */
  className?: string;
}

interface StatCardProps {
  icon: React.ElementType;
  label: string;
  value: string;
  className?: string;
}

// ============================================================================
// 서브 컴포넌트
// ============================================================================

/**
 * 통계 카드
 */
function StatCard({ icon: Icon, label, value, className }: StatCardProps) {
  return (
    <div className={cn('flex items-center gap-3 p-3 rounded-lg bg-muted/50', className)}>
      <div className="p-2 rounded-md bg-background">
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>
      <div>
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className="text-sm font-semibold">{value}</p>
      </div>
    </div>
  );
}

/**
 * 진행 단계 표시기
 */
function PhaseIndicator({ phase }: { phase: JobProgressType['phase'] }) {
  const config = PHASE_CONFIG[phase];
  const Icon = config.icon;
  const isAnimated = phase !== 'completed';

  return (
    <div className="flex items-center gap-2">
      <Icon
        className={cn(
          'h-5 w-5',
          config.color,
          isAnimated && 'animate-spin'
        )}
      />
      <span className={cn('text-sm font-medium', config.color)}>
        {config.label}
      </span>
    </div>
  );
}

/**
 * 진행 단계 스텝퍼
 */
function PhaseStepper({ currentPhase }: { currentPhase: JobProgressType['phase'] }) {
  const phases: JobProgressType['phase'][] = [
    'initializing',
    'extracting',
    'transforming',
    'loading',
    'completed',
  ];

  const currentIndex = phases.indexOf(currentPhase);

  return (
    <div className="flex items-center justify-between w-full">
      {phases.map((phase, index) => {
        const config = PHASE_CONFIG[phase];
        const Icon = config.icon;
        const isActive = index === currentIndex;
        const isCompleted = index < currentIndex;

        return (
          <div key={phase} className="flex items-center flex-1 last:flex-none">
            {/* 스텝 아이콘 */}
            <div
              className={cn(
                'flex items-center justify-center w-8 h-8 rounded-full border-2 transition-colors',
                isActive && 'border-primary bg-primary text-primary-foreground',
                isCompleted && 'border-green-500 bg-green-500 text-white',
                !isActive && !isCompleted && 'border-muted-foreground/30 text-muted-foreground/50'
              )}
            >
              <Icon className={cn('h-4 w-4', isActive && 'animate-spin')} />
            </div>

            {/* 연결선 */}
            {index < phases.length - 1 && (
              <div
                className={cn(
                  'flex-1 h-0.5 mx-2',
                  isCompleted ? 'bg-green-500' : 'bg-muted-foreground/20'
                )}
              />
            )}
          </div>
        );
      })}
    </div>
  );
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * Job 진행 상황 표시 컴포넌트
 *
 * @example
 * <JobProgress progress={job.progress} />
 */
export function JobProgress({ progress, className }: JobProgressProps) {
  // 진행 상황 없을 때
  if (!progress) {
    return (
      <Card className={className}>
        <CardContent className="py-8">
          <div className="flex flex-col items-center justify-center text-center">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            <p className="mt-2 text-sm text-muted-foreground">
              진행 상황을 불러오는 중...
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const {
    phase,
    percentage,
    currentTable,
    rowsProcessed,
    totalRows,
    bytesProcessed,
    estimatedTimeRemaining,
  } = progress;

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-semibold">진행 상황</CardTitle>
          <PhaseIndicator phase={phase} />
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* 진행 단계 스텝퍼 */}
        <div className="py-2">
          <PhaseStepper currentPhase={phase} />
          <div className="flex justify-between mt-2 text-xs text-muted-foreground">
            <span>초기화</span>
            <span>추출</span>
            <span>변환</span>
            <span>로딩</span>
            <span>완료</span>
          </div>
        </div>

        {/* 프로그레스 바 */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">전체 진행률</span>
            <span className="font-semibold">{percentage.toFixed(1)}%</span>
          </div>
          <Progress value={percentage} className="h-3" />
        </div>

        {/* 현재 테이블 */}
        {currentTable && (
          <div className="flex items-center gap-2 text-sm">
            <Database className="h-4 w-4 text-muted-foreground" />
            <span className="text-muted-foreground">현재 테이블:</span>
            <span className="font-mono font-medium">{currentTable}</span>
          </div>
        )}

        {/* 통계 그리드 */}
        <div className="grid grid-cols-2 gap-3">
          <StatCard
            icon={Rows3}
            label="처리된 행"
            value={`${formatNumber(rowsProcessed)} / ${formatNumber(totalRows)}`}
          />
          <StatCard
            icon={HardDrive}
            label="처리된 데이터"
            value={formatBytes(bytesProcessed)}
          />
          {estimatedTimeRemaining !== undefined && estimatedTimeRemaining > 0 && (
            <StatCard
              icon={Clock}
              label="예상 남은 시간"
              value={formatDuration(estimatedTimeRemaining * 1000)}
              className="col-span-2"
            />
          )}
        </div>

        {/* 행 처리 진행률 */}
        {totalRows > 0 && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">행 처리 진행률</span>
              <span className="font-semibold">
                {((rowsProcessed / totalRows) * 100).toFixed(1)}%
              </span>
            </div>
            <Progress
              value={(rowsProcessed / totalRows) * 100}
              className="h-2"
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}
