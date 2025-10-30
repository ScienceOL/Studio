import { BellIcon } from '@heroicons/react/20/solid';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import NotificationItem, { type NotificationProps } from './NotificationItem';

interface NotificationsSectionProps {
  notifications: NotificationProps[];
  isAuthenticated: boolean;
}

export const NotificationsSection = ({
  notifications,
  isAuthenticated,
}: NotificationsSectionProps) => {
  const { t } = useTranslation('translation');

  return (
    <div className="mb-6 rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-800">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="flex items-center text-lg font-bold text-neutral-900 dark:text-white">
          <BellIcon className="mr-2 h-5 w-5 text-neutral-500 dark:text-neutral-400" />
          {t('Notifications')}
        </h2>
        {notifications.some((n) => !n.isRead) && (
          <span className="rounded-full bg-blue-500 px-2 py-0.5 text-xs font-medium text-white">
            {notifications.filter((n) => !n.isRead).length}
          </span>
        )}
      </div>

      {isAuthenticated ? (
        <div className="flex flex-col space-y-3">
          {notifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
            />
          ))}
          {notifications.length === 0 && (
            <p className="text-center text-sm text-neutral-500 dark:text-neutral-400">
              {t('No new notifications')}
            </p>
          )}
        </div>
      ) : (
        <div className="rounded-lg border border-neutral-200 bg-neutral-50 p-4 text-center dark:border-neutral-700 dark:bg-neutral-900">
          <p className="text-sm text-neutral-600 dark:text-neutral-300">
            {t('Sign in to view your notifications')}
          </p>
          <Link
            to="/login"
            className="mt-3 inline-flex items-center rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800"
          >
            {t('Sign In')}
          </Link>
        </div>
      )}
    </div>
  );
};

export default NotificationsSection;
