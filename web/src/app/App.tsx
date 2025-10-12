import LogoLoading from '@/components/basic/loading';
import { useAuthStore } from '@/store/authStore';
import { useUiStore } from '@/store/uiStore';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import LandscapePage from './landscape';

export default function App() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const isLoading = useAuthStore((s) => s.isLoading);
  const hasHydrated = useUiStore((s) => s._hasHydrated);
  const applyTheme = useUiStore((s) => s.applyTheme);
  const navigate = useNavigate();

  // 等待 Zustand 状态恢复完成后再应用主题
  useEffect(() => {
    if (hasHydrated) {
      applyTheme();
    }
  }, [hasHydrated, applyTheme]);

  // 初始化已在 main.tsx 中完成，这里只负责渲染
  console.log('🔄 App render:', { isAuthenticated, isLoading, hasHydrated });

  // 如果已登录，重定向到 dashboard
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, isLoading, navigate]);

  if (isLoading) {
    return <LogoLoading variant="large" animationType="galaxy" />;
  }

  // 未登录显示 Landscape
  return <LandscapePage />;
}
