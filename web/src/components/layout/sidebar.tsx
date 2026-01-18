/**
 * 사이드바 네비게이션 컴포넌트
 * 대시보드, Transports, Jobs 링크 제공
 */

'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  Truck,
  ClipboardList,
  Database,
  Settings,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useState } from 'react';

// ============================================================================
// 네비게이션 메뉴 정의
// ============================================================================

/**
 * 네비게이션 메뉴 아이템 타입
 */
interface NavItem {
  /** 메뉴 이름 */
  name: string;
  /** 링크 경로 */
  href: string;
  /** 아이콘 컴포넌트 */
  icon: React.ComponentType<{ className?: string }>;
  /** 설명 */
  description?: string;
}

/**
 * 메인 네비게이션 메뉴
 */
const mainNavItems: NavItem[] = [
  {
    name: '대시보드',
    href: '/',
    icon: LayoutDashboard,
    description: '전체 현황 및 통계',
  },
  {
    name: 'Transports',
    href: '/transports',
    icon: Truck,
    description: 'ETL Transport 관리',
  },
  {
    name: 'Jobs',
    href: '/jobs',
    icon: ClipboardList,
    description: '실행 작업 모니터링',
  },
  {
    name: 'Tables',
    href: '/tables',
    icon: Database,
    description: 'Oracle 테이블 조회',
  },
];

/**
 * 하단 네비게이션 메뉴
 */
const bottomNavItems: NavItem[] = [
  {
    name: '설정',
    href: '/settings',
    icon: Settings,
    description: 'API 키 및 환경 설정',
  },
];

// ============================================================================
// 사이드바 컴포넌트
// ============================================================================

/**
 * 사이드바 Props
 */
interface SidebarProps {
  /** 추가 클래스명 */
  className?: string;
}

/**
 * 사이드바 네비게이션 컴포넌트
 *
 * @param props - 컴포넌트 Props
 * @returns 사이드바 컴포넌트
 */
export function Sidebar({ className }: SidebarProps) {
  const pathname = usePathname();
  const [isCollapsed, setIsCollapsed] = useState(false);

  /**
   * 현재 경로가 메뉴 항목과 일치하는지 확인
   */
  const isActive = (href: string) => {
    if (href === '/') {
      return pathname === '/';
    }
    return pathname.startsWith(href);
  };

  return (
    <aside
      className={cn(
        'flex flex-col border-r border-border bg-sidebar transition-all duration-300',
        isCollapsed ? 'w-16' : 'w-64',
        className
      )}
    >
      {/* 로고 영역 */}
      <div className="flex h-16 items-center justify-between border-b border-border px-4">
        {!isCollapsed && (
          <Link href="/" className="flex items-center gap-2">
            <Database className="h-6 w-6 text-primary" />
            <span className="font-semibold text-lg text-sidebar-foreground">
              Oracle ETL
            </span>
          </Link>
        )}
        {isCollapsed && (
          <Link href="/" className="mx-auto">
            <Database className="h-6 w-6 text-primary" />
          </Link>
        )}
      </div>

      {/* 메인 네비게이션 */}
      <nav className="flex-1 space-y-1 p-2">
        {mainNavItems.map((item) => {
          const Icon = item.icon;
          const active = isActive(item.href);

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                active
                  ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                  : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                isCollapsed && 'justify-center px-2'
              )}
              title={isCollapsed ? item.name : undefined}
            >
              <Icon className="h-5 w-5 shrink-0" />
              {!isCollapsed && <span>{item.name}</span>}
            </Link>
          );
        })}
      </nav>

      {/* 하단 네비게이션 */}
      <div className="border-t border-border p-2">
        {bottomNavItems.map((item) => {
          const Icon = item.icon;
          const active = isActive(item.href);

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                active
                  ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                  : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                isCollapsed && 'justify-center px-2'
              )}
              title={isCollapsed ? item.name : undefined}
            >
              <Icon className="h-5 w-5 shrink-0" />
              {!isCollapsed && <span>{item.name}</span>}
            </Link>
          );
        })}

        {/* 사이드바 접기/펼치기 버튼 */}
        <button
          onClick={() => setIsCollapsed(!isCollapsed)}
          className={cn(
            'mt-2 flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium',
            'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground transition-colors',
            isCollapsed && 'justify-center px-2'
          )}
          aria-label={isCollapsed ? '사이드바 펼치기' : '사이드바 접기'}
        >
          {isCollapsed ? (
            <ChevronRight className="h-5 w-5" />
          ) : (
            <>
              <ChevronLeft className="h-5 w-5" />
              <span>접기</span>
            </>
          )}
        </button>
      </div>
    </aside>
  );
}
