import { CubeTransparentIcon } from '@heroicons/react/20/solid';
import { Link } from 'react-router-dom';

// Local types inferred from usage
interface Creator {
  username: string;
  avatar?: string;
}

export interface NodeTemplateProps {
  name: string;
  version: string;
  description?: string;
  data: { header?: string };
  creator: Creator;
  updated_at: string | number | Date;
}

interface NodeItemProps {
  node: NodeTemplateProps;
}

export const NodeItem = ({ node }: NodeItemProps) => {
  const nodeLink = `/flociety/nodes/${node.creator.username}/${node.name}`;

  return (
    <Link
      to={nodeLink}
      className="flex cursor-pointer items-start space-x-3 rounded-lg border border-neutral-100 p-3 transition-all duration-200 hover:translate-y-[-2px] hover:border-blue-200 hover:bg-blue-50 hover:shadow-md dark:border-neutral-800 dark:hover:border-blue-800 dark:hover:bg-neutral-800/70"
    >
      <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600 dark:bg-indigo-900/20 dark:text-indigo-400">
        <CubeTransparentIcon className="h-5 w-5" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex justify-between">
          <h3 className="truncate text-sm font-medium text-neutral-900 dark:text-white">
            {node.data.header || node.name}
          </h3>
          <span className="text-xs text-neutral-500 dark:text-neutral-400">
            {node.version}
          </span>
        </div>
        <p className="mt-1 line-clamp-2 text-xs text-neutral-500 dark:text-neutral-400">
          {node.description || 'No description'}
        </p>
        <div className="mt-1 flex items-center gap-2">
          <div className="flex items-center">
            <img
              src={node.creator?.avatar || '/placeholder-avatar.png'}
              alt={node.creator?.username || 'User'}
              className="h-4 w-4 rounded-full"
            />
            <span className="ml-1 text-xs text-neutral-500 dark:text-neutral-400">
              {node.creator?.username || 'Anonymous'}
            </span>
          </div>
          <span className="text-xs text-neutral-400 dark:text-neutral-500">
            {new Date(node.updated_at).toLocaleDateString()}
          </span>
        </div>
      </div>
    </Link>
  );
};

export default NodeItem;
