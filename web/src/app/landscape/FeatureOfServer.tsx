import LogoLoading from '@/components/basic/loading';
import {
  BeakerIcon,
  CircleStackIcon,
  ComputerDesktopIcon,
  CpuChipIcon,
  CubeTransparentIcon,
  PuzzlePieceIcon,
} from '@heroicons/react/20/solid';
import { lazy, Suspense } from 'react';
import { useTranslation } from 'react-i18next';

// 懒加载 3D 组件以优化首屏性能
const LabScene3D = lazy(() => import('./LabScene3D'));

export default function FeatureOfServer() {
  const { t } = useTranslation();
  const features = [
    {
      name: t('server.features.robotics.name'),
      description: t('server.features.robotics.description'),
      icon: PuzzlePieceIcon,
    },
    {
      name: t('server.features.ai.name'),
      description: t('server.features.ai.description'),
      icon: CpuChipIcon,
    },
    {
      name: t('server.features.data.name'),
      description: t('server.features.data.description'),
      icon: CircleStackIcon,
    },
    {
      name: t('server.features.digital-twin.name'),
      description: t('server.features.digital-twin.description'),
      icon: CubeTransparentIcon,
    },
    {
      name: t('server.features.remote.name'),
      description: t('server.features.remote.description'),
      icon: ComputerDesktopIcon,
    },
    {
      name: t('server.features.modular.name'),
      description: t('server.features.modular.description'),
      icon: BeakerIcon,
    },
  ];
  return (
    <div className="bg-white py-24 dark:bg-neutral-950 sm:py-32">
      <div className="mx-auto max-w-7xl px-6 lg:px-8">
        <div className="mx-auto max-w-2xl sm:text-center">
          <h2 className="text-base font-semibold leading-7 text-indigo-600 dark:text-indigo-400">
            {t('server.subtitle')}
          </h2>
          <p className="mt-2 text-3xl font-bold tracking-tight text-neutral-900 dark:text-neutral-50 sm:text-4xl">
            {t('server.title')}
          </p>
          <p className="mt-6 text-lg leading-8 text-neutral-600 dark:text-neutral-300">
            {t('server.description')}
          </p>
        </div>
      </div>
      <div className="relative overflow-hidden pt-16">
        <div className="mx-auto max-w-7xl px-6 lg:px-8">
          {/* 3D 模型展示区域 */}
          <div className="relative h-[500px] overflow-hidden rounded-xl bg-gradient-to-br from-indigo-50 to-purple-50 shadow-2xl ring-1 ring-neutral-950/10 dark:from-neutral-900 dark:to-neutral-800 dark:ring-neutral-50/10">
            <Suspense
              fallback={
                <div className="flex h-full w-full items-center justify-center">
                  <LogoLoading variant="large" animationType="galaxy" />
                </div>
              }
            >
              <LabScene3D />
            </Suspense>
          </div>
          <div className="relative" aria-hidden="true">
            <div className="absolute -inset-x-20 bottom-0 bg-gradient-to-t from-white pt-[7%] dark:from-neutral-950" />
          </div>
        </div>
      </div>
      <div className="mx-auto mt-16 max-w-7xl px-6 sm:mt-20 md:mt-24 lg:px-8">
        <dl className="mx-auto grid max-w-2xl grid-cols-1 gap-x-6 gap-y-10 text-base leading-7 text-neutral-600 dark:text-neutral-300 sm:grid-cols-2 lg:mx-0 lg:max-w-none lg:grid-cols-3 lg:gap-x-8 lg:gap-y-16">
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
        </dl>
      </div>
    </div>
  );
}
