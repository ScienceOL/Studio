'use client';

import { useRouter, useSearchParams } from 'next/navigation';
import { Suspense, useEffect, useState } from 'react';
import { AuthUtils } from '../../../lib/auth';

function LoginCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<'processing' | 'success' | 'error'>(
    'processing'
  );
  const [message, setMessage] = useState('正在处理登录...');

  useEffect(() => {
    const handleCallback = async () => {
      try {
        // 从 URL 参数中获取 token 信息
        const token = searchParams.get('token');
        const refreshToken = searchParams.get('refresh_token');
        const expiresIn = searchParams.get('expires_in');
        const userEncoded = searchParams.get('user');
        const error = searchParams.get('error');

        // 检查是否有错误
        if (error) {
          setStatus('error');
          setMessage(`登录失败: ${decodeURIComponent(error)}`);
          return;
        }

        // 检查必要的参数
        if (!token || !refreshToken || !expiresIn) {
          setStatus('error');
          setMessage('登录参数不完整，请重新登录');
          return;
        }

        // 解析用户信息
        let userInfo = null;
        if (userEncoded) {
          try {
            const userJSON = atob(userEncoded);
            userInfo = JSON.parse(userJSON);
          } catch (parseError) {
            console.warn('Failed to parse user info from URL:', parseError);
          }
        }

        // 保存认证信息和用户信息
        AuthUtils.saveAuthInfo(
          {
            accessToken: token,
            refreshToken: refreshToken,
            expiresIn: parseInt(expiresIn, 10),
            tokenType: 'Bearer',
          },
          userInfo
        );

        // 如果没有从URL获取到用户信息，尝试通过API获取
        if (!userInfo) {
          try {
            const response = await fetch(
              `${window.location.protocol}//${window.location.host}/api/user/info`,
              {
                headers: {
                  Authorization: `Bearer ${token}`,
                },
              }
            );

            if (response.ok) {
              const userData = await response.json();
              if (userData.code === 0 && userData.data) {
                // 更新用户信息
                AuthUtils.saveAuthInfo(
                  {
                    accessToken: token,
                    refreshToken: refreshToken,
                    expiresIn: parseInt(expiresIn, 10),
                    tokenType: 'Bearer',
                  },
                  userData.data
                );
              }
            }
          } catch (userInfoError) {
            console.warn('Failed to fetch user info:', userInfoError);
            // 不阻止登录流程，用户信息可以后续获取
          }
        }

        setStatus('success');
        setMessage('登录成功，正在跳转...');

        // 延迟跳转到首页
        setTimeout(() => {
          router.push('/');
        }, 1500);
      } catch (error) {
        console.error('Login callback error:', error);
        setStatus('error');
        setMessage('登录处理过程中发生错误');
      }
    };

    handleCallback();
  }, [searchParams, router]);

  const handleRetry = () => {
    AuthUtils.redirectToLogin();
  };

  const handleGoHome = () => {
    router.push('/');
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <h2 className="mt-6 text-3xl font-extrabold text-gray-900 dark:text-white">
            OAuth2 登录回调
          </h2>
        </div>

        <div className="bg-white dark:bg-gray-800 shadow rounded-lg px-8 py-6">
          <div className="text-center">
            {status === 'processing' && (
              <div className="space-y-4">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
                <p className="text-gray-600 dark:text-gray-400">{message}</p>
              </div>
            )}

            {status === 'success' && (
              <div className="space-y-4">
                <div className="w-12 h-12 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center mx-auto">
                  <svg
                    className="w-6 h-6 text-green-600 dark:text-green-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                </div>
                <p className="text-green-600 dark:text-green-400 font-medium">
                  {message}
                </p>
              </div>
            )}

            {status === 'error' && (
              <div className="space-y-4">
                <div className="w-12 h-12 bg-red-100 dark:bg-red-900 rounded-full flex items-center justify-center mx-auto">
                  <svg
                    className="w-6 h-6 text-red-600 dark:text-red-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </div>
                <p className="text-red-600 dark:text-red-400 font-medium">
                  {message}
                </p>
                <div className="flex space-x-4 justify-center">
                  <button
                    onClick={handleRetry}
                    className="
                      px-4 py-2 text-sm font-medium
                      text-white bg-blue-600 border border-transparent rounded-md
                      hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500
                      transition-colors
                    "
                  >
                    重新登录
                  </button>
                  <button
                    onClick={handleGoHome}
                    className="
                      px-4 py-2 text-sm font-medium
                      text-gray-700 bg-gray-100 border border-gray-300 rounded-md
                      hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500
                      dark:text-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600
                      transition-colors
                    "
                  >
                    返回首页
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="text-center">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            如果长时间未响应，请{' '}
            <button
              onClick={handleRetry}
              className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline"
            >
              重新登录
            </button>
          </p>
        </div>
      </div>
    </div>
  );
}

function LoadingFallback() {
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <h2 className="mt-6 text-3xl font-extrabold text-gray-900 dark:text-white">
            OAuth2 登录回调
          </h2>
        </div>
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg px-8 py-6">
          <div className="text-center space-y-4">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
            <p className="text-gray-600 dark:text-gray-400">正在加载...</p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function LoginCallback() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <LoginCallbackContent />
    </Suspense>
  );
}
