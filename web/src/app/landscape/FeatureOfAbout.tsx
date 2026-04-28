import {
  CommandLineIcon,
  CpuChipIcon,
  CloudIcon,
} from '@heroicons/react/24/solid';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

const layers = [
  {
    id: 'osdl',
    icon: CpuChipIcon,
    color: 'emerald',
    href: '#osdl',
  },
  {
    id: 'opensdl',
    icon: CommandLineIcon,
    color: 'sky',
    href: 'https://github.com/ScienceOL/OpenSDL',
  },
  {
    id: 'xyzen',
    icon: CloudIcon,
    color: 'amber',
    href: 'https://xyzen.cc',
  },
];

const colorMap: Record<string, { bg: string; text: string; ring: string; icon: string }> = {
  emerald: {
    bg: 'bg-emerald-500/10',
    text: 'text-emerald-400',
    ring: 'ring-emerald-500/20',
    icon: 'text-emerald-500',
  },
  sky: {
    bg: 'bg-sky-500/10',
    text: 'text-sky-400',
    ring: 'ring-sky-500/20',
    icon: 'text-sky-500',
  },
  amber: {
    bg: 'bg-amber-500/10',
    text: 'text-amber-400',
    ring: 'ring-amber-500/20',
    icon: 'text-amber-500',
  },
};

export default function FeatureOfAbout() {
  const { t } = useTranslation();

  return (
    <section id="introduction" className="relative bg-neutral-950 py-32 overflow-hidden">
      {/* Subtle accent */}
      <div className="absolute inset-0 bg-gradient-to-b from-neutral-950 via-neutral-900/50 to-neutral-950" />

      <div className="relative mx-auto max-w-7xl px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="max-w-2xl"
        >
          <p className="text-sm font-semibold uppercase tracking-widest text-emerald-400">
            Ecosystem
          </p>
          <h2 className="mt-4 text-4xl font-bold tracking-tight text-white sm:text-5xl">
            {t('about.title')}
          </h2>
          <p className="mt-6 text-lg leading-8 text-neutral-400">
            {t('about.description')}
          </p>
        </motion.div>

        {/* Three-layer cards */}
        <div className="mt-20 grid grid-cols-1 gap-6 sm:grid-cols-3">
          {layers.map((layer, i) => {
            const colors = colorMap[layer.color];
            return (
              <motion.a
                key={layer.id}
                href={layer.href}
                target={layer.href.startsWith('http') ? '_blank' : undefined}
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: i * 0.1 }}
                whileHover={{ y: -4 }}
                className="group relative rounded-2xl bg-neutral-900/80 p-8 ring-1 ring-neutral-800 transition-colors hover:ring-neutral-700"
              >
                {/* Layer number */}
                <span className="absolute top-6 right-6 text-6xl font-bold text-neutral-800/60">
                  {String(i + 1).padStart(2, '0')}
                </span>

                <div className={`inline-flex rounded-lg p-2.5 ${colors.bg} ring-1 ${colors.ring}`}>
                  <layer.icon className={`h-6 w-6 ${colors.icon}`} />
                </div>

                <h3 className="mt-6 text-xl font-bold text-white">
                  {t(`products.${layer.id}.name`)}
                </h3>
                <p className="mt-3 text-sm leading-6 text-neutral-400">
                  {t(`products.${layer.id}.description`)}
                </p>

                <div className={`mt-6 flex items-center gap-1 text-sm font-medium ${colors.text}`}>
                  <span className="transition-transform group-hover:translate-x-0.5">
                    Learn more &rarr;
                  </span>
                </div>
              </motion.a>
            );
          })}
        </div>
      </div>
    </section>
  );
}
