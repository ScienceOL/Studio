import { AuthUtils, type UserInfo } from '@/lib/auth';
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

// 模块级变量：跟踪初始化状态（更轻量、更合理）
let isInitialized = false;
let storageListener: ((e: StorageEvent) => void) | null = null;

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
          console.log('🔄 Refreshing token...');
          const success = await AuthUtils.refreshToken();
          console.log('🔄 Refresh token result:', success);

          if (success) {
            const userInfo = AuthUtils.getUserInfo();
            set({
              isAuthenticated: true,
              user: userInfo,
              isLoading: false,
            });
          } else {
            set({
              isAuthenticated: false,
              user: null,
              isLoading: false,
            });
          }
          return success;
        },

        // 检查认证状态
        checkAuthStatus: async () => {
          console.log('🔍 Checking auth status...');
          set({ isLoading: true });

          try {
            const authenticated = AuthUtils.isAuthenticated();
            console.log('🔑 Is authenticated:', authenticated);

            if (authenticated) {
              const userInfo = AuthUtils.getUserInfo();
              console.log('👤 User info:', userInfo?.name || 'no user');
              set({
                isAuthenticated: true,
                user: userInfo,
                isLoading: false,
              });
            } else {
              console.log('🔄 Not authenticated, trying to refresh token...');
              // 检查是否有 refresh token，没有就直接跳过
              const hasRefreshToken = AuthUtils.getRefreshToken();
              if (!hasRefreshToken) {
                console.log('❌ No refresh token, setting unauthenticated');
                set({
                  isAuthenticated: false,
                  user: null,
                  isLoading: false,
                });
                return;
              }

              // 尝试刷新 token
              const refreshed = await get().refreshToken();
              console.log('🔄 Refresh result:', refreshed);
              if (!refreshed) {
                set({
                  isAuthenticated: false,
                  user: null,
                  isLoading: false,
                });
              }
            }
          } catch (error) {
            console.error('❌ Auth check failed:', error);
            set({
              isAuthenticated: false,
              user: null,
              isLoading: false,
            });
          }
        },

        // 初始化认证状态
        initialize: () => {
          console.log('🔐 Initialize called, isInitialized:', isInitialized);

          // 防止重复初始化（使用模块级变量，更轻量）
          if (isInitialized) {
            console.log('⚠️ Already initialized, skipping');
            return;
          }

          console.log('✅ Starting initialization');
          isInitialized = true;

          // 移除旧的监听器（如果存在）
          if (storageListener) {
            window.removeEventListener('storage', storageListener);
          }

          // 立即执行认证检查
          get().checkAuthStatus();

          // 监听 storage 变化（支持多 tab 同步）
          // 注意：storage 事件只在其他 tab 修改 localStorage 时触发
          // 当前 tab 的修改不会触发此事件（浏览器标准行为）
          storageListener = (e: StorageEvent) => {
            if (e.key === 'auth-storage') {
              console.log(
                '📦 Storage changed from another tab, re-checking auth'
              );
              get().checkAuthStatus();
            }
          };

          window.addEventListener('storage', storageListener);
        },
      }),
      {
        name: 'auth-storage',
        // 只持久化用户数据，不持久化加载状态
        partialize: (state) => ({
          user: state.user,
          isAuthenticated: state.isAuthenticated,
        }),
        // 重要：在恢复状态后，确保 isLoading 为 false
        onRehydrateStorage: () => (state) => {
          if (state) {
            console.log('💾 State rehydrated from storage');
            state.isLoading = false; // 恢复后确保不是加载状态
          }
        },
      }
    ),
    {
      name: 'auth-store',
    }
  )
);
