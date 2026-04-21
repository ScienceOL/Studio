import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

const products = [
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

        <div className="mt-20 grid grid-cols-1 gap-4 sm:grid-cols-2">
          {products.map((product, i) => {
            const isAvailable = product.statusKey === 'status_available';
            return (
              <motion.div
                key={product.id}
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6, delay: i * 0.1 }}
                className="group relative aspect-[4/3] overflow-hidden rounded-2xl bg-white/[0.02] ring-1 ring-white/[0.06] transition-all hover:ring-white/[0.12] cursor-pointer"
              >
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="relative">
                    <div className="h-24 w-24 rounded-2xl border border-white/[0.06] rotate-12 transition-transform group-hover:rotate-6" />
                    <div className="absolute inset-0 h-24 w-24 rounded-2xl border border-white/[0.04] -rotate-6 transition-transform group-hover:rotate-0" />
                  </div>
                </div>

                <div className="absolute inset-x-0 bottom-0 p-6 bg-gradient-to-t from-black/80 via-black/40 to-transparent">
                  <div className="flex items-end justify-between">
                    <div>
                      <h3 className="text-lg font-semibold text-white">
                        {t(`landing.products.${product.nameKey}`)}
                      </h3>
                      <p className="mt-0.5 text-sm text-white/40 font-mono">
                        {t(`landing.products.${product.subtitleKey}`)}
                      </p>
                    </div>
                    <span
                      className={`rounded-full px-3 py-1 text-[11px] font-mono uppercase tracking-wider ${
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
