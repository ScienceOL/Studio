import '@/i18n';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import Router from './router';
import { useAuthStore } from './store/authStore';

// 应用启动时初始化认证状态（应用级，只执行一次）
console.log('🚀 Application starting, initializing auth...');
useAuthStore.getState().initialize();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Router />
  </StrictMode>
);
