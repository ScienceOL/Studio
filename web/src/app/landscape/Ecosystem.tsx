import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

export default function Ecosystem() {
  const { t } = useTranslation();

  return (
    <section className="relative bg-black py-32 overflow-hidden border-t border-white/[0.04]">
      <div className="mx-auto max-w-screen-2xl px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.7 }}
        >
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-white/30">
            {t('landing.ecosystem.label')}
          </p>
          <h2 className="mt-4 text-4xl sm:text-5xl font-bold tracking-tight text-white">
            {t('landing.ecosystem.title')}
          </h2>
        </motion.div>

        <div className="mt-20 grid grid-cols-1 gap-6 lg:grid-cols-2">
          {/* OpenSDL */}
          <motion.a
            href="https://github.com/ScienceOL/OpenSDL"
            target="_blank"
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
            className="group relative rounded-2xl bg-white/[0.02] p-8 sm:p-10 ring-1 ring-white/[0.06] transition-all hover:ring-white/[0.12]"
          >
            <div className="inline-flex items-center gap-2 rounded-full bg-white/[0.05] px-3 py-1 ring-1 ring-white/[0.08]">
              <div className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
              <span className="font-mono text-[11px] uppercase tracking-wider text-white/50">
                {t('landing.ecosystem.opensdl_tag')}
              </span>
            </div>

            <h3 className="mt-6 text-2xl font-bold text-white">OpenSDL</h3>
            <p className="mt-3 text-sm leading-relaxed text-white/35 max-w-md">
              {t('landing.ecosystem.opensdl_description')}
            </p>

            <div className="mt-8 flex items-center gap-2 font-mono text-xs text-white/30 group-hover:text-white/50 transition-colors">
              <span>github.com/ScienceOL/OpenSDL</span>
              <span className="transition-transform group-hover:translate-x-0.5">&rarr;</span>
            </div>
          </motion.a>

          {/* Xyzen */}
          <motion.a
            href="https://xyzen.cc"
            target="_blank"
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, delay: 0.1 }}
            className="group relative rounded-2xl bg-white/[0.02] p-8 sm:p-10 ring-1 ring-white/[0.06] transition-all hover:ring-white/[0.12]"
          >
            <div className="inline-flex items-center gap-2 rounded-full bg-white/[0.05] px-3 py-1 ring-1 ring-white/[0.08]">
              <div className="h-1.5 w-1.5 rounded-full bg-blue-400" />
              <span className="font-mono text-[11px] uppercase tracking-wider text-white/50">
                {t('landing.ecosystem.xyzen_tag')}
              </span>
            </div>

            <h3 className="mt-6 text-2xl font-bold text-white">Xyzen</h3>
            <p className="mt-3 text-sm leading-relaxed text-white/35 max-w-md">
              {t('landing.ecosystem.xyzen_description')}
            </p>

            <div className="mt-8 flex items-center gap-2 font-mono text-xs text-white/30 group-hover:text-white/50 transition-colors">
              <span>xyzen.cc</span>
              <span className="transition-transform group-hover:translate-x-0.5">&rarr;</span>
            </div>
          </motion.a>
        </div>

        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 1, delay: 0.3 }}
          className="mt-16 flex items-center justify-center gap-4"
        >
          <div className="h-px flex-1 bg-gradient-to-r from-transparent to-white/[0.06]" />
          <p className="font-mono text-[11px] text-white/20 uppercase tracking-[0.15em] whitespace-nowrap px-4">
            {t('landing.ecosystem.flow')}
          </p>
          <div className="h-px flex-1 bg-gradient-to-l from-transparent to-white/[0.06]" />
        </motion.div>
      </div>
    </section>
  );
}
