import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { AuthUtils } from '@/lib/auth';

export default function LoginCallback() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
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

        console.log('Callback params:', { token, refreshToken, expiresIn, userEncoded, error });

        // 检查是否有错误
        if (error) {
          const errorMsg = decodeURIComponent(error);
          setStatus('error');
          
          // 提供更友好的错误提示
          if (errorMsg.includes('state verification failed')) {
            setMessage('登录验证失败：State 验证失败。这通常是因为 Casdoor 回调地址配置不正确，请检查 Casdoor 应用的 Redirect URL 是否设置为后端地址。');
          } else {
            setMessage(`登录失败: ${errorMsg}`);
          }
          
          console.error('OAuth2 callback error:', errorMsg);
          return;
        }

        // 检查必要的参数
        if (!token || !refreshToken || !expiresIn) {
          setStatus('error');
          setMessage('登录参数不完整，请重新登录');
          console.error('Missing required params:', { token: !!token, refreshToken: !!refreshToken, expiresIn: !!expiresIn });
          return;
        }

        // 解析用户信息
        let userInfo = null;
        if (userEncoded) {
          try {
            const userJSON = atob(userEncoded);
            userInfo = JSON.parse(userJSON);
            console.log('Parsed user info:', userInfo);
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

        console.log('Auth info saved successfully');
        console.log('Stored access_token:', localStorage.getItem('access_token'));

        setStatus('success');
        setMessage('登录成功，正在跳转...');

        // 延迟跳转到首页
        setTimeout(() => {
          navigate('/');
        }, 1500);
      } catch (error) {
        console.error('Login callback error:', error);
        setStatus('error');
        setMessage('登录处理过程中发生错误');
      }
    };

    handleCallback();
  }, [searchParams, navigate]);

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
