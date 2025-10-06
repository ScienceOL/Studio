import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

// 侧边栏宽度配置
const SIDEBAR_MIN_WIDTH = 64; // 最小宽度 (折叠状态)
const SIDEBAR_MAX_WIDTH = 320; // 最大宽度
const SIDEBAR_DEFAULT_WIDTH = 256; // 默认宽度

export interface UiState {
  // 侧边栏状态
  sidebarWidth: number;
  isSidebarCollapsed: boolean;
  isSidebarHovered: boolean;

  // 动作
  setSidebarWidth: (width: number) => void;
  toggleSidebarCollapsed: () => void;
  setSidebarHovered: (hovered: boolean) => void;
  resetSidebarWidth: () => void;

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
      }),
      {
        name: 'ui-storage',
        // 只持久化宽度和折叠状态
        partialize: (state) => ({
          sidebarWidth: state.sidebarWidth,
          isSidebarCollapsed: state.isSidebarCollapsed,
        }),
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
