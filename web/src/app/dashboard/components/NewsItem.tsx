import { ArrowUpRightIcon } from '@heroicons/react/20/solid';
import { Link } from 'react-router-dom';

// Local types inferred from usage
export interface NewsProps {
  id: string | number;
  title: string;
  content: string;
  category: string;
  created_at: string | number | Date;
  link?: string;
}

interface NewsItemProps {
  news: NewsProps;
}

export const NewsItem = ({ news }: NewsItemProps) => {
  if (news.link) {
    return (
      <Link
        to={news.link}
        className={`group -mx-2 flex cursor-pointer flex-col rounded-lg border-b border-neutral-100 p-3 transition-all duration-200
      last:border-0 hover:-translate-y-0.5 hover:border-indigo-200 hover:bg-indigo-50/60 hover:shadow-sm dark:border-neutral-800 dark:hover:border-indigo-700 dark:hover:bg-indigo-900/10`}
      >
        <div className="mb-1 flex items-center gap-2">
          <span className="rounded bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-300">
            {news.category}
          </span>
          <span className="text-xs text-neutral-500 dark:text-neutral-400">
            {new Date(news.created_at).toLocaleDateString()}
          </span>
        </div>
        <h4
          className={`mb-1 font-medium text-neutral-900 group-hover:text-indigo-700 dark:text-white dark:group-hover:text-indigo-400`}
        >
          {news.title}
        </h4>
        <p className="line-clamp-2 text-sm text-neutral-600 dark:text-neutral-300">
          {news.content}
        </p>
        <div className="mt-2 inline-flex items-center text-xs font-medium text-indigo-600 group-hover:text-indigo-700 dark:text-indigo-400 dark:group-hover:text-indigo-300">
          Read more
          <ArrowUpRightIcon className="ml-1 h-3 w-3 transform transition-transform group-hover:translate-x-0.5 group-hover:translate-y-[-0.5px]" />
        </div>
      </Link>
    );
  } else {
    return (
      <div
        className={`group -mx-2 flex flex-col rounded-lg border-b border-neutral-100 p-3 transition-all duration-200 last:border-0
      hover:bg-neutral-50 dark:border-neutral-800 dark:hover:bg-neutral-800/70`}
      >
        <div className="mb-1 flex items-center gap-2">
          <span className="rounded bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-300">
            {news.category}
          </span>
          <span className="text-xs text-neutral-500 dark:text-neutral-400">
            {new Date(news.created_at).toLocaleDateString()}
          </span>
        </div>
        <h4 className={`mb-1 font-medium text-neutral-900 dark:text-white`}>
          {news.title}
        </h4>
        <p className="line-clamp-2 text-sm text-neutral-600 dark:text-neutral-300">
          {news.content}
        </p>
      </div>
    );
  }
};

export default NewsItem;
