/**
 * Transport 생성/수정 폼 컴포넌트
 * FR-02.4: 유효성 검증을 포함한 생성 폼
 * FR-02.5: Oracle 테이블 선택 드롭다운
 */

'use client';

import { useForm } from 'react-hook-form';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { createTransport } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Label } from '@/components/ui/label';
import { TableSelector } from './table-selector';
import { transportQueryKeys } from './transport-list';
import { AlertCircle } from 'lucide-react';

// ============================================================================
// 타입 정의
// ============================================================================

/**
 * Transport 폼 데이터 타입
 */
interface TransportFormData {
  name: string;
  description: string;
  sourceTable: string;
  sourceSchema: string;
  targetPath: string;
  targetFormat: 'csv' | 'json' | 'parquet';
  batchSize: number;
  parallelism: number;
}

// ============================================================================
// 기본값
// ============================================================================

const DEFAULT_VALUES: TransportFormData = {
  name: '',
  description: '',
  sourceTable: '',
  sourceSchema: '',
  targetPath: '/data/exports/',
  targetFormat: 'csv',
  batchSize: 10000,
  parallelism: 4,
};

// ============================================================================
// 유효성 검증
// ============================================================================

/**
 * 폼 유효성 검증
 */
function validateForm(data: TransportFormData): Record<string, string> {
  const errors: Record<string, string> = {};

  if (!data.name.trim()) {
    errors.name = '이름을 입력해주세요.';
  } else if (data.name.length > 100) {
    errors.name = '이름은 100자 이하여야 합니다.';
  } else if (!/^[a-zA-Z0-9가-힣\s_-]+$/.test(data.name)) {
    errors.name = '이름에는 특수문자를 사용할 수 없습니다.';
  }

  if (data.description && data.description.length > 500) {
    errors.description = '설명은 500자 이하여야 합니다.';
  }

  if (!data.sourceTable) {
    errors.sourceTable = '소스 테이블을 선택해주세요.';
  }

  if (!data.sourceSchema) {
    errors.sourceSchema = '소스 스키마를 선택해주세요.';
  }

  if (!data.targetPath.trim()) {
    errors.targetPath = '대상 경로를 입력해주세요.';
  } else if (!data.targetPath.startsWith('/')) {
    errors.targetPath = '경로는 /로 시작해야 합니다.';
  }

  if (data.batchSize < 100 || data.batchSize > 100000) {
    errors.batchSize = '배치 크기는 100에서 100,000 사이여야 합니다.';
  }

  if (data.parallelism < 1 || data.parallelism > 16) {
    errors.parallelism = '병렬 처리 수는 1에서 16 사이여야 합니다.';
  }

  return errors;
}

// ============================================================================
// Props 타입
// ============================================================================

interface TransportFormProps {
  /** 수정 모드 시 기존 데이터 */
  initialData?: Partial<TransportFormData>;
  /** 폼 제목 */
  title?: string;
}

// ============================================================================
// 메인 컴포넌트
// ============================================================================

/**
 * Transport 생성/수정 폼
 *
 * @example
 * // 생성 모드
 * <TransportForm />
 *
 * // 수정 모드
 * <TransportForm initialData={existingTransport} title="Transport 수정" />
 */
