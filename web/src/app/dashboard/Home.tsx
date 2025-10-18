import {
  EllipsisVerticalIcon,
  PlusIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import { useCallback, useEffect, useRef, useState } from 'react';
import GridLayout, { type Layout } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';

// --- Static Widget Components ---
const ArticlesWidget = () => (
  <div className="flex h-full w-full items-center justify-center rounded-lg bg-blue-100 p-4 text-lg font-bold text-blue-800 dark:bg-blue-900/50 dark:text-blue-200">
    Articles
  </div>
);
const NewsWidget = () => (
  <div className="flex h-full w-full items-center justify-center rounded-lg bg-green-100 p-4 text-lg font-bold text-green-800 dark:bg-green-900/50 dark:text-green-200">
    News
  </div>
);
const NodesWidget = () => (
  <div className="flex h-full w-full items-center justify-center rounded-lg bg-red-100 p-4 text-lg font-bold text-red-800 dark:bg-red-900/50 dark:text-red-200">
    Nodes
  </div>
);
const NotificationsWidget = () => (
  <div className="flex h-full w-full items-center justify-center rounded-lg bg-yellow-100 p-4 text-lg font-bold text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-200">
    Notifications
  </div>
);
const TrendingTopicsWidget = () => (
  <div className="flex h-full w-full items-center justify-center rounded-lg bg-purple-100 p-4 text-lg font-bold text-purple-800 dark:bg-purple-900/50 dark:text-purple-200">
    Trending Topics
  </div>
);
const UserActivityWidget = () => (
  <div className="flex h-full w-full items-center justify-center rounded-lg bg-indigo-100 p-4 text-lg font-bold text-indigo-800 dark:bg-indigo-900/50 dark:text-indigo-200">
    User Activity
  </div>
);
const WorkflowsWidget = () => (
  <div className="flex h-full w-full items-center justify-center rounded-lg bg-pink-100 p-4 text-lg font-bold text-pink-800 dark:bg-pink-900/50 dark:text-pink-200">
    Workflows
  </div>
);

// --- Widget Definitions ---
const WIDGET_DEFINITIONS = {
  articles: {
    w: 4,
    h: 2,
    component: ArticlesWidget,
    name: 'Articles',
    size: 'medium',
  },
  news: { w: 2, h: 2, component: NewsWidget, name: 'News', size: 'small' },
  nodes: { w: 4, h: 2, component: NodesWidget, name: 'Nodes', size: 'medium' },
  notifications: {
    w: 2,
    h: 2,
    component: NotificationsWidget,
    name: 'Notifications',
    size: 'small',
  },
  'trending-topics': {
    w: 2,
    h: 2,
    component: TrendingTopicsWidget,
    name: 'Trending Topics',
    size: 'small',
  },
  'user-activity': {
    w: 2,
    h: 2,
    component: UserActivityWidget,
    name: 'User Activity',
    size: 'small',
  },
  workflows: {
    w: 4,
    h: 4,
    component: WorkflowsWidget,
    name: 'Workflows',
    size: 'large',
  },
};

type WidgetType = keyof typeof WIDGET_DEFINITIONS;

interface Widget {
  id: string;
  type: WidgetType;
}

// --- Add Widget Modal ---
const AddWidgetModal = ({
  isOpen,
  onClose,
  onAddWidget,
}: {
  isOpen: boolean;
  onClose: () => void;
  onAddWidget: (type: WidgetType) => void;
}) => {
  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="flex h-[80vh] w-full max-w-4xl flex-col rounded-2xl bg-white shadow-xl dark:bg-neutral-800"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex shrink-0 items-center justify-between border-b p-4 dark:border-neutral-700">
          <h2 className="text-xl font-bold">Add Widget</h2>
          <button
            onClick={onClose}
            className="rounded-full p-1 hover:bg-neutral-200 dark:hover:bg-neutral-700"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </header>
        <main className="flex-grow overflow-y-auto p-6">
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {Object.entries(WIDGET_DEFINITIONS).map(([type, def]) => (
              <div
                key={type}
                className="flex cursor-pointer flex-col items-center justify-center rounded-lg border p-4 transition-colors hover:bg-neutral-100 dark:border-neutral-700 dark:hover:bg-neutral-700/50"
                onClick={() => onAddWidget(type as WidgetType)}
              >
                <div className="mb-2 text-lg font-semibold">{def.name}</div>
                <div className="capitalize text-neutral-500">
                  {def.size} ({def.w}x{def.h})
                </div>
              </div>
            ))}
          </div>
        </main>
      </div>
    </div>
  );
};

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

      const newLayoutItem: Layout = {
        i: newWidget.id,
        x: (layout.length * newWidgetDef.w) % cols,
        y: Infinity, // Places it at the bottom
        w: newWidgetDef.w,
        h: newWidgetDef.h,
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
        onAddWidget={addWidget}
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
            className="layout"
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
                  className="group relative overflow-hidden rounded-lg bg-white shadow-md dark:bg-neutral-800"
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
