import Logo from '@/assets/Logo';
import { GitHubIconOutline } from '@/assets/SocialIcons';
import { DropdownMenu } from '@/components/layout/DropdownMenu';
import { useAuthStore } from '@/store/authStore';
import { SIDEBAR_CONFIG, useUiStore } from '@/store/uiStore';
import type { DragMoveEvent } from '@dnd-kit/core';
import {
  DndContext,
  PointerSensor,
  useDraggable,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import { restrictToHorizontalAxis } from '@dnd-kit/modifiers';
import { MenuButton } from '@headlessui/react';
import {
  ArrowTopRightOnSquareIcon,
  HomeIcon,
  Square3Stack3DIcon,
} from '@heroicons/react/24/outline';
import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';

function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(' ');
}

interface NavigationItem {
  name: string;
  href: string;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  current: boolean;
}

// 默认导航配置 - Dashboard Layout 下的所有路由
const defaultNavigation: NavigationItem[] = [
  {
    name: 'sidebar.dashboard',
    href: '/dashboard',
    icon: HomeIcon,
    current: true,
  },
  {
    name: 'sidebar.environment',
    href: '/dashboard/environment',
    icon: Square3Stack3DIcon,
    current: false,
  },
];

// 拖拽手柄组件
const DragHandle = ({
  isActive,
  onDoubleClick,
}: {
  isActive: boolean;
  onDoubleClick: (e: React.MouseEvent) => void;
}) => {
  const { attributes, listeners, setNodeRef } = useDraggable({
    id: 'sidebar-resizer',
  });

  return (
    <div
      ref={setNodeRef}
      className={`absolute right-0 top-0 z-50 h-full w-1 cursor-col-resize ${
        isActive
          ? 'bg-indigo-500 shadow-md dark:bg-indigo-400'
          : 'bg-transparent hover:bg-indigo-400/60 hover:shadow-sm dark:hover:bg-indigo-500/60'
      } transition-all duration-150 ease-in-out`}
      {...listeners}
      {...attributes}
      onDoubleClick={onDoubleClick}
    >
      {/* 扩大拖拽区域 */}
      <div className="absolute right-0 top-0 h-full w-4 -translate-x-1.5" />
    </div>
  );
};

