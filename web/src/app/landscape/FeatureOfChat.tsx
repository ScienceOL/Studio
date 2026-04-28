import {
  BoltIcon,
  CurrencyDollarIcon,
  CpuChipIcon,
} from '@heroicons/react/24/solid';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

export default function FeatureOfChat() {
  const { t } = useTranslation();

  const features = [
    {
      name: t('chat.features.discover.name'),
      description: t('chat.features.discover.description'),
      icon: BoltIcon,
    },
    {
      name: t('chat.features.connect.name'),
      description: t('chat.features.connect.description'),
      icon: CurrencyDollarIcon,
    },
    {
      name: t('chat.features.accelerate.name'),
      description: t('chat.features.accelerate.description'),
      icon: CpuChipIcon,
    },
  ];

  return (
    <section className="relative bg-neutral-950 py-32 overflow-hidden">
      {/* Gradient accent */}
      <div className="absolute top-1/2 right-0 -translate-y-1/2 w-[500px] h-[500px] bg-gradient-to-l from-amber-500/5 to-transparent blur-3xl" />

      <div className="relative mx-auto max-w-7xl px-6 lg:px-8">
        <div className="grid grid-cols-1 gap-16 lg:grid-cols-2 lg:gap-24">
          {/* Left: content */}
          <motion.div
            initial={{ opacity: 0, x: -30 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <p className="text-sm font-semibold uppercase tracking-widest text-amber-400">
              {t('chat.subtitle')}
            </p>
            <h2 className="mt-4 text-4xl font-bold tracking-tight text-white sm:text-5xl">
              {t('chat.title')}
            </h2>
            <p className="mt-6 text-lg leading-8 text-neutral-400">
              {t('chat.description')}
            </p>

            <div className="mt-12 space-y-8">
              {features.map((feature, i) => (
                <motion.div
                  key={feature.name}
                  initial={{ opacity: 0, x: -20 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.5, delay: i * 0.1 }}
                  className="flex gap-4"
                >
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-amber-500/10 ring-1 ring-amber-500/20">
                    <feature.icon className="h-5 w-5 text-amber-500" />
                  </div>
                  <div>
                    <h3 className="text-base font-semibold text-white">{feature.name}</h3>
                    <p className="mt-1 text-sm leading-6 text-neutral-400">{feature.description}</p>
                  </div>
                </motion.div>
              ))}
            </div>

            <motion.a
              href="https://github.com/ScienceOL"
              target="_blank"
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.4 }}
              className="mt-12 inline-flex items-center gap-2 text-sm font-semibold text-neutral-300 transition-colors hover:text-white"
            >
              GitHub &rarr;
            </motion.a>
          </motion.div>

          {/* Right: visual placeholder */}
          <motion.div
            initial={{ opacity: 0, x: 30 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.7, delay: 0.2 }}
            className="flex items-center justify-center"
          >
            <div className="relative w-full max-w-lg aspect-square rounded-2xl bg-neutral-900/60 ring-1 ring-neutral-800 overflow-hidden">
              {/* Decorative grid */}
              <div
                className="absolute inset-0 opacity-10"
                style={{
                  backgroundImage:
                    'linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)',
                  backgroundSize: '40px 40px',
                }}
              />
              {/* Agent nodes visualization */}
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="relative">
                  {/* Center node */}
                  <div className="h-16 w-16 rounded-full bg-amber-500/20 ring-2 ring-amber-500/40 flex items-center justify-center">
                    <div className="h-8 w-8 rounded-full bg-amber-500/40" />
                  </div>
                  {/* Orbiting nodes */}
                  {[0, 60, 120, 180, 240, 300].map((deg, i) => {
                    const rad = (deg * Math.PI) / 180;
                    const r = 80;
                    return (
                      <motion.div
                        key={i}
                        className="absolute h-6 w-6 rounded-full bg-neutral-700 ring-1 ring-neutral-600"
                        style={{
                          left: `calc(50% + ${Math.cos(rad) * r}px - 12px)`,
                          top: `calc(50% + ${Math.sin(rad) * r}px - 12px)`,
                        }}
                        animate={{
                          scale: [1, 1.2, 1],
                          opacity: [0.5, 1, 0.5],
                        }}
                        transition={{
                          duration: 3,
                          delay: i * 0.4,
                          repeat: Infinity,
                        }}
                      />
                    );
                  })}
                  {/* Outer ring */}
                  <div
                    className="absolute rounded-full border border-neutral-800"
                    style={{
                      width: 200,
                      height: 200,
                      left: 'calc(50% - 100px)',
                      top: 'calc(50% - 100px)',
                    }}
                  />
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
