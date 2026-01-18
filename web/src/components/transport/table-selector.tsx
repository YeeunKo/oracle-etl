/**
 * Oracle 테이블 선택 드롭다운 컴포넌트
 * FR-02.5: Create form에서 Oracle 테이블 선택
 */

'use client';

import { useQuery } from '@tanstack/react-query';
import { getTables } from '@/lib/api';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import type { OracleTable } from '@/types/api';

// ============================================================================
// Query Keys
// ============================================================================

export const tableQueryKeys = {
  all: ['tables'] as const,
  list: () => ['tables', 'list'] as const,
};

// ============================================================================
// Props 타입
// ============================================================================

interface TableSelectorProps {
  /** 선택된 테이블 (SCHEMA.TABLE 형식) */
  value?: string;
  /** 테이블 선택 시 콜백 */
  onValueChange: (value: string, schema: string, table: string) => void;
  /** 비활성화 여부 */
  disabled?: boolean;
  /** placeholder 텍스트 */
  placeholder?: string;
}

// ============================================================================
// 유틸리티 함수
// ============================================================================

/**
 * 테이블을 스키마별로 그룹화
 */
function groupTablesBySchema(tables: OracleTable[]): Map<string, OracleTable[]> {
  const grouped = new Map<string, OracleTable[]>();

  tables.forEach((table) => {
    const existing = grouped.get(table.owner) ?? [];
    grouped.set(table.owner, [...existing, table]);
  });

  return grouped;
}

/**
 * 테이블 ID 생성 (SCHEMA.TABLE)
 */
function getTableId(table: OracleTable): string {
  return `${table.owner}.${table.tableName}`;
}

/**
 * 테이블 ID 파싱
 */
function parseTableId(id: string): { schema: string; table: string } | null {
  const parts = id.split('.');
  if (parts.length !== 2) return null;
  return { schema: parts[0], table: parts[1] };
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * Oracle 테이블 선택 드롭다운
 * 스키마별로 그룹화되어 표시됨
 *
 * @example
 * <TableSelector
 *   value={selectedTable}
 *   onValueChange={(value, schema, table) => {
 *     setSchema(schema);
 *     setTable(table);
 *   }}
 * />
 */
export function TableSelector({
  value,
  onValueChange,
  disabled,
  placeholder = '테이블 선택...',
}: TableSelectorProps) {
  // 테이블 목록 조회
  const {
    data: tables,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: tableQueryKeys.list(),
    queryFn: getTables,
    staleTime: 5 * 60 * 1000, // 5분
  });

  // 값 변경 핸들러
  const handleValueChange = (newValue: string) => {
    const parsed = parseTableId(newValue);
    if (parsed) {
      onValueChange(newValue, parsed.schema, parsed.table);
    }
  };

  // 로딩 상태
  if (isLoading) {
    return <Skeleton className="h-9 w-full" />;
  }

  // 에러 상태
  if (isError) {
    return (
      <div className="text-sm text-destructive">
        테이블 목록을 불러올 수 없습니다.
        {error instanceof Error ? `: ${error.message}` : ''}
      </div>
    );
  }

  // 빈 테이블 목록
  if (!tables || tables.length === 0) {
    return (
      <div className="text-sm text-muted-foreground">
        사용 가능한 테이블이 없습니다.
      </div>
    );
  }

  // 스키마별 그룹화
  const groupedTables = groupTablesBySchema(tables);

  return (
    <Select value={value} onValueChange={handleValueChange} disabled={disabled}>
      <SelectTrigger className="w-full">
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent className="max-h-80">
        {Array.from(groupedTables.entries()).map(([schema, schemaTables]) => (
          <SelectGroup key={schema}>
            <SelectLabel className="font-semibold">{schema}</SelectLabel>
            {schemaTables.map((table) => (
              <SelectItem
                key={getTableId(table)}
                value={getTableId(table)}
              >
                <div className="flex items-center justify-between w-full gap-4">
                  <span>{table.tableName}</span>
                  {table.numRows !== undefined && (
                    <span className="text-xs text-muted-foreground">
                      {table.numRows.toLocaleString()} rows
                    </span>
                  )}
                </div>
              </SelectItem>
            ))}
          </SelectGroup>
        ))}
      </SelectContent>
    </Select>
  );
}
