/**
 * Transport 생성 페이지
 * FR-02.4: 유효성 검증을 포함한 생성 폼
 * FR-02.5: Oracle 테이블 선택 드롭다운
 */

'use client';

import Link from 'next/link';
import { ChevronLeft } from 'lucide-react';
import { MainLayout } from '@/components/layout/main-layout';
import { TransportForm } from '@/components/transport/transport-form';
import { Button } from '@/components/ui/button';

// ============================================================================
// 메인 페이지 컴포넌트
// ============================================================================

/**
 * Transport 생성 페이지
 * - Transport 생성 폼
 * - 유효성 검증
 * - Oracle 테이블 선택
 */
export default function NewTransportPage() {
  return (
    <MainLayout>
      <div className="space-y-6">
        {/* 페이지 헤더 */}
        <div className="flex items-center gap-4">
          <Link href="/transports">
            <Button variant="ghost" size="sm">
              <ChevronLeft className="mr-1 h-4 w-4" />
              목록으로
            </Button>
          </Link>
        </div>

        <div>
          <h1 className="text-2xl font-bold tracking-tight">새 Transport 생성</h1>
          <p className="text-muted-foreground">
            Oracle 테이블에서 데이터를 추출할 Transport를 설정합니다.
          </p>
        </div>

        {/* Transport 생성 폼 */}
        <div className="max-w-2xl">
          <TransportForm />
        </div>
      </div>
    </MainLayout>
  );
}
