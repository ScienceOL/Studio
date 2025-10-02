import { useEffect } from 'react';
import { useAuthStore } from '@/store/authStore';
import type { UserInfo } from '@/lib/auth';

export interface UseAuthReturn {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: UserInfo | null;
  login: () => void;
  logout: () => void;
  refreshToken: () => Promise<boolean>;
}

export function useAuth(): UseAuthReturn {
  const {
    isAuthenticated,
    isLoading,
    user,
    login,
    logout,
    refreshToken,
    initialize,
  } = useAuthStore();

  // 组件挂载时初始化
  useEffect(() => {
    initialize();
  }, [initialize]);

  return {
    isAuthenticated,
    isLoading,
    user,
    login,
    logout,
    refreshToken,
  };
}
