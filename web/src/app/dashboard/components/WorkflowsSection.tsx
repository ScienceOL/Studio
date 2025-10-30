import { ArrowUpRightIcon, BeakerIcon } from '@heroicons/react/20/solid';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

// Local types inferred from usage
export interface WorkflowPostProps {
  workflow: { uuid: string };
}

interface WorkflowsSectionProps {
  workflows: WorkflowPostProps[];
}

export const WorkflowsSection = ({ workflows }: WorkflowsSectionProps) => {
  const { t } = useTranslation('translation');

  return (
    <div className="rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-800">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="flex items-center text-lg font-bold text-neutral-900 dark:text-white">
          <BeakerIcon className="mr-2 h-5 w-5 text-neutral-500 dark:text-neutral-400" />
          {t('Recently Created Workflows')}
        </h2>

        <Link
          to="/flociety/workflows"
          className="-m-2 flex items-center rounded-xl px-3 py-2 text-sm text-neutral-600 opacity-70 hover:bg-sky-50 hover:text-teal-600 hover:opacity-100 dark:text-white dark:hover:bg-neutral-800 dark:hover:text-teal-400"
        >
          <span>{t('View All')}</span>
          <ArrowUpRightIcon className="ml-2 h-5 w-5 sm:flex" />
        </Link>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {workflows && workflows.length > 0 ? (
          workflows.slice(0, 2).map((post: WorkflowPostProps) => (
            <div
              key={post.workflow.uuid}
              className="group relative isolate flex h-48 flex-col justify-end overflow-hidden rounded-2xl bg-neutral-800 px-6 pb-6 sm:h-56 lg:h-64"
            >
              <div className="text-white">Workflow {post.workflow.uuid}</div>
            </div>
          ))
        ) : (
          <p className="col-span-2 text-center text-sm text-neutral-500 dark:text-neutral-400">
            {t('No workflows found')}
          </p>
        )}
      </div>
    </div>
  );
};

export default WorkflowsSection;
