// src/lib/theme.ts

import { useEffect, useState } from 'react';

export type Theme = 'light' | 'dark' | 'system';

const THEME_KEY = 'theme';

function getSystemTheme(): Theme {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light';
}

export function getTheme(): Theme {
  const theme = localStorage.getItem(THEME_KEY);
  if (theme === 'light' || theme === 'dark') return theme;
  return 'system';
}

export function applyTheme(theme: Theme) {
  let effective = theme;
  if (theme === 'system') effective = getSystemTheme();
  document.documentElement.classList.toggle('dark', effective === 'dark');
}

export function setTheme(theme: Theme) {
  localStorage.setItem(THEME_KEY, theme);
  applyTheme(theme);
}

export function useTheme(): [Theme, (t: Theme) => void] {
  const [theme, setThemeState] = useState<Theme>(() => getTheme());

  useEffect(() => {
    setTheme(theme);
    if (theme === 'system') {
      const mq = window.matchMedia('(prefers-color-scheme: dark)');
      const handler = () => applyTheme('system');
      mq.addEventListener('change', handler);
      return () => mq.removeEventListener('change', handler);
    }
  }, [theme]);

  return [theme, setThemeState];
}
