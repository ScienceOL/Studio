import {
  EllipsisVerticalIcon,
  PlusIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import { useCallback, useEffect, useRef, useState } from 'react';
import GridLayout, { type Layout } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';

// --- Real Dashboard Sections ---
import AddWidgetModal from './AddWidgetModal';
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

// --- Wrapper Widgets with sample data ---
const ArticlesWidget = () => {
  const sampleArticles: ArticlesPageProps = {
    results: Array.from({ length: 6 }).map((_, i) => ({
      uuid: `article-${i + 1}`,
    })) as ArticleItem[],
  };
  return <ArticlesSection articles={sampleArticles} />;
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
  return <NewsSection news={sampleNews} />;
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
  return <NodesSection nodes={sampleNodes} />;
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
    <NotificationsSection
      notifications={sampleNotifications}
      isAuthenticated={true}
    />
  );
};

const TrendingTopicsWidget = () => {
  const topics = [
    { name: 'AI', count: 120 },
    { name: 'Robotics', count: 80 },
    { name: 'Cloud', count: 65 },
    { name: 'Data', count: 50 },
  ];
  return <TrendingTopicsSection topics={topics} />;
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
  return <UserActivitySection activities={activities} isAuthenticated={true} />;
};

const WorkflowsWidget = () => {
  const workflows: WorkflowPostProps[] = Array.from({ length: 3 }).map(
    (_, i) => ({
      workflow: { uuid: `workflow-${i + 1}` },
    })
  );
  return <WorkflowsSection workflows={workflows} />;
};

// --- Widget Definitions ---
type WidgetSize = 'small' | 'medium' | 'large';

const SIZE_PRESET: Record<WidgetSize, { w: number; h: number }> = {
  small: { w: 2, h: 2 },
  medium: { w: 4, h: 2 },
  large: { w: 6, h: 4 },
};
type WidgetCategory =
  | 'Node'
  | 'Workflow'
  | 'Chat'
  | 'Environment'
  | 'UserActivity'
  | 'Notification'
  | 'Trending'
  | 'Others';

const WIDGET_CATEGORIES: { key: WidgetCategory; label: string }[] = [
  { key: 'Node', label: 'Node' },
  { key: 'Workflow', label: 'Workflow' },
  { key: 'Chat', label: 'Chat' },
  { key: 'Environment', label: 'Environment' },
  { key: 'UserActivity', label: 'UserActivity' },
  { key: 'Notification', label: 'Notification' },
  { key: 'Trending', label: 'Trending' },
  { key: 'Others', label: 'Others' },
];

const WIDGET_DEFINITIONS = {
  articles: {
    component: ArticlesWidget,
    name: 'Articles',
    size: 'medium' as WidgetSize,
    category: 'Others' as WidgetCategory,
  },
  news: {
    component: NewsWidget,
    name: 'News',
    size: 'small' as WidgetSize,
    category: 'Others' as WidgetCategory,
  },
  nodes: {
    component: NodesWidget,
    name: 'Nodes',
    size: 'medium' as WidgetSize,
    category: 'Node' as WidgetCategory,
  },
  notifications: {
    component: NotificationsWidget,
    name: 'Notifications',
    size: 'small' as WidgetSize,
    category: 'Notification' as WidgetCategory,
  },
  'trending-topics': {
    component: TrendingTopicsWidget,
    name: 'Trending Topics',
    size: 'small' as WidgetSize,
    category: 'Trending' as WidgetCategory,
  },
  'user-activity': {
    component: UserActivityWidget,
    name: 'User Activity',
    size: 'small' as WidgetSize,
    category: 'UserActivity' as WidgetCategory,
  },
  workflows: {
    component: WorkflowsWidget,
    name: 'Workflows',
    size: 'large' as WidgetSize,
    category: 'Workflow' as WidgetCategory,
  },
};

type WidgetType = keyof typeof WIDGET_DEFINITIONS;

interface Widget {
  id: string;
  type: WidgetType;
}

// (Modal extracted to standalone component)

// --- Widget Toolbar ---
const WidgetToolbar = ({ onRemove }: { onRemove: () => void }) => {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    const handleClickOutside = () => {
      setIsOpen(false);
    };
    if (isOpen) {
      window.addEventListener('click', handleClickOutside, { once: true });
    }
    return () => {
      window.removeEventListener('click', handleClickOutside);
    };
  }, [isOpen]);

  return (
    <div
      className="absolute right-2 top-2 z-10 no-drag"
      data-no-drag="true"
      style={{ pointerEvents: 'auto' }}
    >
      <button
        onClick={(e) => {
          e.stopPropagation();
          e.preventDefault();
          setIsOpen(!isOpen);
        }}
        onMouseDown={(e) => {
          e.stopPropagation();
        }}
        className="flex h-6 w-6 items-center justify-center rounded-full bg-black/20 text-white opacity-0 transition-opacity hover:bg-black/40 group-hover:opacity-100"
        style={{ pointerEvents: 'auto' }}
      >
        <EllipsisVerticalIcon className="h-4 w-4" />
      </button>

      {isOpen && (
        <div
          className="absolute right-0 top-8 w-32 rounded-lg bg-white p-1 shadow-lg ring-1 ring-black ring-opacity-5 dark:bg-neutral-800 dark:ring-neutral-700"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={(e) => {
              e.stopPropagation();
              onRemove();
              setIsOpen(false);
            }}
            className="flex w-full items-center rounded-md px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
          >
            <XMarkIcon className="mr-2 h-4 w-4" />
            Remove
          </button>
        </div>
      )}
    </div>
  );
};

