import { PlusIcon, XMarkIcon } from '@heroicons/react/24/outline';
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

// --- Context Menu ---
const ContextMenu = ({
  isOpen,
  x,
  y,
  onClose,
  onAddWidget,
}: {
  isOpen: boolean;
  x: number;
  y: number;
  onClose: () => void;
  onAddWidget: (type: WidgetType) => void;
}) => {
  useEffect(() => {
    const handleClickOutside = () => {
      onClose();
    };
    // Close on next click or context menu open
    if (isOpen) {
      window.addEventListener('click', handleClickOutside, { once: true });
      window.addEventListener('contextmenu', handleClickOutside, {
        once: true,
      });
    }
    return () => {
      window.removeEventListener('click', handleClickOutside);
      window.removeEventListener('contextmenu', handleClickOutside);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div
      style={{ top: y, left: x }}
      className="fixed z-[100] w-56 rounded-lg bg-white p-2 shadow-xl ring-1 ring-black ring-opacity-5 dark:bg-neutral-800 dark:ring-neutral-700"
    >
      <div className="py-1">
        <div className="px-3 py-1 text-sm font-semibold text-neutral-700 dark:text-neutral-300">
          Add Widget
        </div>
        <div className="my-1 border-t border-neutral-200 dark:border-neutral-700" />
        {Object.entries(WIDGET_DEFINITIONS).map(([type, def]) => (
          <button
            key={type}
            onClick={() => onAddWidget(type as WidgetType)}
            className="block w-full rounded-md px-3 py-1.5 text-left text-sm text-neutral-900 hover:bg-neutral-100 dark:text-neutral-200 dark:hover:bg-neutral-700"
          >
            {def.name}
          </button>
        ))}
      </div>
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
  const [contextMenu, setContextMenu] = useState({ isOpen: false, x: 0, y: 0 });

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

  const handleContextMenu = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    // Prevent context menu on widgets themselves
    if ((e.target as HTMLElement).closest('.react-grid-item')) {
      return;
    }
    setContextMenu({ isOpen: true, x: e.clientX, y: e.clientY });
  }, []);

  const closeContextMenu = useCallback(() => {
    setContextMenu((prev) => ({ ...prev, isOpen: false }));
  }, []);

  const handleAddWidgetFromContextMenu = useCallback(
    (type: WidgetType) => {
      addWidget(type);
      closeContextMenu();
    },
    [addWidget, closeContextMenu]
  );

  return (
    <div
      ref={gridContainerRef}
      className="h-full w-full bg-neutral-50 p-4 dark:bg-neutral-900"
      onContextMenu={handleContextMenu}
    >
      <AddWidgetModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onAddWidget={addWidget}
      />

      <ContextMenu
        isOpen={contextMenu.isOpen}
        x={contextMenu.x}
        y={contextMenu.y}
        onClose={closeContextMenu}
        onAddWidget={handleAddWidgetFromContextMenu}
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
          >
            {widgets.map((widget) => {
              const WidgetComponent = WIDGET_DEFINITIONS[widget.type].component;
              return (
                <div
                  key={widget.id}
                  className="group relative overflow-hidden rounded-lg bg-white shadow-md dark:bg-neutral-800"
                >
                  <WidgetComponent />
                  <button
                    onClick={() => removeWidget(widget.id)}
                    className="absolute right-2 top-2 z-10 flex h-6 w-6 items-center justify-center rounded-full bg-black/20 p-1 text-white opacity-0 transition-opacity hover:bg-red-500 group-hover:opacity-100"
                  >
                    <XMarkIcon className="h-4 w-4" />
                  </button>
                </div>
              );
            })}
          </GridLayout>
          <button
            onClick={() => setIsModalOpen(true)}
            className="fixed bottom-10 right-10 z-40 rounded-full bg-indigo-600 p-4 text-white shadow-lg transition-colors hover:bg-indigo-700"
          >
            <PlusIcon className="h-8 w-8" />
          </button>
        </>
      )}
    </div>
  );
}
