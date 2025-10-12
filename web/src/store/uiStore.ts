import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

// 侧边栏宽度配置
const SIDEBAR_MIN_WIDTH = 64; // 最小宽度 (折叠状态)
const SIDEBAR_MAX_WIDTH = 320; // 最大宽度
const SIDEBAR_DEFAULT_WIDTH = 256; // 默认宽度

// 主题类型
export type Theme = 'light' | 'dark' | 'system';

export interface UiState {
  // 侧边栏状态
  sidebarWidth: number;
  isSidebarCollapsed: boolean;
  isSidebarHovered: boolean;

  // 主题状态
  theme: Theme;

  // 内部状态
  _hasHydrated: boolean;

  // 侧边栏动作
  setSidebarWidth: (width: number) => void;
  toggleSidebarCollapsed: () => void;
  setSidebarHovered: (hovered: boolean) => void;
  resetSidebarWidth: () => void;

  // 主题动作
  setTheme: (theme: Theme) => void;
  cycleTheme: () => void;
  applyTheme: (theme?: Theme) => void;

  // 内部动作
  setHasHydrated: (hasHydrated: boolean) => void;

  // 计算属性
  getEffectiveSidebarWidth: () => number;
}

export const useUiStore = create<UiState>()(
  devtools(
    persist(
      (set, get) => ({
        // 初始状态
        sidebarWidth: SIDEBAR_DEFAULT_WIDTH,
        isSidebarCollapsed: false,
        isSidebarHovered: false,
        theme: 'system',
        _hasHydrated: false,

        // 设置侧边栏宽度（带约束）
        setSidebarWidth: (width) => {
          const constrainedWidth = Math.max(
            SIDEBAR_MIN_WIDTH,
            Math.min(SIDEBAR_MAX_WIDTH, width)
          );
          set({ sidebarWidth: constrainedWidth });
        },

        // 切换折叠状态
        toggleSidebarCollapsed: () => {
          set((state) => ({
            isSidebarCollapsed: !state.isSidebarCollapsed,
          }));
        },

        // 设置悬停状态
        setSidebarHovered: (hovered) => {
          set({ isSidebarHovered: hovered });
        },

        // 重置宽度为默认值
        resetSidebarWidth: () => {
          set({ sidebarWidth: SIDEBAR_DEFAULT_WIDTH });
        },

        // 获取有效宽度（考虑折叠和悬停状态）
        getEffectiveSidebarWidth: () => {
          const state = get();
          if (state.isSidebarCollapsed && !state.isSidebarHovered) {
            return SIDEBAR_MIN_WIDTH;
          }
          return state.sidebarWidth;
        },

        // 设置主题
        setTheme: (theme) => {
          set({ theme });
          get().applyTheme(theme);
        },

        // 循环切换主题
        cycleTheme: () => {
          const themes: Theme[] = ['light', 'dark', 'system'];
          const currentTheme = get().theme;
          const currentIndex = themes.indexOf(currentTheme);
          const nextTheme = themes[(currentIndex + 1) % themes.length];
          get().setTheme(nextTheme);
        },

        // 设置恢复状态
        setHasHydrated: (hasHydrated) => {
          set({ _hasHydrated: hasHydrated });
        },

        // 应用主题到 DOM
        applyTheme: (selectedTheme) => {
          const theme = selectedTheme ?? get().theme;
          const root = window.document.documentElement;
          const isDark =
            theme === 'dark' ||
            (theme === 'system' &&
              window.matchMedia('(prefers-color-scheme: dark)').matches);

          // 注入CSS以禁用过渡
          const style = document.createElement('style');
          style.innerHTML =
            '*, *::before, *::after { transition: none !important; }';
          document.head.appendChild(style);

          root.classList.toggle('dark', isDark);

          // 在短时间后移除样式，以恢复过渡效果
          setTimeout(() => {
            if (document.head.contains(style)) {
              document.head.removeChild(style);
            }
          }, 50);
        },
      }),
      {
        name: 'ui-storage',
        // 持久化宽度、折叠状态和主题
        partialize: (state) => ({
          sidebarWidth: state.sidebarWidth,
          isSidebarCollapsed: state.isSidebarCollapsed,
          theme: state.theme,
        }),
        // 恢复完成后的回调
        onRehydrateStorage: () => (state) => {
          if (state) {
            state.setHasHydrated(true);
            // 恢复完成后立即应用主题
            state.applyTheme();
          }
        },
      }
    ),
    {
      name: 'ui-store',
    }
  )
);

// 导出常量供组件使用
export const SIDEBAR_CONFIG = {
  MIN_WIDTH: SIDEBAR_MIN_WIDTH,
  MAX_WIDTH: SIDEBAR_MAX_WIDTH,
  DEFAULT_WIDTH: SIDEBAR_DEFAULT_WIDTH,
} as const;
