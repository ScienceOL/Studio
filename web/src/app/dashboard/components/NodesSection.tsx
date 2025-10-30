import {
  ArrowUpRightIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  CubeTransparentIcon,
} from '@heroicons/react/20/solid';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import NodeItem, { type NodeTemplateProps } from './NodeItem';

interface NodesSectionProps {
  nodes: NodeTemplateProps[];
}

export const NodesSection = ({ nodes }: NodesSectionProps) => {
  const { t } = useTranslation('translation');
  const [currentPage, setCurrentPage] = useState(0);
  const [hovering, setHovering] = useState(false);
  const itemsPerPage = 5; // 每页显示的节点数量
  const pageCount = Math.ceil(nodes.length / itemsPerPage);

  // 获取当前页的节点
  const currentNodes = nodes.slice(
    currentPage * itemsPerPage,
    (currentPage + 1) * itemsPerPage
  );

  const goToNextPage = () => {
    setCurrentPage((prev) => (prev + 1) % pageCount);
  };

  const goToPrevPage = () => {
    setCurrentPage((prev) => (prev - 1 + pageCount) % pageCount);
  };

  return (
    <div className="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-800">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="flex items-center text-lg font-bold text-neutral-900 dark:text-white">
          <CubeTransparentIcon className="mr-2 h-5 w-5 text-neutral-500 dark:text-neutral-400" />
          {t('Recently Published Nodes')}
        </h2>

        <Link
          to="/space"
          className="-m-2 flex items-center rounded-xl px-3 py-2 text-sm text-neutral-600 opacity-70 hover:bg-sky-50 hover:text-teal-600 hover:opacity-100 dark:text-white dark:hover:bg-neutral-800 dark:hover:text-teal-400"
        >
          <span>{t('View All')}</span>
          <ArrowUpRightIcon className="ml-2 h-5 w-5 sm:flex" />
        </Link>
      </div>

      <div className="flex min-h-[280px] w-full flex-col space-y-3 pr-1 pt-1">
        {currentNodes.map((node) => (
          <NodeItem key={`${node.name}-${node.version}`} node={node} />
        ))}

        {nodes.length === 0 && (
          <p className="text-center text-sm text-neutral-500 dark:text-neutral-400">
            {t('No nodes found')}
          </p>
        )}
      </div>

      {/* 分页指示器 - 小圆点和箭头 */}
      {pageCount > 1 && (
        <div
          className="relative mt-4 flex items-center justify-center gap-2"
          onMouseEnter={() => setHovering(true)}
          onMouseLeave={() => setHovering(false)}
        >
          {/* 左箭头 */}
          <button
            onClick={goToPrevPage}
            className={`flex h-6 w-6 items-center justify-center rounded-full transition-all duration-300 ease-in-out hover:bg-indigo-50 dark:hover:bg-neutral-700 ${
              hovering ? 'opacity-100' : 'opacity-0'
            }`}
            aria-label={t('Previous page')}
            disabled={!hovering}
          >
            <ChevronLeftIcon className="h-4 w-4 text-indigo-500 dark:text-indigo-300" />
          </button>

          {/* 小圆点 */}
          <div className="flex space-x-2">
            {Array.from({ length: pageCount }).map((_, index) => (
              <button
                key={index}
                onClick={() => setCurrentPage(index)}
                className={`h-2 w-2 rounded-full transition-colors ${
                  currentPage === index
                    ? 'bg-indigo-300 dark:bg-indigo-400/70'
                    : 'border border-neutral-300 dark:border-neutral-500'
                }`}
                aria-label={`${t('Page')} ${index + 1}`}
              />
            ))}
          </div>

          {/* 右箭头 */}
          <button
            onClick={goToNextPage}
            className={`flex h-6 w-6 items-center justify-center rounded-full transition-all duration-300 ease-in-out hover:bg-indigo-50 dark:hover:bg-neutral-700 ${
              hovering ? 'opacity-100' : 'opacity-0'
            }`}
            aria-label={t('Next page')}
            disabled={!hovering}
          >
            <ChevronRightIcon className="h-4 w-4 text-indigo-500 dark:text-indigo-300" />
          </button>
        </div>
      )}
    </div>
  );
};

export default NodesSection;
