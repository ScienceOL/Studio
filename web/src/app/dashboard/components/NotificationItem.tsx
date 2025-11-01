import {
  BeakerIcon,
  BellIcon,
  CubeTransparentIcon,
  DocumentTextIcon,
} from '@heroicons/react/20/solid';
import { Link } from 'react-router-dom';

// Local types inferred from usage
export type NotificationType = 'lab' | 'device' | 'alert' | 'default';

export interface NotificationProps {
  id: string | number;
  type: NotificationType;
  title: string;
  content?: string;
  isRead: boolean;
  timestamp: string | number | Date;
  link?: string;
}

interface NotificationItemProps {
  notification: NotificationProps;
}

export const NotificationItem = ({ notification }: NotificationItemProps) => {
  // 根据通知类型选择不同的颜色和图标
  const getTypeStyles = () => {
    switch (notification.type) {
      case 'lab':
        return {
          icon: <BeakerIcon className="h-5 w-5" />,
          bgColor: 'bg-indigo-50 dark:bg-indigo-900/20',
          textColor: 'text-indigo-600 dark:text-indigo-400',
          borderColor:
            'group-hover:border-indigo-300 dark:group-hover:border-indigo-700',
        };
      case 'device':
        return {
          icon: <CubeTransparentIcon className="h-5 w-5" />,
          bgColor: 'bg-green-50 dark:bg-green-900/20',
          textColor: 'text-green-600 dark:text-green-400',
          borderColor:
            'group-hover:border-green-300 dark:group-hover:border-green-700',
        };
      case 'alert':
        return {
          icon: <BellIcon className="h-5 w-5" />,
          bgColor: 'bg-red-50 dark:bg-red-900/20',
          textColor: 'text-red-600 dark:text-red-400',
          borderColor:
            'group-hover:border-red-300 dark:group-hover:border-red-700',
        };
      default:
        return {
          icon: <DocumentTextIcon className="h-5 w-5" />,
          bgColor: 'bg-indigo-50 dark:bg-indigo-900/20',
          textColor: 'text-indigo-600 dark:text-indigo-400',
          borderColor:
            'group-hover:border-indigo-300 dark:group-hover:border-indigo-700',
        };
    }
  };

  const styles = getTypeStyles();

  // 格式化时间，显示到分钟
  const formattedTime = new Date(notification.timestamp).toLocaleString(
    undefined,
    {
      year: 'numeric',
      month: 'numeric',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }
  );

  return (
    <Link
      to={notification.link || '#'}
      className={`group flex items-start space-x-3 rounded-lg border ${
        notification.isRead
          ? 'border-neutral-100 dark:border-neutral-800'
          : 'border-indigo-100 dark:border-indigo-900/30'
      } p-3 transition-all duration-200 hover:translate-y-[-2px] hover:bg-neutral-50 hover:shadow-sm dark:hover:bg-neutral-800/70`}
    >
      <div
        className={`flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg ${styles.bgColor} ${styles.textColor}`}
      >
        {styles.icon}
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between">
          <h4
            className={`text-sm font-medium ${
              !notification.isRead
                ? 'font-semibold text-black dark:text-white'
                : 'text-neutral-900 dark:text-neutral-200'
            }`}
          >
            {notification.title}
          </h4>
          {!notification.isRead && (
            <span className="ml-2 h-2 w-2 flex-shrink-0 rounded-full bg-indigo-500"></span>
          )}
        </div>
        {notification.content && (
          <p className="mt-1 line-clamp-2 text-xs text-neutral-500 dark:text-neutral-400">
            {notification.content}
          </p>
        )}
        <span className="mt-1 text-xs text-neutral-400 dark:text-neutral-500">
          {formattedTime}
        </span>
      </div>
    </Link>
  );
};

export default NotificationItem;
