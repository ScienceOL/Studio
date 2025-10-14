import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import ProductMatrix from './ProductMatrix';

export default function FeatureOfAbout() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  return (
    <div
      id="introduction"
      className="relative isolate overflow-hidden bg-gradient-to-b from-indigo-100/20 py-4 dark:from-black/10"
    >
      <div
        className="absolute inset-y-0 right-1/2 -z-20 -mr-96 w-[200%] origin-top-right skew-x-[-30deg] bg-white
      shadow-xl shadow-indigo-600/10 ring-1 ring-indigo-50 dark:bg-black/10 dark:ring-indigo-950 sm:-mr-80 lg:-mr-96"
        aria-hidden="true"
      />
      <div className="absolute inset-x-0 top-0 -z-10 h-24 bg-gradient-to-b from-white dark:from-black/10 sm:h-36" />
      <div className="mx-auto max-w-8xl px-6 py-24 sm:py-32 lg:px-8 2xl:max-w-9xl">
        <div className="mx-auto flex max-w-2xl flex-col lg:mx-0 lg:max-w-none">
          <h1
            className="max-w-2xl text-4xl font-bold tracking-tight text-neutral-900
         dark:text-white sm:text-6xl sm:leading-16"
          >
            {t('about.title')}
          </h1>
          <div className="mt-6 max-w-xl">
            <p className="text-lg leading-8 text-neutral-600 dark:text-neutral-300">
              {t('about.description')}
            </p>
            <div className="mt-10 flex items-center justify-start gap-x-6">
              <div className="group inline-block">
                <button
                  type="button"
                  name="Learn more"
                  title="Learn more"
                  onClick={() => {
                    navigate('/manifesto');
                  }}
                  className="relative text-sm font-semibold leading-6 text-neutral-900 group-hover:text-indigo-600 dark:text-neutral-100 dark:group-hover:text-white"
                >
                  {t('about.button.primary')} →
                </button>
                <span className="mt-1 block h-0.5 w-full origin-left scale-x-0 transform bg-indigo-600 transition-all duration-200 ease-in-out group-hover:scale-x-100 dark:bg-white"></span>
              </div>
            </div>
          </div>
          <ProductMatrix />
        </div>
      </div>
      <div className="absolute inset-x-0 bottom-0 -z-10 h-24 bg-gradient-to-t from-white dark:from-neutral-950 sm:h-28" />
    </div>
  );
}
