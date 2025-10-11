import ECNUBanner from '@/assets/ECNU-Banner.svg';
import SiiLogo from '@/assets/sii-logo.svg';
import { useTranslation } from 'react-i18next';
const Sponsor = () => {
  const { t } = useTranslation();

  return (
    <div className="pb-2 dark:bg-neutral-950">
      <div className="mx-auto mt-8 max-w-7xl px-6 sm:mt-16 lg:px-8">
        <h2 className="text-center text-base font-semibold leading-8 text-neutral-400">
          {t('sponsor')}
        </h2>

        <div className="mx-auto w-full max-w-8xl px-6 py-10 sm:py-10 lg:px-8 lg:py-8 2xl:max-w-screen-2xl">
          <div
            className="mx-auto mb-4 grid w-full grid-cols-1 items-center gap-x-8
        gap-y-12 sm:max-w-xl sm:grid-cols-2 sm:gap-x-10
      lg:mx-0 lg:max-w-none lg:grid-cols-2"
          >
            <a href="https://sii.edu.cn/" className="z-0">
              <img
                className="col-span-2 max-h-14  w-full rounded-lg object-contain px-3 py-2 grayscale transition-transform duration-300 ease-in-out hover:scale-110 hover:filter-none dark:invert dark:hover:bg-white dark:hover:filter-none lg:col-span-1"
                src={SiiLogo}
                alt="Shanghai Innovation Institute"
                width={158}
                height={48}
              />
            </a>
            <a href="https://english.ecnu.edu.cn/index.htm/" className="z-0">
              <img
                className="z-10 col-span-2 max-h-14 w-full rounded-lg object-contain px-3 py-2 grayscale transition-transform duration-300 ease-in-out hover:scale-110 hover:filter-none dark:invert dark:hover:bg-white dark:hover:filter-none lg:col-span-1"
                src={ECNUBanner}
                alt="East China Normal University"
                width={158}
                height={48}
              />
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Sponsor;
