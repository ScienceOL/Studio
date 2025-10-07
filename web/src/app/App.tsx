import LogoLoading from '@/components/basic/loading';
import { useAuthStore } from '@/store/authStore';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import LandscapePage from './landscape';

export default function App() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const isLoading = useAuthStore((s) => s.isLoading);
  const navigate = useNavigate();

  // 初始化已在 main.tsx 中完成，这里只负责渲染
  console.log('🔄 App render:', { isAuthenticated, isLoading });

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