// --- Main Dashboard Component ---
export default function DashboardHome() {
  const [widgets, setWidgets] = useState<Widget[]>([]);
  const [layout, setLayout] = useState<Layout[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const gridContainerRef = useRef<HTMLDivElement>(null);
  const [gridWidth, setGridWidth] = useState(1200);

  const COL_WIDTH = 120;
  const ROW_HEIGHT = 120;

  useEffect(() => {
    const observer = new ResizeObserver((entries) => {
      if (entries[0]) {
        setGridWidth(entries[0].contentRect.width);
      }
    });

    const currentRef = gridContainerRef.current;
    if (currentRef) {
      observer.observe(currentRef);
    }

    return () => {
      if (currentRef) {
        observer.unobserve(currentRef);
      }
    };
  }, []);

  const cols = Math.max(1, Math.floor(gridWidth / COL_WIDTH));

  const addWidget = useCallback(
    (type: WidgetType) => {
      const newWidgetDef = WIDGET_DEFINITIONS[type];
      if (!newWidgetDef) return;

      const newWidget: Widget = {
        id: `${type}-${Date.now()}`,
        type: type,
      };

      const preset = SIZE_PRESET[newWidgetDef.size];

      const newLayoutItem: Layout = {
        i: newWidget.id,
        x: (layout.length * preset.w) % cols,
        y: Infinity, // Places it at the bottom
        w: preset.w,
        h: preset.h,
        minW: preset.w,
        maxW: preset.w,
        minH: preset.h,
        maxH: preset.h,
      };

      setWidgets((prev) => [...prev, newWidget]);
      setLayout((prev) => [...prev, newLayoutItem]);
      setIsModalOpen(false);
    },
    [layout, cols]
  );

  const removeWidget = useCallback((widgetId: string) => {
    setWidgets((prev) => prev.filter((w) => w.id !== widgetId));
    setLayout((prev) => prev.filter((l) => l.i !== widgetId));
  }, []);

  const onLayoutChange = (newLayout: Layout[]) => {
    setLayout(newLayout);
  };

  return (
    <div
      ref={gridContainerRef}
      className="relative h-full w-full bg-neutral-50 p-4 dark:bg-neutral-900"
    >
      {/* 右上角添加组件按钮 */}
      <button
        onClick={() => setIsModalOpen(true)}
        className="absolute right-4 top-4 z-50 flex h-10 w-10 items-center justify-center rounded-full bg-indigo-600 text-white shadow-lg transition-colors hover:bg-indigo-700"
        title="Add Widget"
      >
        <PlusIcon className="h-5 w-5" />
      </button>

      <AddWidgetModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onAddWidget={(type) => addWidget(type as WidgetType)}
        categories={WIDGET_CATEGORIES}
        items={Object.entries(WIDGET_DEFINITIONS).map(([type, def]) => {
          const PreviewComp = def.component;
          const preset = SIZE_PRESET[def.size];
          return {
            type,
            name: def.name,
            size: def.size,
            w: preset.w,
            h: preset.h,
            category: def.category,
            preview: <PreviewComp />,
          };
        })}
      />

      {widgets.length === 0 ? (
        <div className="flex h-full items-center justify-center">
          <button
            onClick={() => setIsModalOpen(true)}
            className="flex h-40 w-64 flex-col items-center justify-center rounded-lg border-2 border-dashed border-neutral-400 transition-colors hover:bg-neutral-100 dark:hover:bg-neutral-800"
          >
            <PlusIcon className="h-12 w-12 text-neutral-500" />
            <span className="mt-2 text-neutral-600 dark:text-neutral-400">
              Add New Widget
            </span>
          </button>
        </div>
      ) : (
        <>
          <GridLayout
            className="layout [&_.react-grid-placeholder]:rounded-lg [&_.react-grid-placeholder]:bg-indigo-500/15 dark:[&_.react-grid-placeholder]:bg-indigo-500/25"
            layout={layout}
            cols={cols}
            rowHeight={ROW_HEIGHT}
            width={gridWidth}
            onLayoutChange={onLayoutChange}
            margin={[16, 16]}
            containerPadding={[0, 0]}
            draggableCancel=".no-drag"
            isResizable={false}
          >
            {widgets.map((widget) => {
              const WidgetComponent = WIDGET_DEFINITIONS[widget.type].component;
              return (
                <div
                  key={widget.id}
                  className="group relative overflow-hidden rounded-lg bg-white shadow-md ring-1 ring-black/5 dark:bg-neutral-800 dark:ring-white/10"
                >
                  <WidgetComponent />
                  <WidgetToolbar onRemove={() => removeWidget(widget.id)} />
                </div>
              );
            })}
          </GridLayout>
        </>
      )}
    </div>
  );
}
