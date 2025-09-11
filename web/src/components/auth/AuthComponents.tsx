'use client';

import Image from 'next/image';
import { useAuthContext } from '../../providers/AuthProvider';

export interface LoginButtonProps {
  className?: string;
  children?: React.ReactNode;
}

export function LoginButton({ className, children }: LoginButtonProps) {
  const { login, isLoading } = useAuthContext();

  const handleLogin = () => {
    if (!isLoading) {
      login();
    }
  };

  return (
    <button
      onClick={handleLogin}
      disabled={isLoading}
      className={`
        inline-flex items-center justify-center
        px-4 py-2 text-sm font-medium
        text-white bg-blue-600 border border-transparent rounded-md
        hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500
        disabled:opacity-50 disabled:cursor-not-allowed
        transition-colors
        ${className || ''}
      `}
    >
      {isLoading ? '加载中...' : children || '登录'}
    </button>
  );
}

export interface UserInfoProps {
  className?: string;
}

export function UserInfo({ className }: UserInfoProps) {
  const { isAuthenticated, user, logout, isLoading } = useAuthContext();

  if (isLoading) {
    return (
      <div className={`flex items-center space-x-2 ${className || ''}`}>
        <div className="animate-pulse">
          <div className="h-8 w-8 bg-gray-300 rounded-full"></div>
        </div>
      </div>
    );
  }

  if (!isAuthenticated || !user) {
    return <LoginButton className={className} />;
  }

  return (
    <div className={`flex items-center space-x-3 ${className || ''}`}>
      {user.avatar && (
        <Image
          src={user.avatar}
          alt={user.displayName}
          width={32}
          height={32}
          className="rounded-full"
        />
      )}
      <div className="flex flex-col">
        <span className="text-sm font-medium text-gray-900 dark:text-white">
          {user.displayName}
        </span>
        <span className="text-xs text-gray-500 dark:text-gray-400">
          {user.email}
        </span>
      </div>
      <button
        onClick={logout}
        className="
          px-3 py-1 text-xs font-medium
          text-gray-600 bg-gray-100 border border-gray-300 rounded
          hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500
          dark:text-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600
          transition-colors
        "
      >
        登出
      </button>
    </div>
  );
}

export interface AuthGuardProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export function AuthGuard({ children, fallback }: AuthGuardProps) {
  const { isAuthenticated, isLoading } = useAuthContext();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <span className="ml-2 text-gray-600">加载中...</span>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen space-y-4">
        {fallback || (
          <>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              需要登录
            </h1>
            <p className="text-gray-600 dark:text-gray-400">
              请登录以继续使用应用
            </p>
            <LoginButton />
          </>
        )}
      </div>
    );
  }

  return <>{children}</>;
}
