/**
 * 🎣 Hook Layer - UI 能力暴露
 *
 * 职责：
 * 1. 封装 Core 层的 UI 能力
 * 2. 提供响应式的状态订阅
 * 3. 暴露简洁的 API 给组件
 */

import { UiCore } from '@/core/uiCore';
import { SIDEBAR_CONFIG, useUiStore, type Theme } from '@/store/uiStore';
import { useCallback, useEffect } from 'react';

export interface UseUIReturn {
  // ===== 状态 =====
  theme: Theme;
  sidebarWidth: number;
  isSidebarCollapsed: boolean;
  isSidebarHovered: boolean;
  hasHydrated: boolean;

  // ===== 派生状态 =====
  effectiveSidebarWidth: number;

  // ===== 动作 =====
  setTheme: (theme: Theme) => void;
  cycleTheme: () => void;
  setSidebarWidth: (width: number) => void;
  toggleSidebarCollapsed: () => void;
  setSidebarHovered: (hovered: boolean) => void;
  resetSidebarWidth: () => void;
}

/**
 * UI Hook
 * 组件通过这个 Hook 访问所有 UI 相关功能
 */
export function useUI(): UseUIReturn {
  // 订阅 Store 状态
  const theme = useUiStore((state) => state.theme);
  const sidebarWidth = useUiStore((state) => state.sidebarWidth);
  const isSidebarCollapsed = useUiStore((state) => state.isSidebarCollapsed);
  const isSidebarHovered = useUiStore((state) => state.isSidebarHovered);
  const hasHydrated = useUiStore((state) => state._hasHydrated);

  // 初始化（应用主题）
  useEffect(() => {
    if (hasHydrated) {
      UiCore.initialize();
    }
  }, [hasHydrated]);

  // 封装 Core 层方法
  const setTheme = useCallback((theme: Theme) => {
    UiCore.setTheme(theme);
  }, []);

  const cycleTheme = useCallback(() => {
    UiCore.cycleTheme();
  }, []);

  const setSidebarWidth = useCallback((width: number) => {
    UiCore.setSidebarWidth(width);
  }, []);

  const toggleSidebarCollapsed = useCallback(() => {
    UiCore.toggleSidebarCollapsed();
  }, []);

  const setSidebarHovered = useCallback((hovered: boolean) => {
    UiCore.setSidebarHovered(hovered);
  }, []);

  const resetSidebarWidth = useCallback(() => {
    UiCore.resetSidebarWidth();
  }, []);

  // 派生状态（调用 Core 层计算）
  const effectiveSidebarWidth = UiCore.getEffectiveSidebarWidth();

  return {
    // 状态
    theme,
    sidebarWidth,
    isSidebarCollapsed,
    isSidebarHovered,
    hasHydrated,

    // 派生状态
    effectiveSidebarWidth,

    // 动作
    setTheme,
    cycleTheme,
    setSidebarWidth,
    toggleSidebarCollapsed,
    setSidebarHovered,
    resetSidebarWidth,
  };
}

/**
 * 主题 Hook（简化版，只关注主题）
 */
export function useTheme() {
  const theme = useUiStore((state) => state.theme);
  const hasHydrated = useUiStore((state) => state._hasHydrated);

  useEffect(() => {
    if (hasHydrated) {
      UiCore.initialize();
    }
  }, [hasHydrated]);

  const setTheme = useCallback((theme: Theme) => {
    UiCore.setTheme(theme);
  }, []);

  const cycleTheme = useCallback(() => {
    UiCore.cycleTheme();
  }, []);

  return {
    theme,
    setTheme,
    cycleTheme,
  };
}

/**
 * 侧边栏 Hook（简化版，只关注侧边栏）
 */
export function useSidebar() {
  const sidebarWidth = useUiStore((state) => state.sidebarWidth);
  const isSidebarCollapsed = useUiStore((state) => state.isSidebarCollapsed);
  const isSidebarHovered = useUiStore((state) => state.isSidebarHovered);

  const setSidebarWidth = useCallback((width: number) => {
    UiCore.setSidebarWidth(width);
  }, []);

  const toggleSidebarCollapsed = useCallback(() => {
    UiCore.toggleSidebarCollapsed();
  }, []);

  const setSidebarHovered = useCallback((hovered: boolean) => {
    UiCore.setSidebarHovered(hovered);
  }, []);

  const resetSidebarWidth = useCallback(() => {
    UiCore.resetSidebarWidth();
  }, []);

  const effectiveSidebarWidth = UiCore.getEffectiveSidebarWidth();

  return {
    sidebarWidth,
    isSidebarCollapsed,
    isSidebarHovered,
    effectiveSidebarWidth,
    setSidebarWidth,
    toggleSidebarCollapsed,
    setSidebarHovered,
    resetSidebarWidth,
  };
}

// 导出配置常量
export { SIDEBAR_CONFIG };
