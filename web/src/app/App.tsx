/**
 * 🎨 Component Layer - 应用入口（重构版）
 *
 * 职责：
 * 1. 只负责 UI 渲染
 * 2. 通过 Hook 获取状态和方法
 * 3. 处理用户交互事件
 */

import LogoLoading from '@/components/basic/loading';
import { useAuth } from '@/hooks/useAuth';

import { useUI } from '@/hooks/useUI';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import LandscapePage from './landscape';

export default function App() {
  // 🎣 使用认证 Hook
  const { isAuthenticated, isLoading } = useAuth();
  // 🎣 使用 UI Hook，它会在内部处理主题初始化等副作用
  useUI();

  const navigate = useNavigate();

  // 认证重定向逻辑
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, isLoading, navigate]);

  // 加载中状态
  if (isLoading) {
    return (
      <div className="h-screen w-screen flex items-center justify-center">
        <LogoLoading variant="large" animationType="galaxy" />
      </div>
    );
  }

  // 渲染落地页
  return <LandscapePage />;
}
