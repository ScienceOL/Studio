'use client';

import { useUI } from '@/hooks/useUI';
import { useEffect, useState } from 'react';

export function useTheme() {
  const { theme } = useUI();
  const [isDark, setIsDark] = useState(false);

  useEffect(() => {
    // 根据主题设置 isDark
    const updateIsDark = () => {
      setIsDark(theme === 'dark' || (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches));
    };

    updateIsDark();

    // 监听系统主题变化（当主题设置为 system 时）
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', updateIsDark);

    return () => {
      mediaQuery.removeEventListener('change', updateIsDark);
    };
  }, [theme]);

  return { isDark };
}
