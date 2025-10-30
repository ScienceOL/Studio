import { BellIcon } from '@heroicons/react/20/solid';
import { useTranslation } from 'react-i18next';
import NewsItem, { type NewsProps } from './NewsItem';

interface NewsSectionProps {
  news: NewsProps[];
}

export const NewsSection = ({ news }: NewsSectionProps) => {
  const { t } = useTranslation('translation');

  return (
    <div className="mb-6 overflow-x-hidden rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-800">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="flex items-center text-lg font-bold text-neutral-900 dark:text-white">
          <BellIcon className="mr-2 h-5 w-5 text-neutral-500 dark:text-neutral-400" />
          {t('News & Updates')}
        </h2>
      </div>

      <div className="flex max-w-full flex-col space-y-4 overflow-y-auto overflow-x-hidden pr-1">
        {news.map((item) => (
          <NewsItem key={item.id} news={item} />
        ))}

        {news.length === 0 && (
          <p className="text-center text-sm text-neutral-500 dark:text-neutral-400">
            {t('No news available')}
          </p>
        )}
      </div>
    </div>
  );
};

export default NewsSection;
