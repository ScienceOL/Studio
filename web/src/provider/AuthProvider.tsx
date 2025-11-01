import { AuthCore } from '@/core/authCore';
import { AuthUtils } from '@/utils/auth';
import { useEffect, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

interface AuthProviderProps {
  children: React.ReactNode;
  /**
   * 是否需要认证才能访问
   * @default true
   */
  requireAuth?: boolean;
  /**
   * 未认证时重定向的路径
   * @default '/login'
   */
  redirectTo?: string;
  /**
   * 是否显示弹窗提示
   * @default true
   */
  showModal?: boolean;
  /**
   * 弹窗延迟时间（毫秒）
   * @default 3000
   */
  modalDelay?: number;
}

export default function AuthProvider({
  children,
  requireAuth = true,
  redirectTo = '/login',
  showModal = true,
  modalDelay = 3000,
}: AuthProviderProps) {
  const [isChecking, setIsChecking] = useState(requireAuth);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [showModalState, setShowModalState] = useState(false);
  const [countdown, setCountdown] = useState(Math.ceil(modalDelay / 1000));
  const location = useLocation();
  const navigate = useNavigate();
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const countdownIntervalRef = useRef<ReturnType<typeof setInterval> | null>(
    null
  );

  useEffect(() => {
    if (!requireAuth) {
      setIsChecking(false);
      return;
    }

    const checkAuth = async () => {
      const authenticated = await AuthCore.checkAuthStatus();
      setIsAuthenticated(authenticated);
      setIsChecking(false);
    };
    checkAuth();
  }, [requireAuth]);

  useEffect(() => {
    if (!requireAuth || isChecking || isAuthenticated) return;

    if (showModal) {
      setShowModalState(true);

      // 启动倒计时
      let remainingTime = Math.ceil(modalDelay / 1000);
      setCountdown(remainingTime);

      countdownIntervalRef.current = setInterval(() => {
        remainingTime -= 1;
        setCountdown(remainingTime);
        if (remainingTime <= 0 && countdownIntervalRef.current) {
          clearInterval(countdownIntervalRef.current);
        }
      }, 1000);

      timerRef.current = setTimeout(() => {
        // 保存当前路径并跳转到登录
        console.log(
          '🔐 Saving return URL to sessionStorage:',
          location.pathname
        );
        AuthUtils.redirectToLogin(location.pathname);
      }, modalDelay);
    } else {
      // 保存当前路径并跳转到登录
      console.log('🔐 Saving return URL to sessionStorage:', location.pathname);
      AuthUtils.redirectToLogin(location.pathname);
    }

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
      if (countdownIntervalRef.current)
        clearInterval(countdownIntervalRef.current);
    };
  }, [
    requireAuth,
    isChecking,
    isAuthenticated,
    showModal,
    modalDelay,
    redirectTo,
    location,
    navigate,
  ]);

  // 不需要认证，直接渲染
  if (!requireAuth) {
    return <>{children}</>;
  }

  // 正在检查认证状态
  if (isChecking) {
    return (
      <div className="min-h-screen bg-neutral-50 dark:bg-neutral-900 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-neutral-600 dark:text-neutral-400">
            正在验证身份...
          </p>
        </div>
      </div>
    );
  }

  // 显示弹窗提示
  if (showModalState) {
    const progress =
      ((modalDelay / 1000 - countdown) / (modalDelay / 1000)) * 100;

    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40">
        <div className="bg-white dark:bg-neutral-800 rounded-lg shadow-lg p-8 text-center max-w-md relative overflow-hidden">
          {/* 进度条 */}
          <div
            className="absolute top-0 left-0 h-1 bg-indigo-600 transition-all duration-1000 ease-linear"
            style={{ width: `${progress}%` }}
          />

          <h2 className="text-lg font-semibold mb-4 text-neutral-900 dark:text-neutral-100">
            访问的内容需要登录后才可以访问
          </h2>
          <p className="mb-6 text-neutral-600 dark:text-neutral-300">
            请先登录后再访问该页面
          </p>

          {/* 倒计时圆环 */}
          <div className="flex items-center justify-center mb-6">
            <div className="relative">
              <svg className="w-20 h-20 transform -rotate-90">
                <circle
                  cx="40"
                  cy="40"
                  r="36"
                  stroke="currentColor"
                  strokeWidth="4"
                  fill="none"
                  className="text-neutral-200 dark:text-neutral-700"
                />
                <circle
                  cx="40"
                  cy="40"
                  r="36"
                  stroke="currentColor"
                  strokeWidth="4"
                  fill="none"
                  strokeDasharray={`${2 * Math.PI * 36}`}
                  strokeDashoffset={`${
                    2 * Math.PI * 36 * (1 - progress / 100)
                  }`}
                  className="text-indigo-600 transition-all duration-1000 ease-linear"
                  strokeLinecap="round"
                />
              </svg>
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="text-2xl font-bold text-indigo-600">
                  {countdown}
                </span>
              </div>
            </div>
          </div>

          <button
            className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 transition-colors"
            onClick={() => {
              if (timerRef.current) clearTimeout(timerRef.current);
              if (countdownIntervalRef.current)
                clearInterval(countdownIntervalRef.current);
              // 保存当前路径并跳转到登录
              AuthUtils.redirectToLogin(location.pathname);
            }}
          >
            立即登录
          </button>
        </div>
      </div>
    );
  }

  // 已认证，渲染子组件
  return <>{children}</>;
}