export function TransportForm({
  initialData,
  title = '새 Transport 생성',
}: TransportFormProps) {
  const router = useRouter();
  const queryClient = useQueryClient();

  // 폼 초기화
  const {
    register,
    handleSubmit,
    watch,
    setValue,
    setError,
    formState: { errors },
  } = useForm<TransportFormData>({
    defaultValues: {
      ...DEFAULT_VALUES,
      ...initialData,
    },
  });

  // Transport 생성 mutation
  const createMutation = useMutation({
    mutationFn: createTransport,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: transportQueryKeys.all });
      router.push(`/transports/${data.id}`);
    },
  });

  // 테이블 선택 핸들러
  const handleTableSelect = (
    _value: string,
    schema: string,
    table: string
  ) => {
    setValue('sourceSchema', schema);
    setValue('sourceTable', table);
  };

  // 폼 제출 핸들러
  const onSubmit = (data: TransportFormData) => {
    // 유효성 검증
    const validationErrors = validateForm(data);
    if (Object.keys(validationErrors).length > 0) {
      Object.entries(validationErrors).forEach(([field, message]) => {
        setError(field as keyof TransportFormData, { message });
      });
      return;
    }

    createMutation.mutate({
      name: data.name,
      description: data.description || undefined,
      sourceTable: data.sourceTable,
      sourceSchema: data.sourceSchema,
      targetPath: data.targetPath,
      targetFormat: data.targetFormat,
      config: {
        batchSize: data.batchSize,
        parallelism: data.parallelism,
        retryAttempts: 3,
        retryDelay: 1000,
        timeout: 300000,
      },
    });
  };

  // 현재 선택된 테이블 값
  const watchedSchema = watch('sourceSchema');
  const watchedTable = watch('sourceTable');
  const selectedTableValue =
    watchedSchema && watchedTable
      ? `${watchedSchema}.${watchedTable}`
      : undefined;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* 에러 표시 */}
      {createMutation.isError && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {createMutation.error instanceof Error
              ? createMutation.error.message
              : 'Transport 생성에 실패했습니다.'}
          </AlertDescription>
        </Alert>
      )}

      {/* 기본 정보 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{title}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 이름 */}
          <div className="space-y-2">
            <Label htmlFor="name">이름 *</Label>
            <Input
              id="name"
              placeholder="예: Daily Sales Export"
              {...register('name')}
            />
            <p className="text-sm text-muted-foreground">
              Transport를 식별할 수 있는 이름을 입력하세요.
            </p>
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name.message}</p>
            )}
          </div>

          {/* 설명 */}
          <div className="space-y-2">
            <Label htmlFor="description">설명</Label>
            <Input
              id="description"
              placeholder="예: 매일 판매 데이터를 CSV로 추출"
              {...register('description')}
            />
            {errors.description && (
              <p className="text-sm text-destructive">{errors.description.message}</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 소스 설정 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">소스 설정</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 테이블 선택 (FR-02.5) */}
          <div className="space-y-2">
            <Label>Oracle 테이블 *</Label>
            <TableSelector
              value={selectedTableValue}
              onValueChange={handleTableSelect}
            />
            <p className="text-sm text-muted-foreground">
              데이터를 추출할 Oracle 테이블을 선택하세요.
            </p>
            {errors.sourceTable && (
              <p className="text-sm text-destructive">{errors.sourceTable.message}</p>
            )}
          </div>

          {/* 선택된 테이블 정보 표시 */}
          {watchedSchema && watchedTable && (
            <div className="rounded-md bg-muted p-3 text-sm">
              <span className="text-muted-foreground">선택된 테이블: </span>
              <span className="font-mono font-medium">
                {watchedSchema}.{watchedTable}
              </span>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 대상 설정 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">대상 설정</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 대상 경로 */}
          <div className="space-y-2">
            <Label htmlFor="targetPath">대상 경로 *</Label>
            <Input
              id="targetPath"
              placeholder="/data/exports/sales/"
              {...register('targetPath')}
            />
            <p className="text-sm text-muted-foreground">
              추출된 데이터가 저장될 경로를 입력하세요.
            </p>
            {errors.targetPath && (
              <p className="text-sm text-destructive">{errors.targetPath.message}</p>
            )}
          </div>

          {/* 출력 형식 */}
          <div className="space-y-2">
            <Label>출력 형식 *</Label>
            <Select
              value={watch('targetFormat')}
              onValueChange={(value: 'csv' | 'json' | 'parquet') =>
                setValue('targetFormat', value)
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="출력 형식 선택" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="csv">CSV</SelectItem>
                <SelectItem value="json">JSON</SelectItem>
                <SelectItem value="parquet">Parquet</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-sm text-muted-foreground">
              데이터 출력 형식을 선택하세요.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* 고급 설정 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">고급 설정</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            {/* 배치 크기 */}
            <div className="space-y-2">
              <Label htmlFor="batchSize">배치 크기</Label>
              <Input
                id="batchSize"
                type="number"
                min={100}
                max={100000}
                {...register('batchSize', { valueAsNumber: true })}
              />
              <p className="text-sm text-muted-foreground">
                한 번에 처리할 행 수
              </p>
              {errors.batchSize && (
                <p className="text-sm text-destructive">{errors.batchSize.message}</p>
              )}
            </div>

            {/* 병렬 처리 수 */}
            <div className="space-y-2">
              <Label htmlFor="parallelism">병렬 처리 수</Label>
              <Input
                id="parallelism"
                type="number"
                min={1}
                max={16}
                {...register('parallelism', { valueAsNumber: true })}
              />
              <p className="text-sm text-muted-foreground">
                동시 처리 워커 수
              </p>
              {errors.parallelism && (
                <p className="text-sm text-destructive">{errors.parallelism.message}</p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 버튼 */}
      <div className="flex items-center justify-end gap-3">
        <Button
          type="button"
          variant="outline"
          onClick={() => router.back()}
          disabled={createMutation.isPending}
        >
          취소
        </Button>
        <Button type="submit" disabled={createMutation.isPending}>
          {createMutation.isPending ? '생성 중...' : 'Transport 생성'}
        </Button>
      </div>
    </form>
  );
}
