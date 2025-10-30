import type { UserInfo } from '@/utils/auth';
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

export interface AuthState {
  // ===== 状态 =====
  isAuthenticated: boolean;
  isLoading: boolean;
  user: UserInfo | null;

  // ===== 动作（仅状态更新，无业务逻辑）=====
  setUser: (user: UserInfo | null) => void;
  setLoading: (loading: boolean) => void;
  setAuthenticated: (authenticated: boolean) => void;

  // ===== 重置 =====
  reset: () => void;
}

const initialState = {
  isAuthenticated: false,
  isLoading: true,
  user: null,
};

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set) => ({
        // 初始状态
        ...initialState,

        // 设置用户信息
        setUser: (user) => set({ user }),

        // 设置加载状态
        setLoading: (isLoading) => set({ isLoading }),

        // 设置认证状态
        setAuthenticated: (isAuthenticated) => set({ isAuthenticated }),

        // 重置到初始状态
        reset: () => set(initialState),
      }),
      {
        name: 'auth-storage',
        // 只持久化用户数据和认证状态
        partialize: (state) => ({
          user: state.user,
          isAuthenticated: state.isAuthenticated,
        }),
        // 恢复状态后，确保 isLoading 为 false
        onRehydrateStorage: () => (state) => {
          if (state) {
            console.log('💾 [AuthStore] State rehydrated from storage');
            state.isLoading = false;
          }
        },
      }
    ),
    {
      name: 'auth-store',
    }
  )
);
