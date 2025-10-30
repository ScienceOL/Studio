/**
 * 🎯 UI Core - UI 相关业务逻辑
 *
 * 职责：
 * 1. 主题切换逻辑
 * 2. 侧边栏宽度计算
 * 3. DOM 操作（应用主题）
 */

import { SIDEBAR_CONFIG, useUiStore, type Theme } from '@/store/uiStore';

export class UiCore {
  /**
   * 初始化 UI（恢复主题）
   */
  static initialize(): void {
    const store = useUiStore.getState();
    this.applyTheme(store.theme);
  }

  /**
   * 应用主题到 DOM
   * 业务逻辑：根据主题设置修改 DOM class
   */
  static applyTheme(selectedTheme?: Theme): void {
    const store = useUiStore.getState();
    const theme = selectedTheme ?? store.theme;
    const root = window.document.documentElement;

    // 判断是否应该应用暗色主题
    const isDark =
      theme === 'dark' ||
      (theme === 'system' &&
        window.matchMedia('(prefers-color-scheme: dark)').matches);

    // 注入 CSS 禁用过渡（避免闪烁）
    const style = document.createElement('style');
    style.innerHTML = '*, *::before, *::after { transition: none !important; }';
    document.head.appendChild(style);

    // 应用主题
    root.classList.toggle('dark', isDark);

    // 恢复过渡效果
    setTimeout(() => {
      if (document.head.contains(style)) {
        document.head.removeChild(style);
      }
    }, 50);
  }

  /**
   * 设置主题
   * 业务流程：更新状态 → 应用到 DOM
   */
  static setTheme(theme: Theme): void {
    const store = useUiStore.getState();
    store.setTheme(theme);
    this.applyTheme(theme);
  }

  /**
   * 循环切换主题
   * 业务逻辑：light → dark → system → light
   */
  static cycleTheme(): void {
    const store = useUiStore.getState();
    const themes: Theme[] = ['light', 'dark', 'system'];
    const currentTheme = store.theme;
    const currentIndex = themes.indexOf(currentTheme);
    const nextTheme = themes[(currentIndex + 1) % themes.length];

    this.setTheme(nextTheme);
  }

  /**
   * 获取有效的侧边栏宽度
   * 业务逻辑：考虑折叠和悬停状态
   */
  static getEffectiveSidebarWidth(): number {
    const store = useUiStore.getState();
    const { isSidebarCollapsed, isSidebarHovered, sidebarWidth } = store;

    // 折叠且未悬停时，使用最小宽度
    if (isSidebarCollapsed && !isSidebarHovered) {
      return SIDEBAR_CONFIG.MIN_WIDTH;
    }

    return sidebarWidth;
  }

  /**
   * 设置侧边栏宽度（带约束）
   * 业务逻辑：限制在最小和最大宽度之间
   */
  static setSidebarWidth(width: number): void {
    const store = useUiStore.getState();

    const constrainedWidth = Math.max(
      SIDEBAR_CONFIG.MIN_WIDTH,
      Math.min(SIDEBAR_CONFIG.MAX_WIDTH, width)
    );

    store.setSidebarWidth(constrainedWidth);
  }

  /**
   * 切换侧边栏折叠状态
   */
  static toggleSidebarCollapsed(): void {
    const store = useUiStore.getState();
    store.toggleSidebarCollapsed();
  }

  /**
   * 设置侧边栏悬停状态
   */
  static setSidebarHovered(hovered: boolean): void {
    const store = useUiStore.getState();
    store.setSidebarHovered(hovered);
  }

  /**
   * 重置侧边栏宽度为默认值
   */
  static resetSidebarWidth(): void {
    const store = useUiStore.getState();
    store.resetSidebarWidth();
  }
}
