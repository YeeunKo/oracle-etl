/**
 * 인증 상태 관리 스토어
 * Zustand를 사용한 API 키 및 인증 상태 관리
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// ============================================================================
// 타입 정의
// ============================================================================

/**
 * 인증 스토어 상태 인터페이스
 */
interface AuthState {
  /** API 키 */
  apiKey: string | null;

  /** API 키 설정 여부 */
  isAuthenticated: boolean;
}

/**
 * 인증 스토어 액션 인터페이스
 */
interface AuthActions {
  /** API 키 설정 */
  setApiKey: (apiKey: string) => void;

  /** API 키 제거 (로그아웃) */
  clearApiKey: () => void;
}

/**
 * 인증 스토어 전체 타입
 */
type AuthStore = AuthState & AuthActions;

// ============================================================================
// 스토어 생성
// ============================================================================

/**
 * 인증 상태 관리 스토어
 *
 * @example
 * // 컴포넌트에서 사용
 * const { apiKey, setApiKey, clearApiKey } = useAuthStore();
 *
 * // API 키 설정
 * setApiKey('your-api-key');
 *
 * // API 키 조회 (스토어 외부에서)
 * const key = useAuthStore.getState().apiKey;
 */
export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      // 초기 상태
      apiKey: null,
      isAuthenticated: false,

      // 액션: API 키 설정
      setApiKey: (apiKey: string) => {
        set({
          apiKey,
          isAuthenticated: true,
        });
      },

      // 액션: API 키 제거
      clearApiKey: () => {
        set({
          apiKey: null,
          isAuthenticated: false,
        });
      },
    }),
    {
      // localStorage 키 이름
      name: 'oracle-etl-auth',

      // 저장할 필드 선택 (isAuthenticated는 apiKey 유무로 계산됨)
      partialize: (state) => ({
        apiKey: state.apiKey,
      }),

      // 복원 시 isAuthenticated 상태 복구
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.isAuthenticated = !!state.apiKey;
        }
      },
    }
  )
);

// ============================================================================
// 셀렉터 함수
// ============================================================================

/**
 * API 키 셀렉터
 */
export const selectApiKey = (state: AuthStore) => state.apiKey;

/**
 * 인증 여부 셀렉터
 */
export const selectIsAuthenticated = (state: AuthStore) => state.isAuthenticated;
