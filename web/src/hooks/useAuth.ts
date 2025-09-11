'use client';

import { useCallback, useEffect, useState } from 'react';
import { AuthUtils, UserInfo } from '../lib/auth';

export interface UseAuthReturn {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: UserInfo | null;
  login: () => void;
  logout: () => void;
  refreshToken: () => Promise<boolean>;
}

export function useAuth(): UseAuthReturn {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState<UserInfo | null>(null);

  // 检查认证状态
  const checkAuthStatus = useCallback(async () => {
    setIsLoading(true);

    try {
      const authenticated = AuthUtils.isAuthenticated();
      setIsAuthenticated(authenticated);

      if (authenticated) {
        const userInfo = AuthUtils.getUserInfo();
        setUser(userInfo);
      } else {
        setUser(null);
        // 尝试刷新 token
        const refreshed = await AuthUtils.refreshToken();
        if (refreshed) {
          setIsAuthenticated(true);
          const userInfo = AuthUtils.getUserInfo();
          setUser(userInfo);
        }
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      setIsAuthenticated(false);
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // 登录（重定向到后端）
  const login = useCallback(() => {
    AuthUtils.redirectToLogin();
  }, []);

  // 登出
  const logout = useCallback(() => {
    AuthUtils.logout();
    setIsAuthenticated(false);
    setUser(null);
  }, []);

  // 刷新 token
  const refreshToken = useCallback(async (): Promise<boolean> => {
    const success = await AuthUtils.refreshToken();
    if (success) {
      setIsAuthenticated(true);
      const userInfo = AuthUtils.getUserInfo();
      setUser(userInfo);
    } else {
      setIsAuthenticated(false);
      setUser(null);
    }
    return success;
  }, []);

  // 初始化时检查认证状态
  useEffect(() => {
    checkAuthStatus();
  }, [checkAuthStatus]);

  // 监听 storage 变化（支持多 tab 同步）
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (
        e.key?.startsWith('studio_') ||
        e.key === null // localStorage.clear() was called
      ) {
        checkAuthStatus();
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, [checkAuthStatus]);

  return {
    isAuthenticated,
    isLoading,
    user,
    login,
    logout,
    refreshToken,
  };
}
