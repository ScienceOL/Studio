import {
  ArrowsPointingOutIcon,
  MinusIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import {
  Activity,
  Bell,
  Camera,
  Cpu,
  Home,
  MessageSquare,
  Newspaper,
  Server,
  TrendingUp,
  Workflow,
} from 'lucide-react';
import { AnimatePresence, motion } from 'motion/react';
import React, { useCallback, useEffect, useState } from 'react';
import { Rnd } from 'react-rnd';

// --- Components ---
import CameraMonitor from '@/components/CameraMonitor';
import { FloatingDock } from '@/components/ui/floating-dock';
import { EnvironmentPage } from './environment';

// --- Sections ---
import { Xyzen } from '@sciol/xyzen';
import {
  ArticlesSection,
  type ArticleItem,
  type ArticlesPageProps,
} from './components/ArticlesSection';
import { type NewsProps } from './components/NewsItem';
import { NewsSection } from './components/NewsSection';
import { type NodeTemplateProps } from './components/NodeItem';
import { NodesSection } from './components/NodesSection';
import { type NotificationProps } from './components/NotificationItem';
import { NotificationsSection } from './components/NotificationsSection';
import { TrendingTopicsSection } from './components/TrendingTopicsSection';
import {
  UserActivitySection,
  type ActivityProps,
} from './components/UserActivitySection';
import {
  WorkflowsSection,
  type WorkflowPostProps,
} from './components/WorkflowsSection';

// --- Widget Wrappers (Same as before) ---
const ArticlesWidget = () => {
  const sampleArticles: ArticlesPageProps = {
    results: Array.from({ length: 6 }).map((_, i) => ({
      uuid: `article-${i + 1}`,
    })) as ArticleItem[],
  };
  return (
    <div className="h-full w-full overflow-auto">
      <ArticlesSection articles={sampleArticles} />
    </div>
  );
};

const NewsWidget = () => {
  const sampleNews: NewsProps[] = Array.from({ length: 6 }).map((_, i) => ({
    id: i + 1,
    title: `News Title ${i + 1}`,
    content: 'This is a brief description of the news content.',
    category: i % 2 === 0 ? 'Update' : 'Announcement',
    created_at: new Date().toISOString(),
    link: '/news',
  }));
  return (
    <div className="h-full w-full overflow-auto">
      <NewsSection news={sampleNews} />
    </div>
  );
};

const NodesWidget = () => {
  const sampleNodes: NodeTemplateProps[] = Array.from({ length: 8 }).map(
    (_, i) => ({
      name: `node-${i + 1}`,
      version: `v${1 + i}.0.0`,
      description: 'A sample node description.',
      data: { header: `Node Header ${i + 1}` },
      creator: { username: `user${i + 1}` },
      updated_at: new Date().toISOString(),
    })
  );
  return (
    <div className="h-full w-full overflow-auto">
      <NodesSection nodes={sampleNodes} />
    </div>
  );
};

const NotificationsWidget = () => {
  const sampleNotifications: NotificationProps[] = Array.from({
    length: 5,
  }).map((_, i) => ({
    id: i + 1,
    type: (
      ['lab', 'device', 'alert', 'default'] as NotificationProps['type'][]
    )[i % 4],
    title: `Notification ${i + 1}`,
    content: 'This is a notification message preview.',
    isRead: i % 3 === 0,
    timestamp: new Date().toISOString(),
    link: '/notifications',
  }));
  return (
    <div className="h-full w-full overflow-auto">
      <NotificationsSection
        notifications={sampleNotifications}
        isAuthenticated={true}
      />
    </div>
  );
};

const TrendingTopicsWidget = () => {
  const topics = [
    { name: 'AI', count: 120 },
    { name: 'Robotics', count: 80 },
    { name: 'Cloud', count: 65 },
    { name: 'Data', count: 50 },
  ];
  return (
    <div className="h-full w-full overflow-auto">
      <TrendingTopicsSection topics={topics} />
    </div>
  );
};

const UserActivityWidget = () => {
  const activities: ActivityProps[] = Array.from({ length: 5 }).map((_, i) => ({
    id: i + 1,
    type: (['workflow', 'article', 'node', 'fork'] as ActivityProps['type'][])[
      i % 4
    ],
    title: `Activity ${i + 1}`,
    description: 'You performed an action recently.',
    link: '/activity',
    created_at: new Date().toISOString(),
  }));
  return (
    <div className="h-full w-full overflow-auto">
      <UserActivitySection activities={activities} isAuthenticated={true} />
    </div>
  );
};

const WorkflowsWidget = () => {
  const workflows: WorkflowPostProps[] = Array.from({ length: 3 }).map(
    (_, i) => ({
      workflow: { uuid: `workflow-${i + 1}` },
    })
  );
  return (
    <div className="h-full w-full overflow-auto">
      <WorkflowsSection workflows={workflows} />
    </div>
  );
};

const CameraWidget = () => {
  return (
    <div className="h-full w-full overflow-hidden">
      <CameraMonitor hostId="demo-host" cameraId="camera-1" />
    </div>
  );
};

interface AppDefinition {
  id: string;
  title: string;
  icon: React.ReactNode;
  component: React.FC | null; // null for simple links like Home
  defaultWidth: number;
  defaultHeight: number;
  href?: string;
}

const APPS: AppDefinition[] = [
  {
    id: 'home',
    title: 'Home',
    icon: (
      <Home className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: null,
    defaultWidth: 0,
    defaultHeight: 0,
    href: '/dashboard',
  },
  {
    id: 'camera',
    title: 'Camera',
    icon: (
      <Camera className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: CameraWidget,
    defaultWidth: 520,
    defaultHeight: 360,
  },
  {
    id: 'environment',
    title: 'Environment',
    icon: (
      <Server className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: EnvironmentPage,
    defaultWidth: 1000,
    defaultHeight: 700,
  },
  {
    id: 'nodes',
    title: 'Nodes',
    icon: (
      <Cpu className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: NodesWidget,
    defaultWidth: 800,
    defaultHeight: 600,
  },
  {
    id: 'workflows',
    title: 'Workflows',
    icon: (
      <Workflow className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: WorkflowsWidget,
    defaultWidth: 900,
    defaultHeight: 600,
  },
  {
    id: 'notifications',
    title: 'Notifications',
    icon: (
      <Bell className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: NotificationsWidget,
    defaultWidth: 400,
    defaultHeight: 500,
  },
  {
    id: 'news',
    title: 'News',
    icon: (
      <Newspaper className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: NewsWidget,
    defaultWidth: 400,
    defaultHeight: 600,
  },
  {
    id: 'activity',
    title: 'Activity',
    icon: (
      <Activity className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: UserActivityWidget,
    defaultWidth: 400,
    defaultHeight: 500,
  },
  {
    id: 'trending',
    title: 'Trending',
    icon: (
      <TrendingUp className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: TrendingTopicsWidget,
    defaultWidth: 400,
    defaultHeight: 400,
  },
  {
    id: 'settings',
    title: 'Settings',
    icon: (
      <MessageSquare className="h-full w-full text-neutral-500 dark:text-neutral-300" />
    ),
    component: ArticlesWidget,
    defaultWidth: 600,
    defaultHeight: 500,
  },
];

interface DesktopWindow {
  id: string; // unique instance id
  appId: string;
  x: number;
  y: number;
  width: number;
  height: number;
  zIndex: number;
  isMinimized: boolean;
}

export default function DashboardDesktop() {
  const [windows, setWindows] = useState<DesktopWindow[]>([]);
  const [maxZIndex, setMaxZIndex] = useState(100);

  const bringToFront = useCallback((windowId: string) => {
    setMaxZIndex((prev) => {
      const next = prev + 1;
      setWindows((curr) =>
        curr.map((w) => (w.id === windowId ? { ...w, zIndex: next } : w))
      );
      return next;
    });
  }, []);

  const openApp = useCallback(
    (app: AppDefinition) => {
      if (!app.component) {
        // Handle pure links or special actions
        if (app.href) window.location.href = app.href;
        return;
      }

      // Single instance behavior: check if already open
      const existing = windows.find((w) => w.appId === app.id);
      if (existing) {
        // If minimized, restore it
        if (existing.isMinimized) {
          setWindows((prev) =>
            prev.map((w) =>
              w.id === existing.id ? { ...w, isMinimized: false } : w
            )
          );
        }
        // Bring to front
        bringToFront(existing.id);
      } else {
        // Open new window
        const newZ = maxZIndex + 1;
        setMaxZIndex(newZ);
        const newWindow: DesktopWindow = {
          id: `${app.id}-${Date.now()}`,
          appId: app.id,
          x: 100 + ((windows.length * 20) % 200),
          y: 50 + ((windows.length * 20) % 200),
          width: app.defaultWidth,
          height: app.defaultHeight,
          zIndex: newZ,
          isMinimized: false,
        };
        setWindows((prev) => [...prev, newWindow]);
      }
    },
    [windows, maxZIndex, bringToFront]
  );

  // const openAppById = useCallback(
  //   (appId: string) => {
  //     const app = APPS.find((a) => a.id === appId);
  //     if (app) openApp(app);
  //   },
  //   [openApp]
  // );

  const closeWindow = useCallback((id: string) => {
    setWindows((prev) => prev.filter((w) => w.id !== id));
  }, []);

  const minimizeWindow = useCallback((id: string) => {
    setWindows((prev) =>
      prev.map((w) => (w.id === id ? { ...w, isMinimized: true } : w))
    );
  }, []);

  const updateWindow = useCallback(
    (id: string, data: Partial<DesktopWindow>) => {
      setWindows((prev) =>
        prev.map((w) => (w.id === id ? { ...w, ...data } : w))
      );
    },
    []
  );

  // Keyboard shortcut for minimizing active window
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Check for Command+Shift+M (Mac) or Ctrl+Shift+M (Windows/Linux)
      if (
        (e.metaKey || e.ctrlKey) &&
        e.shiftKey &&
        e.key.toLowerCase() === 'm'
      ) {
        e.preventDefault();

        // Find active window (highest zIndex and not minimized)
        const activeWindow = windows
          .filter((w) => !w.isMinimized)
          .sort((a, b) => b.zIndex - a.zIndex)[0];

        if (activeWindow) {
          minimizeWindow(activeWindow.id);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [windows, minimizeWindow]);

  // Dock items
  const dockItems = APPS.map((app) => ({
    title: app.title,
    icon: app.icon,
    href: app.href || '#',
    onClick: () => openApp(app),
    isActive: windows.some((w) => w.appId === app.id),
  }));

  return (
    <div
      className="relative h-full w-full overflow-hidden bg-neutral-100 dark:bg-neutral-950 font-sans"
      style={{
        backgroundImage:
          'url("https://images.unsplash.com/photo-1477346611705-65d1883cee1e?auto=format&fit=crop&q=80&w=2070&ixlib=rb-4.0.3")',
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }}
    >
      {/* Desktop Area */}
      {/* <div className="absolute inset-0 z-0">
        <button
          type="button"
          onClick={() => openAppById('camera')}
          className="absolute top-8 left-8 w-80 opacity-90 hover:opacity-100 transition-opacity text-left"
        >
          <div className="rounded-xl overflow-hidden shadow-2xl border border-white/20 backdrop-blur-md">
            <div className="bg-black/40 p-2 text-white text-xs font-medium">
              Camera Feed
            </div>
            <CameraMonitor hostId="demo-host" cameraId="camera-1" />
          </div>
        </button>
      </div> */}

      {/* Windows Layer */}
      <AnimatePresence>
        {windows.map((win) => {
          if (win.isMinimized) return null;

          const app = APPS.find((a) => a.id === win.appId);
          if (!app || !app.component) return null;
          const Component = app.component;

          return (
            <motion.div
              key={win.id}
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{
                opacity: 0,
                scale: 0,
                y: window.innerHeight - win.y,
                x: window.innerWidth / 2 - (win.x + win.width / 2),
                transition: { duration: 0.3, ease: 'easeInOut' },
              }}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: 0,
                height: 0,
                zIndex: win.zIndex,
              }}
            >
              <Rnd
                size={{ width: win.width, height: win.height }}
                position={{ x: win.x, y: win.y }}
                onMouseDown={() => bringToFront(win.id)}
                onDragStart={() => bringToFront(win.id)}
                onDragStop={(_e, d) => updateWindow(win.id, { x: d.x, y: d.y })}
                onResizeStart={() => bringToFront(win.id)}
                onResizeStop={(_e, _dir, ref, _d, pos) => {
                  updateWindow(win.id, {
                    width: parseInt(ref.style.width),
                    height: parseInt(ref.style.height),
                    ...pos,
                  });
                }}
                minWidth={300}
                minHeight={200}
                bounds="window"
                dragHandleClassName="window-header"
                enableUserSelectHack={false}
                className="bg-transparent"
              >
                <div className="w-full h-full flex flex-col rounded-xl overflow-hidden shadow-2xl border border-black/5 dark:border-white/10 bg-white/85 dark:bg-neutral-900/85 backdrop-blur-xl transition-shadow">
                  {/* MacOS-like Window Header */}
                  <div
                    className="window-header flex h-10 shrink-0 items-center justify-between px-4 bg-white/50 dark:bg-black/20 select-none border-b border-black/5 dark:border-white/5"
                    onDoubleClick={() => {
                      // Maximize logic placeholder
                    }}
                  >
                    {/* Traffic Lights */}
                    <div className="flex gap-2 group">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          closeWindow(win.id);
                        }}
                        className="w-3 h-3 rounded-full bg-red-500 hover:bg-red-600 flex items-center justify-center text-[8px] text-black/50 opacity-100 shadow-sm"
                      >
                        <XMarkIcon className="w-2 h-2 hidden group-hover:block text-red-900" />
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          minimizeWindow(win.id);
                        }}
                        className="w-3 h-3 rounded-full bg-yellow-500 hover:bg-yellow-600 flex items-center justify-center text-[8px] text-black/50 shadow-sm"
                      >
                        <MinusIcon className="w-2 h-2 hidden group-hover:block text-yellow-900" />
                      </button>
                      <button className="w-3 h-3 rounded-full bg-green-500 hover:bg-green-600 flex items-center justify-center text-[8px] text-black/50 shadow-sm">
                        <ArrowsPointingOutIcon className="w-2 h-2 hidden group-hover:block text-green-900" />
                      </button>
                    </div>

                    {/* Title */}
                    {/*<div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-sm font-medium text-neutral-700 dark:text-neutral-200">
                {app.title}
              </div>*/}
                  </div>

                  {/* Window Content */}
                  <div className="flex-1 custom-scrollbar min-h-0 overflow-hidden relative bg-white/50 dark:bg-neutral-900/50 p-0 flex flex-col">
                    <Component />
                  </div>
                </div>
              </Rnd>
            </motion.div>
          );
        })}
      </AnimatePresence>

      {/* Xyzen Side Panel (Global) */}
      <div className="relative">
        <Xyzen
          backendUrl={
            import.meta.env.DEV ? 'http://localhost:48196' : undefined
          }
          centeredInputPosition="bottom-right"
        />
      </div>

      {/* Dock Layer */}
      <div className="absolute bottom-4 left-1/2 -translate-x-1/2 z-[99999]">
        <FloatingDock items={dockItems} />
      </div>
    </div>
  );
}
