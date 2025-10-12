import { HeroPattern } from '@/components/basic/patterns/HeroPattern';
import ResizableSidebar from '@/components/layout/ResizableSidebar';
import { useUiStore } from '@/store/uiStore';
import { Xyzen, useXyzen } from '@sciol/xyzen';
import { Outlet, useLocation } from 'react-router-dom';

export default function DashboardLayout() {
  const { isXyzenOpen, panelWidth } = useXyzen();
  const location = useLocation();
  const sidebarWidth = useUiStore((s) => s.sidebarWidth);
  const isSidebarCollapsed = useUiStore((s) => s.isSidebarCollapsed);

  // 计算有效的侧边栏宽度
  const effectiveSidebarWidth = isSidebarCollapsed ? 64 : sidebarWidth;

  // 检查是否需要显示背景图案 - 特定页面显示
  const showHeroPattern = location.pathname.match(
    /^\/(space|manuscript|discussion)(\/.*)?$/
  );

  return (
    <div className="relative min-h-screen">
      {/* 可调整大小的侧边栏 */}
      <ResizableSidebar />

      {/* Xyzen 面板 */}
      <Xyzen
        showThemeToggle={false}
        backendUrl={import.meta.env.DEV ? 'http://localhost:48196' : undefined}
      />

      {/* 主内容区域 */}
      <main className="relative min-h-screen dark:bg-neutral-900">
        {/* 背景图案 */}
        {showHeroPattern && <HeroPattern />}

        {/* 内容容器 - 根据侧边栏和 Xyzen 面板调整 */}
        <div
          className="absolute bottom-0 top-0 transition-all duration-300 ease-in-out"
          style={{
            left: `${effectiveSidebarWidth}px`,
            right: isXyzenOpen ? `${panelWidth}px` : '0',
          }}
        >
          {/* 嵌套路由出口 */}
          <Outlet />
        </div>
      </main>
    </div>
  );
}
