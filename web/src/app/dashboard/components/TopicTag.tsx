import { Link } from 'react-router-dom';

interface TopicTagProps {
  topic: string;
  count: number;
}

export const TopicTag = ({ topic, count }: TopicTagProps) => {
  return (
    <Link
      to={`/topics/${topic}`}
      className="flex items-center justify-between rounded-lg border border-neutral-100 p-2 transition-all hover:border-indigo-200 hover:bg-indigo-50 dark:border-neutral-800 dark:hover:border-indigo-700 dark:hover:bg-indigo-900/20"
    >
      <span className="text-sm font-medium text-neutral-800 dark:text-white">
        {topic}
      </span>
      <span className="rounded-full bg-neutral-100 px-2 py-0.5 text-xs text-neutral-500 dark:bg-neutral-700 dark:text-neutral-300">
        {count}
      </span>
    </Link>
  );
};

export default TopicTag;
