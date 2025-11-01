import { AuthUtils } from '@/utils/auth';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

export default function LoginPage() {
  const navigate = useNavigate();

  useEffect(() => {
    // 如果已经登录，重定向到首页
    if (AuthUtils.isAuthenticated()) {
      navigate('/');
    }
  }, [navigate]);

  const handleLogin = () => {
    // 重定向到后端登录接口
    AuthUtils.redirectToLogin();
  };

  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-neutral-900 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <h2 className="mt-6 text-3xl font-extrabold text-neutral-900 dark:text-white">
            欢迎使用 Studio
          </h2>
          <p className="mt-2 text-sm text-neutral-600 dark:text-neutral-400">
            请登录以继续使用
          </p>
        </div>

        <div className="bg-white dark:bg-neutral-800 shadow rounded-lg px-8 py-6">
          <div className="space-y-6">
            <div className="text-center">
              <p className="text-sm text-neutral-600 dark:text-neutral-400 mb-6">
                点击下方按钮使用 OAuth2 登录
              </p>
              <button
                onClick={handleLogin}
                className="
                  w-full flex justify-center items-center
                  px-4 py-3 text-base font-medium
                  text-white bg-indigo-600 border border-transparent rounded-md
                  hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500
                  transition-all duration-200
                  shadow-md hover:shadow-lg
                "
              >
                <svg
                  className="w-5 h-5 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1"
                  />
                </svg>
                使用 OAuth2 登录
              </button>
            </div>

            <div className="border-t border-neutral-200 dark:border-neutral-700 pt-6">
              <div className="text-xs text-neutral-500 dark:text-neutral-400 space-y-2">
                <p>🔒 使用 Casdoor 进行安全认证</p>
                <p>🚀 登录后将自动跳转到首页</p>
                <p>💾 Token 将安全存储在本地</p>
              </div>
            </div>
          </div>
        </div>

        <div className="text-center">
          <p className="text-xs text-neutral-500 dark:text-neutral-400">
            登录即表示您同意我们的服务条款和隐私政策
          </p>
        </div>
      </div>
    </div>
  );
}
