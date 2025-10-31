import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import {
  SiProton,
  SiStmicroelectronics,
  SiUnity,
  SiUnrealengine,
  SiX,
} from 'react-icons/si';

export default function ProductMatrix() {
  const { t } = useTranslation();

  const products = [
    {
      id: 'studio',
      icon: SiStmicroelectronics,
      color: 'text-sky-500',
      href: 'https://github.com/ScienceOL/Studio',
    },
    {
      id: 'protium',
      icon: SiProton,
      color: 'text-indigo-500',
      href: 'https://github.com/ScienceOL/Protium',
    },
    {
      id: 'xyzen',
      icon: SiX,
      color: 'text-amber-500',
      href: 'https://github.com/ScienceOL',
    },
    {
      id: 'anti',
      icon: SiUnity,
      color: 'text-rose-500',
      href: 'https://github.com/ScienceOL',
    },
    {
      id: 'labos',
      icon: SiUnrealengine,
      color: 'text-emerald-500',
      href: 'https://github.com/Uni-Lab-OS',
    },
  ];

  return (
    <div className="mt-12 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
      {products.map((product, i) => (
        <motion.div
          key={product.id}
          initial={{ opacity: 0, y: 20 }}
          animate={{
            opacity: 1,
            y: 0,
            transition: {
              delay: i * 0.1,
              duration: 0.5,
              ease: 'easeOut',
            },
          }}
          whileHover={{ y: -5, scale: 1.03 }}
          className="group relative cursor-pointer"
          onClick={() => window.open(product.href, '_blank')}
        >
          <div className="h-full rounded-xl border border-neutral-200 bg-white/50 p-6 shadow-sm transition-all duration-300 group-hover:border-indigo-300 group-hover:shadow-md dark:border-neutral-800 dark:bg-neutral-900/50 dark:group-hover:border-indigo-700">
            <div className="flex items-center gap-4">
              <product.icon className={`h-8 w-8 ${product.color}`} />
              <h3 className="text-lg font-bold text-neutral-900 dark:text-white">
                {t(`products.${product.id}.name`)}
              </h3>
            </div>
            <p className="mt-3 text-sm text-neutral-600 dark:text-neutral-400">
              {t(`products.${product.id}.description`)}
            </p>
          </div>
        </motion.div>
      ))}
    </div>
  );
}
