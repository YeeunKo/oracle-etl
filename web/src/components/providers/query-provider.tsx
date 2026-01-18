/**
 * TanStack Query Provider
 * React Query 클라이언트 설정 및 Provider 제공
 */

'use client';

import { useState } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
// import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

// ============================================================================
// Query Client 설정
// ============================================================================

/**
 * Query Client 생성 함수
 * 각 사용자 세션마다 새로운 클라이언트 생성
 */
function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // 데이터가 stale 상태로 전환되기까지의 시간 (5분)
        staleTime: 5 * 60 * 1000,

        // 캐시 유지 시간 (10분)
        gcTime: 10 * 60 * 1000,

        // 기본 재시도 횟수
        retry: 1,

        // 재시도 지연 (지수 백오프)
        retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),

        // 윈도우 포커스 시 재조회 비활성화 (필요한 곳에서만 활성화)
        refetchOnWindowFocus: false,

        // 네트워크 재연결 시 재조회
        refetchOnReconnect: true,
      },
      mutations: {
        // 뮤테이션 재시도 없음
        retry: 0,
      },
    },
  });
}

// 브라우저 환경에서 싱글톤 클라이언트 (SSR 고려)
let browserQueryClient: QueryClient | undefined = undefined;

/**
 * Query Client 가져오기
 * 서버: 매번 새 클라이언트 생성
 * 브라우저: 싱글톤 클라이언트 재사용
 */
function getQueryClient() {
  if (typeof window === 'undefined') {
    // 서버 환경: 항상 새 클라이언트 생성
    return makeQueryClient();
  } else {
    // 브라우저 환경: 싱글톤 패턴
    if (!browserQueryClient) {
      browserQueryClient = makeQueryClient();
    }
    return browserQueryClient;
  }
}

// ============================================================================
// Provider 컴포넌트
// ============================================================================

/**
 * QueryProvider Props
 */
interface QueryProviderProps {
  children: React.ReactNode;
}

/**
 * TanStack Query Provider 컴포넌트
 *
 * @param props - children을 포함한 Props
 * @returns QueryClientProvider로 래핑된 children
 *
 * @example
 * // app/layout.tsx
 * export default function RootLayout({ children }) {
 *   return (
 *     <html>
 *       <body>
 *         <QueryProvider>
 *           {children}
 *         </QueryProvider>
 *       </body>
 *     </html>
 *   );
 * }
 */
export function QueryProvider({ children }: QueryProviderProps) {
  // useState를 사용하여 서버/클라이언트 간 일관성 유지
  const [queryClient] = useState(() => getQueryClient());

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {/* 개발 환경에서만 DevTools 표시 */}
      {/* {process.env.NODE_ENV === 'development' && (
        <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-right" />
      )} */}
    </QueryClientProvider>
  );
}
