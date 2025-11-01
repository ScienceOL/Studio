import DashboardLayout from '@/app/dashboard';
import Providers from '@/provider';

/**
 * 受保护的 Dashboard 布局包装器
 * 结合了登录保护和 Dashboard 布局
 * 使用 Outlet 渲染嵌套路由
 */
export default function ProtectedDashboardLayout() {
  return (
    <Providers>
      <DashboardLayout />
    </Providers>
  );
}
