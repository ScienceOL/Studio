import { useUiStore } from '@/store/uiStore';
import { useEffect } from 'react';

/**
 * useTheme hook - 现在使用 Zustand uiStore 来管理主题状态
 * 这样可以确保主题状态在整个应用中保持一致，即使组件被卸载也不会丢失状态
 */
const useTheme = () => {
  const theme = useUiStore((s) => s.theme);
  const cycleTheme = useUiStore((s) => s.cycleTheme);
  const applyTheme = useUiStore((s) => s.applyTheme);

  // 初始化时应用主题
  useEffect(() => {
    applyTheme();
  }, [applyTheme]);

  // 监听系统主题变化
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = () => {
      if (theme === 'system') {
        applyTheme();
      }
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [theme, applyTheme]);

  return { theme, cycleTheme };
};

export default useTheme;
