/**
 * Transport 목록 페이지
 * FR-02.1: Transport 목록 표시
 * FR-02.2: 페이지네이션
 * FR-02.3: 이름 검색
 */

'use client';

import { MainLayout } from '@/components/layout/main-layout';
import { TransportList } from '@/components/transport/transport-list';

// ============================================================================
// 메인 페이지 컴포넌트
// ============================================================================

/**
 * Transport 목록 페이지
 * - Transport 테이블 표시
 * - 검색 및 페이지네이션
 * - 생성/실행/삭제 액션
 */
export default function TransportsPage() {
  return (
    <MainLayout>
      <div className="space-y-6">
        {/* 페이지 헤더 */}
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Transport 관리</h1>
          <p className="text-muted-foreground">
            Oracle 테이블에서 데이터를 추출하는 Transport를 관리합니다.
          </p>
        </div>

        {/* Transport 목록 */}
        <TransportList />
      </div>
    </MainLayout>
  );
}
