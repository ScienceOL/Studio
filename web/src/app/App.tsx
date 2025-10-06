import LogoLoading from '@/components/basic/loading';
import { useAuthStore } from '@/store/authStore';
import DashboardLayout from './dashboard';
import LandscapePage from './landscape';

export default function App() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const isLoading = useAuthStore((s) => s.isLoading);

  // 初始化已在 main.tsx 中完成，这里只负责渲染
  console.log('🔄 App render:', { isAuthenticated, isLoading });

  if (isLoading) {
    return <LogoLoading variant="large" animationType="galaxy" />;
  }

  // 根据认证状态分流：已登录显示 Dashboard，未登录显示 Landscape
  return isAuthenticated ? <DashboardLayout /> : <LandscapePage />;
}
