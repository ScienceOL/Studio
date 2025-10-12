import { useUiStore } from '@/store/uiStore';

/**
 * useTheme hook - 现在使用 Zustand uiStore 来管理主题状态
 * 主题初始化和系统主题监听已移动到 App.tsx 中处理，确保全局一致性
 */
const useTheme = () => {
  const theme = useUiStore((s) => s.theme);
  const cycleTheme = useUiStore((s) => s.cycleTheme);

  return { theme, cycleTheme };
};

export default useTheme;
