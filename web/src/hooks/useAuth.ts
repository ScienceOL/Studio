/**
 * 🎣 Hook Layer - 认证能力暴露
 *
 * 职责：
 * 1. 封装 Core 层的业务能力
 * 2. 提供响应式的状态订阅
 * 3. 暴露简洁的 API 给组件
 */

import { AuthCore } from '@/core/authCore';
import { useAuthStore, type AuthState } from '@/store/authStore';
import { useCallback, useEffect } from 'react';

export interface UseAuthReturn {
  // 状态
  isAuthenticated: boolean;
  isLoading: boolean;
  user: AuthState['user'];

  // 派生状态
  isAdmin: boolean;
  userName: string | null;
  userAvatar: string | null;

  // 动作
  login: (returnUrl?: string) => void;
  logout: () => void;
  refreshToken: () => Promise<boolean>;
  checkAuthStatus: () => Promise<boolean>;
}

/**
 * 认证 Hook
 * 组件通过这个 Hook 访问所有认证相关功能
 */
export function useAuth(): UseAuthReturn {
  // 订阅 Store 状态
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isLoading = useAuthStore((state) => state.isLoading);
  const user = useAuthStore((state) => state.user);

  // 组件挂载时初始化（仅执行一次）
  useEffect(() => {
    let mounted = true;

    const init = async () => {
      if (mounted) {
        await AuthCore.initialize();
      }
    };

    init();

    return () => {
      mounted = false;
    };
  }, []);

  // 封装 Core 层方法
  const login = useCallback((returnUrl?: string) => {
    AuthCore.login(returnUrl);
  }, []);

  const logout = useCallback(() => {
    AuthCore.logout();
  }, []);

  const refreshToken = useCallback(async () => {
    return await AuthCore.refreshToken();
  }, []);

  const checkAuthStatus = useCallback(async () => {
    return await AuthCore.checkAuthStatus();
  }, []);

  // 派生状态
  const isAdmin = AuthCore.isAdmin();
  const userName = user?.name || user?.displayName || null;
  const userAvatar = user?.avatar || null;

  return {
    // 状态
    isAuthenticated,
    isLoading,
    user,

    // 派生状态
    isAdmin,
    userName,
    userAvatar,

    // 动作
    login,
    logout,
    refreshToken,
    checkAuthStatus,
  };
}

/**
 * 登录回调 Hook（用于登录回调页面）
 */
export function useAuthCallback() {
  const handleCallback = useCallback(
    async (params: {
      accessToken: string;
      refreshToken: string;
      expiresIn: number;
    }) => {
      await AuthCore.handleLoginCallback(params);
    },
    []
  );

  return { handleCallback };
}
