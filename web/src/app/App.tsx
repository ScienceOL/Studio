/**
 * 🎨 Component Layer - Entrypoint
 */

import LogoLoading from '@/components/basic/loading';
import { useAuth } from '@/hooks/useAuth';

import { useUI } from '@/hooks/useUI';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import LandscapePage from './landscape';

export default function App() {
  const { isAuthenticated, isLoading } = useAuth();

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
