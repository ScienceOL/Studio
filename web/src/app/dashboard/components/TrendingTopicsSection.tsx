import { DocumentTextIcon } from '@heroicons/react/20/solid';
import { useTranslation } from 'react-i18next';
import TopicTag from './TopicTag';

interface TrendingTopicsSectionProps {
  topics: { name: string; count: number }[];
}

export const TrendingTopicsSection = ({
  topics,
}: TrendingTopicsSectionProps) => {
  const { t } = useTranslation('translation');

  return (
    <div className="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-800">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="flex items-center text-lg font-bold text-neutral-900 dark:text-white">
          <DocumentTextIcon className="mr-2 h-5 w-5 text-neutral-500 dark:text-neutral-400" />
          {t('Trending Topics')}
        </h2>
      </div>

      {topics.length > 0 ? (
        <div className="grid grid-cols-2 gap-2">
          {topics.map((topic) => (
            <TopicTag key={topic.name} topic={topic.name} count={topic.count} />
          ))}
        </div>
      ) : (
        <p className="text-center text-sm text-neutral-500 dark:text-neutral-400">
          {t('No trending topics')}
        </p>
      )}
    </div>
  );
};

export default TrendingTopicsSection;
