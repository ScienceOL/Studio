import { lazy, Suspense } from 'react';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import App from './app/App';
import ProtectedDashboardLayout from './components/layout/ProtectedDashboardPage';

// 路由懒加载
const ChatPage = lazy(() => import('./app/chat/page'));
const EnvironmentPage = lazy(() =>
  import('./app/dashboard/environment').then((module) => ({
    default: module.EnvironmentPage,
  }))
);
const EnvironmentDetail = lazy(
  () => import('./app/dashboard/environment/EnvironmentDetail')
);
const DashboardHome = lazy(() => import('./app/dashboard/Home'));
const CallbackPage = lazy(() => import('./app/login/CallbackPage'));
const LoginPage = lazy(() => import('./app/login/LoginPage'));
const UiTestPage = lazy(() => import('./app/ui/page'));
const Lab3DPage = lazy(() => import('./app/3D_lab/page'));

const LoadingFallback = () => (
  <div className="flex h-screen w-full items-center justify-center bg-gray-50">
    <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-200 border-t-blue-500"></div>
  </div>
);

export default function Router() {
  return (
    <BrowserRouter>
      <Suspense fallback={<LoadingFallback />}>
        <Routes>
          {/* 根路径 - App 组件根据登录状态分流 */}
          <Route path="/" element={<App />} />

          {/* 公开路由 */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/login/callback" element={<CallbackPage />} />
          <Route path="/ui-test" element={<UiTestPage />} />
          <Route path="/chat" element={<ChatPage />} />
          <Route path="/3D_lab" element={<Lab3DPage />} />

          {/* 所有需要侧边栏和登录保护的页面 */}
          <Route element={<ProtectedDashboardLayout />}>
            <Route path="/dashboard" element={<DashboardHome />} />
            <Route
              path="/dashboard/environment"
              element={<EnvironmentPage />}
            />
            <Route
              path="/dashboard/environment/:labUuid"
              element={<EnvironmentDetail />}
            />
          </Route>
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