export default function ResizableSidebar() {
  const { t } = useTranslation('translation');
  const location = useLocation();
  const sidebarRef = useRef<HTMLDivElement>(null);

  // UI Store
  const sidebarWidth = useUiStore((s) => s.sidebarWidth);
  const isSidebarCollapsed = useUiStore((s) => s.isSidebarCollapsed);
  const setSidebarWidth = useUiStore((s) => s.setSidebarWidth);
  const setSidebarHovered = useUiStore((s) => s.setSidebarHovered);

  // Local state
  const [navigation, setNavigation] =
    useState<NavigationItem[]>(defaultNavigation);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const lastWidthRef = useRef(sidebarWidth);

  // User data from auth store
  const user = useAuthStore((s) => s.user);

  // 配置 dnd-kit sensor
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5 },
    })
  );

  // Update navigation current state based on route
  useEffect(() => {
    setNavigation((prevNavigation) =>
      prevNavigation.map((navItem: NavigationItem) => {
        const itemHref = navItem.href;
        const currentPath = location.pathname;

        let isCurrent = false;

        if (itemHref === '/dashboard') {
          isCurrent = currentPath === '/dashboard';
        } else {
          isCurrent =
            currentPath === itemHref || currentPath.startsWith(`${itemHref}/`);
        }

        return {
          ...navItem,
          current: isCurrent,
        };
      })
    );
  }, [location.pathname]);

  // 拖拽处理函数
  const handleDragStart = () => {
    setIsDragging(true);
    lastWidthRef.current = sidebarWidth;
  };

  const handleDragMove = (event: DragMoveEvent) => {
    const newWidth = Math.min(
      Math.max(lastWidthRef.current + event.delta.x, SIDEBAR_CONFIG.MIN_WIDTH),
      SIDEBAR_CONFIG.MAX_WIDTH
    );
    setSidebarWidth(newWidth);
  };

  const handleDragEnd = () => {
    setIsDragging(false);
  };

  const handleResizeDoubleClick = () => {
    setSidebarWidth(SIDEBAR_CONFIG.DEFAULT_WIDTH);
    lastWidthRef.current = SIDEBAR_CONFIG.DEFAULT_WIDTH;
  };

  // 计算有效宽度
  const effectiveWidth =
    isSidebarCollapsed && !isDropdownOpen
      ? SIDEBAR_CONFIG.MIN_WIDTH
      : sidebarWidth;

  return (
    <DndContext
      sensors={sensors}
      onDragStart={handleDragStart}
      onDragMove={handleDragMove}
      onDragEnd={handleDragEnd}
      modifiers={[restrictToHorizontalAxis]}
    >
      <div
        ref={sidebarRef}
        className="group relative z-50 hidden bg-white transition-all duration-300 ease-in-out lg:fixed lg:inset-y-0 lg:flex lg:flex-col lg:overflow-hidden dark:bg-neutral-900"
        style={{
          width: `${effectiveWidth}px`,
          transition: isDragging ? 'none' : 'width 0.2s ease-in-out',
        }}
        onMouseEnter={() => setSidebarHovered(true)}
        onMouseLeave={() => setSidebarHovered(false)}
      >
        {/* Main content */}
        <div className="flex grow flex-col overflow-y-auto border-r border-neutral-200 bg-white px-3 pb-4 dark:border-neutral-700 dark:bg-neutral-900">
          {/* Header with user dropdown */}
          <div className="flex h-16 shrink-0 items-center justify-between overflow-hidden">
            <DropdownMenu
              avatar={user?.avatar || '/default_avatar.png'}
              username={user?.name || 'User'}
              onOpenChange={(open) => setIsDropdownOpen(open)}
            >
              <MenuButton className="flex h-10 w-10 fill-indigo-800 transition-opacity duration-150 ease-in-out hover:opacity-75 dark:fill-white">
                <img
                  className="h-9 w-9 rounded-full bg-neutral-50"
                  src={user?.avatar || '/default_avatar.png'}
                  alt="avatar"
                />
                {effectiveWidth > SIDEBAR_CONFIG.MIN_WIDTH && (
                  <span className="ml-4 flex items-center text-sm font-semibold leading-6 text-neutral-900 dark:text-white">
                    {user?.name || 'User'}
                  </span>
                )}
              </MenuButton>
            </DropdownMenu>
          </div>

          {/* Navigation */}
          <nav className="flex flex-1 flex-col overflow-hidden">
            <ul role="list" className="flex flex-1 flex-col gap-y-2">
              <li className="relative">
                <ul role="list" className="-mx-1 space-y-1">
                  {navigation.map((item: NavigationItem) => (
                    <li key={item.name}>
                      <Link
                        to={item.href}
                        target={item.href.startsWith('http') ? '_blank' : ''}
                        rel={
                          item.href.startsWith('http')
                            ? 'noopener noreferrer'
                            : undefined
                        }
                        className={classNames(
                          item.current
                            ? 'bg-neutral-50 text-indigo-600 dark:bg-neutral-800 dark:text-white'
                            : 'text-neutral-700 hover:bg-neutral-50 hover:text-indigo-600 dark:text-neutral-400 dark:hover:bg-neutral-800 dark:hover:text-white',
                          'group/link my-1 flex items-center gap-x-3 whitespace-nowrap rounded-xl px-2.5 py-2 text-sm font-semibold leading-6'
                        )}
                        onClick={() => {
                          const updatedNavigation = navigation.map(
                            (navItem: NavigationItem) => ({
                              ...navItem,
                              current: navItem.name === item.name,
                            })
                          );
                          setNavigation(updatedNavigation);
                        }}
                      >
                        <item.icon
                          className={classNames(
                            item.current
                              ? 'text-indigo-600 dark:text-inherit'
                              : 'text-neutral-400 group-hover/link:text-indigo-600 dark:text-inherit dark:group-hover/link:text-inherit',
                            'h-6 w-6 shrink-0'
                          )}
                          aria-hidden="true"
                        />
                        {effectiveWidth > SIDEBAR_CONFIG.MIN_WIDTH && (
                          <span className="transition-opacity duration-200">
                            {t(item.name)}
                          </span>
                        )}
                        {item.href.startsWith('http') &&
                          effectiveWidth > SIDEBAR_CONFIG.MIN_WIDTH && (
                            <ArrowTopRightOnSquareIcon className="ml-auto h-4 w-4 text-neutral-400 group-hover/link:text-indigo-600 dark:text-inherit dark:group-hover/link:text-inherit" />
                          )}
                      </Link>
                    </li>
                  ))}
                </ul>
              </li>

              {/* Footer links */}
              <li className="mt-auto flex flex-col">
                <Link
                  to="https://github.com/ScienceOL"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="group -mx-1 flex w-fit gap-x-3 whitespace-nowrap rounded px-3 py-2 text-xs font-semibold leading-6 text-neutral-700 hover:bg-neutral-50 hover:text-neutral-600 dark:text-neutral-600 dark:hover:bg-neutral-800 dark:hover:text-white"
                >
                  <GitHubIconOutline
                    className="h-6 w-6 shrink-0 text-neutral-400 group-hover:text-neutral-600 dark:text-neutral-600 dark:hover:text-white dark:group-hover:bg-inherit dark:group-hover:text-inherit"
                    aria-hidden="true"
                  />
                  {effectiveWidth > SIDEBAR_CONFIG.MIN_WIDTH && (
                    <span className="transition-opacity duration-200">
                      Github
                    </span>
                  )}
                </Link>
                <Link
                  to="/landscape"
                  className="group -mx-1 flex w-fit gap-x-3 whitespace-nowrap rounded px-3 py-2 text-xs font-semibold leading-6 text-neutral-700 hover:bg-neutral-50 hover:text-indigo-600 dark:text-neutral-600 dark:hover:bg-neutral-800 dark:hover:text-white"
                >
                  <Logo
                    className="h-6 w-6 shrink-0 text-neutral-400 group-hover:text-indigo-600 dark:hover:text-white dark:group-hover:bg-inherit dark:group-hover:text-inherit"
                    aria-hidden="true"
                  />
                  {effectiveWidth > SIDEBAR_CONFIG.MIN_WIDTH && (
                    <span className="transition-opacity duration-200">
                      ScienceOL
                    </span>
                  )}
                </Link>
              </li>
            </ul>
          </nav>
        </div>

        {/* Resize handle */}
        {!isSidebarCollapsed && (
          <DragHandle
            isActive={isDragging}
            onDoubleClick={handleResizeDoubleClick}
          />
        )}
      </div>
    </DndContext>
  );
}
