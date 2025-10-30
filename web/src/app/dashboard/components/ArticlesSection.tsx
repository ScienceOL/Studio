import { ArrowUpRightIcon, DocumentTextIcon } from '@heroicons/react/20/solid';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

// Local types inferred from usage
export interface ArticleItem {
  uuid: string;
  // Extend as needed by Article component
}

export interface ArticlesPageProps {
  results?: ArticleItem[];
}

interface ArticlesSectionProps {
  articles: ArticlesPageProps;
}

export const ArticlesSection = ({ articles }: ArticlesSectionProps) => {
  const { t } = useTranslation('translation');
  const articlesList = articles.results || [];

  return (
    <div className="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-800">
      <div className="mb-4 flex items-end justify-between">
        <h2 className="flex items-center text-lg font-bold text-neutral-900 dark:text-white">
          <DocumentTextIcon className="mr-2 h-5 w-5 text-neutral-500 dark:text-neutral-400" />
          {t('Recently Published Articles')}
        </h2>

        <Link
          to="/articles"
          className="-m-2 flex items-center rounded-xl px-3 py-2 text-sm text-neutral-600 opacity-70 hover:bg-sky-50 hover:text-teal-600 hover:opacity-100 dark:text-white dark:hover:bg-neutral-800 dark:hover:text-teal-400"
        >
          <span className="mr-1 hidden sm:flex">{t('Read')}</span>
          <span>{t('More')}</span>
          <ArrowUpRightIcon className="ml-2 h-5 w-5 sm:flex" />
        </Link>
      </div>

      <div className="scrollbar-thin scrollbar-thumb-neutral-300 dark:scrollbar-thumb-neutral-600 flex max-h-[600px] flex-col gap-6 overflow-y-auto pr-1">
        {articlesList.map((article) => (
          <div
            key={article.uuid}
            className="rounded-lg border p-3 dark:border-neutral-700"
          >
            {article.uuid}
          </div>
        ))}

        {articlesList.length === 0 && (
          <p className="text-center text-sm text-neutral-500 dark:text-neutral-400">
            {t('No articles found')}
          </p>
        )}
      </div>
    </div>
  );
};

export default ArticlesSection;
