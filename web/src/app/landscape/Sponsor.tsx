import ECNUBanner from '@/assets/ECNU-Banner.svg';
import SiiLogo from '@/assets/sii-logo.svg';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

export default function Sponsor() {
  const { t } = useTranslation();

  return (
    <section className="bg-neutral-950 py-20">
      <div className="mx-auto max-w-7xl px-6 lg:px-8">
        <motion.p
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          className="text-center text-sm font-medium uppercase tracking-widest text-neutral-600"
        >
          {t('sponsor')}
        </motion.p>

        <div className="mx-auto mt-10 flex max-w-lg items-center justify-center gap-x-12">
          <motion.a
            href="https://sii.edu.cn/"
            target="_blank"
            initial={{ opacity: 0, y: 10 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.1 }}
            className="opacity-40 grayscale invert transition-all hover:opacity-70 hover:grayscale-0"
          >
            <img
              src={SiiLogo}
              alt="Shanghai Innovation Institute"
              className="h-10 w-auto object-contain"
            />
          </motion.a>
          <motion.a
            href="https://english.ecnu.edu.cn/"
            target="_blank"
            initial={{ opacity: 0, y: 10 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.2 }}
            className="opacity-40 grayscale invert transition-all hover:opacity-70 hover:grayscale-0"
          >
            <img
              src={ECNUBanner}
              alt="East China Normal University"
              className="h-10 w-auto object-contain"
            />
          </motion.a>
        </div>
      </div>
    </section>
  );
}
