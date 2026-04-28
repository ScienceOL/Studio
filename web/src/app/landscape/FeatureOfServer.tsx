import {
  PuzzlePieceIcon,
  CpuChipIcon,
  CircleStackIcon,
  CubeTransparentIcon,
  ComputerDesktopIcon,
  WrenchScrewdriverIcon,
} from '@heroicons/react/24/solid';
import { motion } from 'framer-motion';
import { lazy, Suspense } from 'react';
import { useTranslation } from 'react-i18next';

const LabScene3D = lazy(() => import('./LabScene3D'));

export default function FeatureOfServer() {
  const { t } = useTranslation();

  const features = [
    { key: 'robotics', icon: PuzzlePieceIcon },
    { key: 'ai', icon: CpuChipIcon },
    { key: 'data', icon: CircleStackIcon },
    { key: 'digital-twin', icon: CubeTransparentIcon },
    { key: 'remote', icon: ComputerDesktopIcon },
    { key: 'modular', icon: WrenchScrewdriverIcon },
  ];

  return (
    <section id="osdl" className="relative bg-neutral-950 py-32 overflow-hidden">
      {/* Gradient accent */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[600px] h-[400px] bg-gradient-to-b from-sky-500/5 to-transparent blur-3xl" />

      <div className="relative mx-auto max-w-7xl px-6 lg:px-8">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="mx-auto max-w-2xl text-center"
        >
          <p className="text-sm font-semibold uppercase tracking-widest text-sky-400">
            {t('server.subtitle')}
          </p>
          <h2 className="mt-4 text-4xl font-bold tracking-tight text-white sm:text-5xl">
            {t('server.title')}
          </h2>
          <p className="mt-6 text-lg leading-8 text-neutral-400">
            {t('server.description')}
          </p>
        </motion.div>

        {/* 3D Scene */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.7, delay: 0.2 }}
          className="mt-16"
        >
          <div className="relative h-[500px] overflow-hidden rounded-2xl bg-neutral-900/60 ring-1 ring-neutral-800">
            <Suspense
              fallback={
                <div className="flex h-full w-full items-center justify-center">
                  <div className="h-8 w-8 animate-spin rounded-full border-2 border-neutral-700 border-t-sky-400" />
                </div>
              }
            >
              <LabScene3D />
            </Suspense>
          </div>
        </motion.div>

        {/* Feature grid */}
        <div className="mt-20 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature, i) => (
            <motion.div
              key={feature.key}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.08 }}
              className="group rounded-xl bg-neutral-900/40 p-6 ring-1 ring-neutral-800/60 transition-colors hover:ring-neutral-700"
            >
              <div className="flex items-start gap-4">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-sky-500/10 ring-1 ring-sky-500/20">
                  <feature.icon className="h-4.5 w-4.5 text-sky-500" />
                </div>
                <div>
                  <h3 className="text-sm font-semibold text-white">
                    {t(`server.features.${feature.key}.name`)}
                  </h3>
                  <p className="mt-1.5 text-sm leading-6 text-neutral-500">
                    {t(`server.features.${feature.key}.description`)}
                  </p>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
