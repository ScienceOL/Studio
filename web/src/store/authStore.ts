import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { AuthUtils, type UserInfo } from '@/lib/auth';

export interface AuthState {
  // 状态
  isAuthenticated: boolean;
  isLoading: boolean;
  user: UserInfo | null;
  
  // 动作
  login: () => void;
  logout: () => void;
  refreshToken: () => Promise<boolean>;
  setUser: (user: UserInfo | null) => void;
  setLoading: (loading: boolean) => void;
  setAuthenticated: (authenticated: boolean) => void;
  checkAuthStatus: () => Promise<void>;
  initialize: () => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set, get) => ({
        // 初始状态
        isAuthenticated: false,
        isLoading: true,
        user: null,

        // 设置用户信息
        setUser: (user) => set({ user }),

        // 设置加载状态
        setLoading: (isLoading) => set({ isLoading }),

        // 设置认证状态
        setAuthenticated: (isAuthenticated) => set({ isAuthenticated }),

        // 登录（重定向到后端）
        login: () => {
          AuthUtils.redirectToLogin();
        },

        // 登出
        logout: () => {
          AuthUtils.logout();
          set({
            isAuthenticated: false,
            user: null,
          });
        },

        // 刷新 token
        refreshToken: async () => {
          const success = await AuthUtils.refreshToken();
          if (success) {
            const userInfo = AuthUtils.getUserInfo();
            set({
              isAuthenticated: true,
              user: userInfo,
            });
          } else {
            set({
              isAuthenticated: false,
              user: null,
            });
          }
          return success;
        },

        // 检查认证状态
        checkAuthStatus: async () => {
          set({ isLoading: true });

          try {
            const authenticated = AuthUtils.isAuthenticated();
            
            if (authenticated) {
              const userInfo = AuthUtils.getUserInfo();
              set({
                isAuthenticated: true,
                user: userInfo,
              });
            } else {
              // 尝试刷新 token
              const refreshed = await get().refreshToken();
              if (!refreshed) {
                set({
                  isAuthenticated: false,
                  user: null,
                });
              }
            }
          } catch (error) {
            console.error('Auth check failed:', error);
            set({
              isAuthenticated: false,
              user: null,
            });
          } finally {
            set({ isLoading: false });
          }
        },

        // 初始化认证状态
        initialize: () => {
          get().checkAuthStatus();
          
          // 监听 storage 变化（支持多 tab 同步）
          const handleStorageChange = (e: StorageEvent) => {
            if (
              e.key?.startsWith('studio_') ||
              e.key === null // localStorage.clear() was called
            ) {
              get().checkAuthStatus();
            }
          };

          window.addEventListener('storage', handleStorageChange);
        },
      }),
      {
        name: 'auth-storage',
        // 只持久化用户数据，不持久化加载状态
        partialize: (state) => ({
          user: state.user,
          isAuthenticated: state.isAuthenticated,
        }),
      }
    ),
    {
      name: 'auth-store',
    }
  )
);