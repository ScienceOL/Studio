import { XMarkIcon } from '@heroicons/react/24/outline';
import type { ReactNode } from 'react';
import { useState } from 'react';

export type WidgetCategoryKey = string;

export interface WidgetCategoryDef {
  key: WidgetCategoryKey;
  label: string;
}

export interface WidgetCatalogItem<TCategory extends string = string> {
  type: string;
  name: string;
  size: string;
  w: number;
  h: number;
  category: TCategory;
  preview?: ReactNode | (() => ReactNode);
}

export interface AddWidgetModalProps<TCategory extends string = string> {
  isOpen: boolean;
  onClose: () => void;
  onAddWidget: (type: string) => void;
  categories: WidgetCategoryDef[];
  items: WidgetCatalogItem<TCategory>[];
  initialCategoryKey?: TCategory;
}

export default function AddWidgetModal<TCategory extends string = string>({
  isOpen,
  onClose,
  onAddWidget,
  categories,
  items,
  initialCategoryKey,
}: AddWidgetModalProps<TCategory>) {
  const [activeCategory, setActiveCategory] = useState<TCategory | string>(
    (initialCategoryKey as TCategory) ?? categories[0]?.key ?? ''
  );

  if (!isOpen) return null;

  const filteredItems = items.filter(
    (it) => String(it.category) === String(activeCategory)
  );

  const renderPreview = (it: WidgetCatalogItem<TCategory>) => {
    const hasPreview = Boolean(it.preview);
    const content = hasPreview
      ? typeof it.preview === 'function'
        ? (it.preview as () => ReactNode)()
        : it.preview
      : null;

    // Each grid row preview height
    const PREVIEW_ROW_PX = 80;
    const height = Math.max(80, Math.min(400, it.h * PREVIEW_ROW_PX));

    return (
      <div
        className="mt-3 w-full overflow-hidden rounded-md border border-neutral-200 bg-white dark:border-neutral-700 dark:bg-neutral-800"
        style={{ height }}
      >
        <div className="h-full w-full">{content}</div>
      </div>
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="flex h-[80vh] w-full max-w-5xl overflow-hidden rounded-2xl bg-white shadow-xl dark:bg-neutral-800"
        onClick={(e) => e.stopPropagation()}
      >
        <aside className="flex w-56 shrink-0 flex-col border-r bg-neutral-50 p-3 dark:border-neutral-700 dark:bg-neutral-900">
          <div className="mb-2 px-2 text-xs font-semibold uppercase tracking-wider text-neutral-500 dark:text-neutral-400">
            Categories
          </div>
          <nav className="flex flex-1 flex-col gap-1 overflow-y-auto pr-1">
            {categories.map((c) => (
              <button
                key={c.key}
                onClick={() => setActiveCategory(c.key)}
                className={`flex items-center justify-between rounded-lg px-3 py-2 text-sm text-neutral-700 transition-colors hover:bg-neutral-200/70 dark:text-neutral-200 dark:hover:bg-neutral-700/60 ${
                  activeCategory === c.key
                    ? 'bg-neutral-200 font-medium dark:bg-neutral-700'
                    : 'bg-transparent'
                }`}
              >
                <span>{c.label}</span>
              </button>
            ))}
          </nav>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col">
          <header className="flex items-center justify-between border-b p-4 dark:border-neutral-700">
            <h2 className="text-lg font-bold text-neutral-900 dark:text-white">
              Add Widget
            </h2>
            <button
              onClick={onClose}
              className="rounded-full p-1 text-neutral-700 hover:bg-neutral-200 dark:text-neutral-300 dark:hover:bg-neutral-700"
            >
              <XMarkIcon className="h-6 w-6" />
            </button>
          </header>
          <main className="flex-1 overflow-y-auto p-6">
            {filteredItems.length > 0 ? (
              <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                {filteredItems.map((it) => (
                  <div
                    key={it.type}
                    className="flex cursor-pointer flex-col items-stretch justify-start rounded-lg border bg-white p-4 ring-1 ring-black/5 transition-colors hover:bg-neutral-100 dark:border-neutral-700 dark:bg-neutral-800 dark:ring-white/10 dark:hover:bg-neutral-700/50"
                    onClick={() => onAddWidget(it.type)}
                  >
                    <div className="flex items-baseline justify-between gap-2">
                      <div className="mb-2 text-lg font-semibold text-neutral-900 dark:text-white">
                        {it.name}
                      </div>
                      <div className="text-xs capitalize text-neutral-500 dark:text-neutral-400">
                        {it.size} ({it.w}x{it.h})
                      </div>
                    </div>
                    {renderPreview(it)}
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex h-full items-center justify-center text-sm text-neutral-500 dark:text-neutral-400">
                No widgets in this category
              </div>
            )}
          </main>
        </div>
      </div>
    </div>
  );
}
