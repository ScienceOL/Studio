import { GitHubIcon } from '@/assets/SocialIcons';
import {
  BookOpenIcon,
  CloudArrowUpIcon,
  ServerIcon,
} from '@heroicons/react/20/solid';
import { useTranslation } from 'react-i18next';

export default function FeatureOfChat() {
  const { t } = useTranslation();
  const features = [
    {
      name: t('chat.features.discover.name'),
      description: t('chat.features.discover.description'),
      icon: CloudArrowUpIcon,
    },
    {
      name: t('chat.features.connect.name'),
      description: t('chat.features.connect.description'),
      icon: BookOpenIcon,
    },
    {
      name: t('chat.features.accelerate.name'),
      description: t('chat.features.accelerate.description'),
      icon: ServerIcon,
    },
  ];
  return (
    <div className="overflow-hidden bg-white py-24 dark:bg-neutral-950 sm:py-32">
      <div className="mx-auto max-w-7xl px-6 lg:px-8">
        <div className="mx-auto grid max-w-2xl grid-cols-1 gap-x-8 gap-y-16 sm:gap-y-20 lg:mx-0 lg:max-w-none lg:grid-cols-2">
          <div className="lg:pr-8 lg:pt-4">
            <div className="lg:max-w-lg">
              <h2 className="text-base font-semibold leading-7 text-indigo-600">
                {t('chat.subtitle')}
              </h2>
              <p className="mt-2 text-3xl font-bold tracking-tight text-neutral-900 dark:text-neutral-50 sm:text-4xl">
                {t('chat.title')}
              </p>
              <p className="mt-6 text-lg leading-8 text-neutral-600 dark:text-neutral-300">
                {t('chat.description')}
              </p>
              <dl className="mt-10 max-w-xl space-y-8 text-base leading-7 text-neutral-600 dark:text-neutral-300 lg:max-w-none">
                {features.map((feature) => (
                  <div key={feature.name} className="relative pl-9">
                    <dt className="inline font-semibold text-neutral-900 dark:text-neutral-50">
                      <feature.icon
                        className="absolute left-1 top-1 h-5 w-5 text-indigo-600"
                        aria-hidden="true"
                      />
                      {feature.name}
                    </dt>{' '}
                    <dd className="inline">{feature.description}</dd>
                  </div>
                ))}
                <div className="group inline-block pl-1 mt-16 cursor-pointer">
                  <button
                    type="button"
                    name="Github"
                    title="Github"
                    onClick={() => {
                      window.open('https://github.com/ScienceOL');
                    }}
                    className="font-semibold flex items-center leading-8 text-neutral-900 duration-300 dark:text-neutral-100 cursor-pointer"
                  >
                    <GitHubIcon className="size-6 mr-4" />
                    <span className="" aria-hidden="true">
                      Github →
                    </span>
                  </button>
                  <span className="mt-1 block h-[0.1rem] w-full origin-left scale-x-0 transform bg-indigo-600 transition-all duration-200 ease-in-out group-hover:scale-x-100 dark:bg-white"></span>
                </div>
              </dl>
            </div>
          </div>
          <img
            src="https://storage.sciol.ac.cn/library/hero/ScienceOLGithub.png"
            alt="Product screenshot"
            className="w-[48rem] max-w-none rounded-xl shadow-xl ring-1 ring-neutral-400/10 dark:ring-neutral-600/50 sm:w-[57rem] md:-ml-4 lg:-ml-0"
            width={2432}
            height={1442}
          />
        </div>
      </div>
    </div>
  );
}
