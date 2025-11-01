import { UserCircleIcon } from '@heroicons/react/20/solid';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import ActivityItem from './ActivityItem';

// Local types inferred from usage
export interface ActivityProps {
  id: string | number;
  type: 'workflow' | 'article' | 'node' | 'fork' | string;
  title: string;
  description?: string;
  link?: string;
  created_at: string | number | Date;
}

interface UserActivitySectionProps {
  activities: ActivityProps[];
  isAuthenticated: boolean;
}

export const UserActivitySection = ({
  activities,
  isAuthenticated,
}: UserActivitySectionProps) => {
  const { t } = useTranslation('translation');

  return (
    <div className="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-800">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="flex items-center text-lg font-bold text-neutral-900 dark:text-white">
          <UserCircleIcon className="mr-2 h-5 w-5 text-neutral-500 dark:text-neutral-400" />
          {t('Your Activity')}
        </h2>
      </div>

      {isAuthenticated ? (
        <div className="flex flex-col space-y-3">
          {activities.map((activity) => (
            <ActivityItem key={activity.id} activity={activity} />
          ))}
          {activities.length === 0 && (
            <p className="text-center text-sm text-neutral-500 dark:text-neutral-400">
              {t('No recent activity')}
            </p>
          )}
        </div>
      ) : (
        <div className="rounded-lg border border-neutral-200 bg-neutral-50 p-4 text-center dark:border-neutral-700 dark:bg-neutral-900">
          <p className="text-sm text-neutral-600 dark:text-neutral-300">
            {t('Sign in to view your activity')}
          </p>
          <Link
            to="/login"
            className="mt-3 inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-700 dark:bg-indigo-700 dark:hover:bg-indigo-800"
          >
            {t('Sign In')}
          </Link>
        </div>
      )}
    </div>
  );
};

export default UserActivitySection;
