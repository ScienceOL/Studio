import { AuthUtils } from '@/lib/auth';
import { useEffect, useRef, useState } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

// 辅助函数：从 Cookie 中读取指定名称的值
function getCookie(name: string): string | null {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    const cookieValue = parts.pop()?.split(';').shift() || null;
    // 自动进行 URL 解码
    return cookieValue ? decodeURIComponent(cookieValue) : null;
  }
  return null;
}

export default function LoginCallback() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState<'processing' | 'success' | 'error'>(
    'processing'
  );
  const [message, setMessage] = useState('正在处理登录...');
  const hasProcessed = useRef(false);

  useEffect(() => {
    // 防止重复执行
    if (hasProcessed.current) {
      return;
    }
    hasProcessed.current = true;

    const handleCallback = async () => {
      try {
        // 检查 URL 参数中是否有错误
        const error = searchParams.get('error');
        const status = searchParams.get('status');

        console.log('Callback params:', {
          status,
          error,
        });

        // 检查是否有错误
        if (error) {
          const errorMsg = decodeURIComponent(error);
          setStatus('error');

          // 提供更友好的错误提示
          if (errorMsg.includes('state verification failed')) {
            setMessage(
              '登录验证失败：State 验证失败。这通常是因为 Casdoor 回调地址配置不正确，请检查 Casdoor 应用的 Redirect URL 是否设置为后端地址。'
            );
          } else {
            setMessage(`登录失败: ${errorMsg}`);
          }

          console.error('OAuth2 callback error:', errorMsg);
          return;
        }

        // 检查登录状态
        if (status !== 'success') {
          setStatus('error');
          setMessage('登录状态异常，请重新登录');
          return;
        }

        // 从 Cookie 中读取 token 和用户信息
        console.log('📝 All cookies:', document.cookie);
        console.log(
          '📝 Cookie keys:',
          document.cookie.split(';').map((c) => c.trim().split('=')[0])
        );

        const token = getCookie('access_token');
        const refreshToken = getCookie('refresh_token');
        const userInfoEncoded = getCookie('user_info');

        // 调试 getCookie 函数
        console.log('🔍 Debug getCookie for user_info:');
        console.log('  - Raw cookie string:', document.cookie);
        console.log('  - Looking for user_info...');
        console.log('  - getCookie result:', userInfoEncoded);
        console.log(
          '  - getCookie result length:',
          userInfoEncoded?.length || 0
        );

        console.log('Reading from cookies:', {
          hasToken: !!token,
          hasRefreshToken: !!refreshToken,
          hasUserInfo: !!userInfoEncoded,
          tokenLength: token?.length || 0,
          refreshTokenLength: refreshToken?.length || 0,
        });

        // 检查必要的参数
        if (!token || !refreshToken) {
          setStatus('error');
          setMessage('未能从 Cookie 中获取登录信息，请重新登录');
          console.error('Missing token in cookies');
          return;
        }

        // 解析用户信息
        let userInfo = null;
        if (userInfoEncoded) {
          try {
            const userJSON = atob(userInfoEncoded);
            userInfo = JSON.parse(userJSON);
            console.log('✅ Successfully parsed user info:', userInfo);
          } catch (parseError) {
            console.error('Failed to parse user info from cookie:', parseError);
          }
        }

        // 注意：由于 token 已经在 HTTP-Only Cookie 中，我们不需要再存储到 localStorage
        // 但为了兼容现有的 AuthUtils，我们还是将 token 保存到 localStorage
        // 未来可以考虑完全使用 Cookie 方式

        // 从 Cookie 中读取过期时间（如果后端设置了）
        // 这里我们假设 token 的有效期，可以后续从后端 API 获取
        const expiresIn = 3600; // 默认1小时

        // 保存认证信息和用户信息
        AuthUtils.saveAuthInfo(
          {
            accessToken: token,
            refreshToken: refreshToken,
            expiresIn: expiresIn,
            tokenType: 'Bearer',
          },
          userInfo
        );

        console.log('Auth info saved successfully');

        setStatus('success');
        setMessage('登录成功，正在跳转...');

        // 获取登录前的页面路径
        const returnUrl = sessionStorage.getItem('login_return_url');
        console.log(
          '📖 Reading from sessionStorage - login_return_url:',
          returnUrl
        );
        console.log('📖 location.state:', location.state);

        const from =
          returnUrl || (location.state as { from?: string })?.from || '/';
        console.log('🎯 Final redirect target:', from);

        // 清除保存的返回 URL
        sessionStorage.removeItem('login_return_url');
        console.log('🗑️ Cleared sessionStorage');

        // 延迟跳转
        setTimeout(() => {
          console.log('🚀 Navigating to:', from);
          navigate(from, { replace: true });
        }, 1500);
      } catch (error) {
        console.error('Login callback error:', error);
        setStatus('error');
        setMessage('登录处理过程中发生错误');
      }
    };

    handleCallback();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // 只在组件挂载时执行一次

  const handleRetry = () => {
    AuthUtils.redirectToLogin();
  };

  const handleGoHome = () => {
    navigate('/');
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
