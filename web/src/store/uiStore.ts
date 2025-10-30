import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

// Sidebar Width Configuration
export const SIDEBAR_CONFIG = {
  MIN_WIDTH: 64,
  MAX_WIDTH: 320,
  DEFAULT_WIDTH: 256,
} as const;

// Theme Types
export type Theme = 'light' | 'dark' | 'system';

export interface UiState {
  // ===== Status =====
  sidebarWidth: number;
  isSidebarCollapsed: boolean;
  isSidebarHovered: boolean;
  theme: Theme;
  _hasHydrated: boolean;

  // ===== Actions =====
  setSidebarWidth: (width: number) => void;
  toggleSidebarCollapsed: () => void;
  setSidebarHovered: (hovered: boolean) => void;
  resetSidebarWidth: () => void;
  setTheme: (theme: Theme) => void;
  setHasHydrated: (hasHydrated: boolean) => void;
}

const initialState = {
  sidebarWidth: SIDEBAR_CONFIG.DEFAULT_WIDTH,
  isSidebarCollapsed: false,
  isSidebarHovered: false,
  theme: 'system' as Theme,
  _hasHydrated: false,
};

export const useUiStore = create<UiState>()(
  devtools(
    persist(
      (set) => ({
        // 初始状态
        ...initialState,

        // 设置侧边栏宽度
        setSidebarWidth: (width) => set({ sidebarWidth: width }),

        // 切换折叠状态
        toggleSidebarCollapsed: () =>
          set((state) => ({
            isSidebarCollapsed: !state.isSidebarCollapsed,
          })),

        // 设置悬停状态
        setSidebarHovered: (hovered) => set({ isSidebarHovered: hovered }),

        // 重置宽度为默认值
        resetSidebarWidth: () =>
          set({ sidebarWidth: SIDEBAR_CONFIG.DEFAULT_WIDTH }),

        // 设置主题（不应用到 DOM，应用逻辑在 Core 层）
        setTheme: (theme) => set({ theme }),

        // 设置恢复状态
        setHasHydrated: (hasHydrated) => set({ _hasHydrated: hasHydrated }),
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
          }
        },
      }
    ),
    {
      name: 'ui-store',
    }
  )
);
