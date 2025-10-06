import { BrowserRouter, Route, Routes } from 'react-router-dom';
import App from './app/App';
import DashboardHome from './app/dashboard/Home';
import CallbackPage from './app/login/CallbackPage';
import LoginPage from './app/login/LoginPage';
import UiTestPage from './app/ui/page';
import ProtectedDashboardLayout from './components/layout/ProtectedDashboardPage';

export default function Router() {
  return (
    <BrowserRouter>
      <Routes>
        {/* 根路径 - App 组件根据登录状态分流 */}
        <Route path="/" element={<App />} />

        {/* 公开路由 */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/login/callback" element={<CallbackPage />} />
        <Route path="/ui-test" element={<UiTestPage />} />

        {/* 所有需要侧边栏和登录保护的页面 */}
        <Route element={<ProtectedDashboardLayout />}>
          <Route index element={<DashboardHome />} />
          <Route path="/space" element={<div>Space Page</div>} />
          <Route path="/discussion" element={<div>Discussion Page</div>} />
          <Route path="/manuscript" element={<div>Manuscript Page</div>} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
