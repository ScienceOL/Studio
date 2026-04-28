import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

const firstParty = [
  { id: 'science-node', nameKey: 'science_node', subtitleKey: 'science_node_subtitle', descKey: 'science_node_description', statusKey: 'status_available', highlight: true },
  { id: 'eyes', nameKey: 'eyes', subtitleKey: 'eyes_subtitle', descKey: 'eyes_description', statusKey: 'status_available', highlight: false },
  { id: 'dongle', nameKey: 'dongle', subtitleKey: 'dongle_subtitle', descKey: 'dongle_description', statusKey: 'status_available', highlight: false },
];

const thirdParty = [
  { id: 'syringe-pump', nameKey: 'syringe_pump', subtitleKey: 'syringe_pump_subtitle', statusKey: 'status_available' },
  { id: 'liquid-handler', nameKey: 'liquid_handler', subtitleKey: 'liquid_handler_subtitle', statusKey: 'status_available' },
  { id: 'xyz-stage', nameKey: 'xyz_stage', subtitleKey: 'xyz_stage_subtitle', statusKey: 'status_coming_soon' },
  { id: 'temperature-control', nameKey: 'temperature_module', subtitleKey: 'temperature_module_subtitle', statusKey: 'status_coming_soon' },
];

export default function ProductShowcase() {
  const { t } = useTranslation();

  return (
    <section id="products" className="relative bg-black py-32 overflow-hidden">
      <div className="mx-auto max-w-screen-2xl px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.7 }}
        >
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-white/30">
            {t('landing.products.label')}
          </p>
          <h2 className="mt-4 text-4xl sm:text-5xl font-bold tracking-tight text-white">
            {t('landing.products.title')}
          </h2>
          <p className="mt-4 max-w-lg text-base text-white/35 font-mono">
            {t('landing.products.description')}
          </p>
        </motion.div>

        <div className="mt-20 grid grid-cols-1 gap-4 lg:grid-cols-2">
          {firstParty.map((product, i) => {
            const isAvailable = product.statusKey === 'status_available';
            return (
              <motion.div
                key={product.id}
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6, delay: i * 0.1 }}
                className={`group relative overflow-hidden rounded-2xl bg-white/[0.02] ring-1 ring-white/[0.06] transition-all hover:ring-white/[0.12] cursor-pointer ${
                  product.highlight ? 'lg:col-span-2 aspect-[21/9]' : 'aspect-[4/3]'
                }`}
              >
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="relative">
                    <div className={`rounded-2xl border border-white/[0.06] rotate-12 transition-transform group-hover:rotate-6 ${product.highlight ? 'h-32 w-32' : 'h-24 w-24'}`} />
                    <div className={`absolute inset-0 rounded-2xl border border-white/[0.04] -rotate-6 transition-transform group-hover:rotate-0 ${product.highlight ? 'h-32 w-32' : 'h-24 w-24'}`} />
                  </div>
                </div>

                <div className="absolute inset-x-0 bottom-0 p-6 bg-gradient-to-t from-black/80 via-black/40 to-transparent">
                  <div className="flex items-end justify-between gap-4">
                    <div className="flex-1">
                      <h3 className={`font-semibold text-white ${product.highlight ? 'text-xl' : 'text-lg'}`}>
                        {t(`landing.products.${product.nameKey}`)}
                      </h3>
                      <p className="mt-0.5 text-sm text-white/40 font-mono">
                        {t(`landing.products.${product.subtitleKey}`)}
                      </p>
                      <p className="mt-2 text-sm text-white/25 leading-relaxed max-w-lg">
                        {t(`landing.products.${product.descKey}`)}
                      </p>
                    </div>
                    <span
                      className={`shrink-0 rounded-full px-3 py-1 text-[11px] font-mono uppercase tracking-wider ${
                        isAvailable
                          ? 'bg-white/[0.08] text-white/60 ring-1 ring-white/[0.1]'
                          : 'bg-white/[0.03] text-white/25 ring-1 ring-white/[0.05]'
                      }`}
                    >
                      {t(`landing.products.${product.statusKey}`)}
                    </span>
                  </div>
                </div>
              </motion.div>
            );
          })}
        </div>

        {/* Third-party compatible devices */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.7 }}
          className="mt-24"
        >
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-white/30">
            {t('landing.products.third_party_label')}
          </p>
          <h3 className="mt-4 text-2xl sm:text-3xl font-bold tracking-tight text-white">
            {t('landing.products.third_party_title')}
          </h3>
        </motion.div>

        <div className="mt-12 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {thirdParty.map((product, i) => {
            const isAvailable = product.statusKey === 'status_available';
            return (
              <motion.div
                key={product.id}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: i * 0.08 }}
                className="group relative overflow-hidden rounded-xl bg-white/[0.02] ring-1 ring-white/[0.06] transition-all hover:ring-white/[0.12] cursor-pointer p-5"
              >
                <h4 className="text-sm font-semibold text-white">
                  {t(`landing.products.${product.nameKey}`)}
                </h4>
                <p className="mt-1 text-xs text-white/35 font-mono">
                  {t(`landing.products.${product.subtitleKey}`)}
                </p>
                <span
                  className={`mt-4 inline-block rounded-full px-2.5 py-0.5 text-[10px] font-mono uppercase tracking-wider ${
                    isAvailable
                      ? 'bg-white/[0.06] text-white/50 ring-1 ring-white/[0.08]'
                      : 'bg-white/[0.03] text-white/20 ring-1 ring-white/[0.04]'
                  }`}
                >
                  {t(`landing.products.${product.statusKey}`)}
                </span>
              </motion.div>
            );
          })}
        </div>

        <motion.p
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.4 }}
          className="mt-12 text-center font-mono text-xs text-white/20"
        >
          {t('landing.products.note')}
        </motion.p>
      </div>
    </section>
  );
}
