import {
  ArrowUpRightIcon,
  BeakerIcon,
  CubeTransparentIcon,
  DocumentTextIcon,
} from '@heroicons/react/20/solid';
import { Link } from 'react-router-dom';
import type { ActivityProps } from './UserActivitySection';

interface ActivityItemProps {
  activity: ActivityProps;
}

export const ActivityItem = ({ activity }: ActivityItemProps) => {
  // 根据活动类型选择不同的图标
  const getIcon = () => {
    switch (activity.type) {
      case 'workflow':
        return <BeakerIcon className="h-5 w-5" />;
      case 'article':
        return <DocumentTextIcon className="h-5 w-5" />;
      case 'node':
        return <CubeTransparentIcon className="h-5 w-5" />;
      case 'fork':
        return <ArrowUpRightIcon className="h-5 w-5" />;
      default:
        return <DocumentTextIcon className="h-5 w-5" />;
    }
  };

  return (
    <Link
      to={activity.link || '#'}
      className="flex items-start space-x-3 rounded-lg border border-neutral-100 p-3 transition-all hover:bg-neutral-50 dark:border-neutral-800 dark:hover:bg-neutral-800/50"
    >
      <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400">
        {getIcon()}
      </div>
      <div className="min-w-0 flex-1">
        <h4 className="truncate text-sm font-medium text-neutral-900 dark:text-white">
          {activity.title}
        </h4>
        {activity.description && (
          <p className="mt-1 line-clamp-1 text-xs text-neutral-500 dark:text-neutral-400">
            {activity.description}
          </p>
        )}
        <span className="mt-1 text-xs text-neutral-400 dark:text-neutral-500">
          {new Date(activity.created_at).toLocaleDateString()}
        </span>
      </div>
    </Link>
  );
};

export default ActivityItem;
