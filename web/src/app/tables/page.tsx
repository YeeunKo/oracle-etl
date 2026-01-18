/**
 * Oracle 테이블 목록 페이지
 * Oracle 데이터베이스의 테이블 정보 조회 및 검색
 */

'use client';

import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { MainLayout } from '@/components/layout/main-layout';
import { getTables } from '@/lib/api';
import type { OracleTable } from '@/types/api';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Database,
  Search,
  RefreshCw,
  AlertCircle,
  TableIcon,
  Loader2,
} from 'lucide-react';

// ============================================================================
// 유틸리티 함수
// ============================================================================

/**
 * 숫자를 읽기 쉬운 형식으로 포맷팅
 * @param num - 숫자
 * @returns 포맷팅된 문자열 (예: 1,234,567)
 */
function formatNumber(num: number | undefined): string {
  if (num === undefined || num === null) {
    return '-';
  }
  return num.toLocaleString('ko-KR');
}

/**
 * 날짜를 읽기 쉬운 형식으로 포맷팅
 * @param dateStr - ISO 날짜 문자열
 * @returns 포맷팅된 날짜 문자열
 */
function formatDate(dateStr: string | undefined): string {
  if (!dateStr) {
    return '-';
  }
  try {
    return new Date(dateStr).toLocaleString('ko-KR', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateStr;
  }
}

// ============================================================================
// 테이블 스켈레톤 컴포넌트
// ============================================================================

/**
 * 테이블 로딩 스켈레톤
 */
function TableSkeleton() {
  return (
    <div className="space-y-3">
      {/* 테이블 헤더 스켈레톤 */}
      <div className="flex gap-4 border-b pb-3">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-4 w-28" />
      </div>
      {/* 테이블 행 스켈레톤 */}
      {Array.from({ length: 10 }).map((_, index) => (
        <div key={index} className="flex gap-4 py-2">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-4 w-28" />
        </div>
      ))}
    </div>
  );
}

// ============================================================================
// 테이블 목록 페이지 컴포넌트
// ============================================================================

/**
 * Oracle 테이블 목록 페이지
 * - 테이블 목록 표시 (Owner, Table Name, Rows, Last Analyzed)
 * - 테이블 이름으로 검색/필터링
 * - 로딩 및 에러 상태 처리
 */
export default function TablesPage() {
  // 검색어 상태
  const [searchQuery, setSearchQuery] = useState('');

  // 테이블 목록 조회 쿼리
  const {
    data: tables,
    isLoading,
    isError,
    error,
    refetch,
    isFetching,
  } = useQuery<OracleTable[], Error>({
    queryKey: ['tables'],
    queryFn: getTables,
    staleTime: 1000 * 60 * 5, // 5분간 캐시 유지
  });

  // 검색 필터링된 테이블 목록
  const filteredTables = useMemo(() => {
    if (!tables) return [];

    const query = searchQuery.toLowerCase().trim();
    if (!query) return tables;

    return tables.filter(
      (table) =>
        table.tableName.toLowerCase().includes(query) ||
        table.owner.toLowerCase().includes(query)
    );
  }, [tables, searchQuery]);

  // 테이블 통계
  const totalTables = tables?.length ?? 0;
  const filteredCount = filteredTables.length;
  const totalRows = tables?.reduce((sum, t) => sum + (t.numRows ?? 0), 0) ?? 0;

  return (
    <MainLayout>
      <div className="space-y-6">
        {/* 페이지 헤더 */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">Oracle 테이블</h1>
            <p className="text-muted-foreground">
              Oracle 데이터베이스의 테이블 목록을 조회합니다.
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => refetch()}
            disabled={isFetching}
          >
            {isFetching ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4" />
            )}
            새로고침
          </Button>
        </div>

        {/* 통계 카드 */}
        {!isLoading && !isError && tables && (
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center gap-4">
                  <div className="rounded-full bg-primary/10 p-3">
                    <Database className="h-6 w-6 text-primary" />
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">전체 테이블</p>
                    <p className="text-2xl font-bold">{formatNumber(totalTables)}</p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center gap-4">
                  <div className="rounded-full bg-blue-500/10 p-3">
                    <TableIcon className="h-6 w-6 text-blue-500" />
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">필터링된 테이블</p>
                    <p className="text-2xl font-bold">{formatNumber(filteredCount)}</p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center gap-4">
                  <div className="rounded-full bg-green-500/10 p-3">
                    <Database className="h-6 w-6 text-green-500" />
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">전체 행 수</p>
                    <p className="text-2xl font-bold">{formatNumber(totalRows)}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        )}

        {/* 테이블 목록 카드 */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>테이블 목록</CardTitle>
                <CardDescription>
                  Oracle 데이터베이스에서 조회 가능한 테이블입니다.
                </CardDescription>
              </div>
              {/* 검색 입력 */}
              <div className="relative w-72">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="테이블 이름으로 검색..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9"
                />
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {/* 로딩 상태 */}
            {isLoading && <TableSkeleton />}

            {/* 에러 상태 */}
            {isError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>테이블 조회 실패</AlertTitle>
                <AlertDescription>
                  {error?.message || '테이블 목록을 불러오는 중 오류가 발생했습니다.'}
                </AlertDescription>
              </Alert>
            )}

            {/* 테이블 목록 */}
            {!isLoading && !isError && tables && (
              <>
                {filteredTables.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-12 text-center">
                    <TableIcon className="h-12 w-12 text-muted-foreground/50" />
                    <h3 className="mt-4 text-lg font-medium">
                      {searchQuery
                        ? '검색 결과가 없습니다'
                        : '테이블이 없습니다'}
                    </h3>
                    <p className="mt-2 text-sm text-muted-foreground">
                      {searchQuery
                        ? '다른 검색어를 시도해보세요.'
                        : 'Oracle 데이터베이스에 조회 가능한 테이블이 없습니다.'}
                    </p>
                    {searchQuery && (
                      <Button
                        variant="outline"
                        className="mt-4"
                        onClick={() => setSearchQuery('')}
                      >
                        검색 초기화
                      </Button>
                    )}
                  </div>
                ) : (
                  <>
                    {/* 검색 결과 표시 */}
                    {searchQuery && (
                      <div className="mb-4 flex items-center gap-2">
                        <Badge variant="secondary">
                          &quot;{searchQuery}&quot; 검색 결과: {filteredCount}개
                        </Badge>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setSearchQuery('')}
                        >
                          초기화
                        </Button>
                      </div>
                    )}

                    {/* 테이블 */}
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-[150px]">Owner</TableHead>
                          <TableHead>테이블명</TableHead>
                          <TableHead className="w-[120px] text-right">
                            행 수
                          </TableHead>
                          <TableHead className="w-[180px]">
                            마지막 분석일
                          </TableHead>
                          <TableHead className="w-[150px]">
                            테이블스페이스
                          </TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {filteredTables.map((table) => (
                          <TableRow key={`${table.owner}.${table.tableName}`}>
                            <TableCell>
                              <Badge variant="outline">{table.owner}</Badge>
                            </TableCell>
                            <TableCell className="font-medium">
                              {table.tableName}
                            </TableCell>
                            <TableCell className="text-right font-mono">
                              {formatNumber(table.numRows)}
                            </TableCell>
                            <TableCell className="text-muted-foreground">
                              {formatDate(table.lastAnalyzed)}
                            </TableCell>
                            <TableCell>
                              {table.tablespaceName ? (
                                <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
                                  {table.tablespaceName}
                                </code>
                              ) : (
                                <span className="text-muted-foreground">-</span>
                              )}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </>
                )}
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </MainLayout>
  );
}
