/**
 * 메인 레이아웃 컴포넌트
 * 사이드바와 헤더를 포함한 전체 페이지 레이아웃
 */

'use client';

import { Sidebar } from './sidebar';
import { Header } from './header';
import { cn } from '@/lib/utils';

// ============================================================================
// 메인 레이아웃 컴포넌트
// ============================================================================

/**
 * 메인 레이아웃 Props
 */
interface MainLayoutProps {
  /** 페이지 컨텐츠 */
  children: React.ReactNode;
  /** 추가 클래스명 */
  className?: string;
}

/**
 * 메인 레이아웃 컴포넌트
 * 사이드바 네비게이션과 헤더를 포함한 레이아웃 제공
 *
 * @param props - 컴포넌트 Props
 * @returns 메인 레이아웃 컴포넌트
 *
 * @example
 * <MainLayout>
 *   <DashboardPage />
 * </MainLayout>
 */
export function MainLayout({ children, className }: MainLayoutProps) {
  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* 사이드바 */}
      <Sidebar className="hidden md:flex" />

      {/* 메인 컨텐츠 영역 */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* 헤더 */}
        <Header />

        {/* 페이지 컨텐츠 */}
        <main
          className={cn(
            'flex-1 overflow-auto p-6',
            className
          )}
        >
          {children}
        </main>
      </div>
    </div>
  );
}
