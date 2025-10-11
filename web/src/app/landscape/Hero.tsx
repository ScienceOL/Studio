import ColorfulPattern from '@/components/basic/patterns/ColorfulPattern';
import GridBackground from '@/components/basic/patterns/GridBackground';
import { Suspense, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';

import { useNavigate } from 'react-router-dom';

export default function Hero() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (videoRef.current) {
      videoRef.current.play();
    }
  }, []);

  return (
    <div className="bg-white dark:bg-neutral-950">
      <main>
        <div className="relative isolate">
          {/* Background */}
          <GridBackground />
          {/* Background */}
          <ColorfulPattern />

          {/* Content */}
          <div className="overflow-hidden">
            <div className="mx-auto max-w-8xl px-6 pb-32 pt-36 sm:pt-60 lg:px-8 lg:pt-32 2xl:max-w-9xl 2xl:pt-48">
              <div className="lg: mx-auto max-w-2xl justify-between gap-x-14 lg:mx-0 lg:flex lg:max-w-none lg:items-center">
                <div className="relative w-full max-w-xl lg:shrink-0 lg:pb-12 xl:max-w-2xl">
                  <div className="hidden sm:mb-8 sm:flex sm:justify-start">
                    <div className="relative rounded-full px-3 py-1 text-sm leading-6 text-neutral-600 ring-1 ring-neutral-900/10 duration-300 hover:scale-105 hover:ring-neutral-900/20 dark:text-neutral-400 dark:ring-neutral-100/10 dark:hover:ring-neutral-100/20">
                      {t('Find our latest feature update.')}
                      <a
                        href="/login"
                        className="font-semibold text-indigo-600"
                      >
                        <span className="absolute inset-0" aria-hidden="true" />
                        {t('Read more')}
                        <span aria-hidden="true"> &rarr;</span>
                      </a>
                    </div>
                  </div>
                  <h1 className="text-4xl font-bold tracking-tight text-neutral-900 dark:text-white sm:text-6xl">
                    {/* <LogoBanner className="sm:h-24 mb-8 h-16 w-auto max-w-[calc(100%-1rem)] fill-indigo-800 dark:fill-[#6f6be2] " /> */}
                    <div className=" h-fit w-full animate-gradient-flow bg-gradient-flow bg-[length:400%_400%] bg-clip-text pb-8 text-transparent">
                      <span className="text-7xl font-bold">Science OL</span>
                    </div>
                    {t('hero.title')}
                  </h1>
                  <p className="mt-6 text-lg leading-8 text-neutral-600 dark:text-neutral-300 sm:max-w-md lg:max-w-none">
                    {t('hero.description')}
                  </p>
                  <div className="mt-10 flex items-center justify-start gap-x-6">
                    <button
                      type="button"
                      name="Get started"
                      title="Get started"
                      onClick={() => {
                        navigate('/login');
                      }}
                      className="rounded-md bg-indigo-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500
                        focus-visible:outline focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                    >
                      {t('hero.button.primary')}
                    </button>
                    <div className="group inline-block">
                      <button
                        type="button"
                        name="Learn more"
                        title="Learn more"
                        onClick={() => {
                          navigate('/about');
                        }}
                        className="text-sm font-semibold leading-6 text-neutral-900 duration-300 dark:text-neutral-100"
                      >
                        <span className="" aria-hidden="true">
                          {t('hero.button.secondary')} →
                        </span>
                      </button>
                      <span className="mt-1 block h-[0.1rem] w-full origin-left scale-x-0 transform bg-indigo-600 transition-all duration-200 ease-in-out group-hover:scale-x-100 dark:bg-white"></span>
                    </div>
                  </div>
                </div>
                <div className="mt-14 flex justify-end gap-8 sm:-mt-44 sm:justify-start sm:pl-20 lg:mt-0 lg:pl-0">
                  <div className="ml-auto w-44 flex-none space-y-8 pt-32 sm:ml-0 sm:pt-80 lg:order-last lg:pt-36 xl:order-none xl:pt-80">
                    <div className="relative">
                      <img
                        src="/hero5.png"
                        alt=""
                        className="aspect-[2/3] w-full rounded-xl bg-neutral-900/5 object-cover shadow-lg"
                      />
                      <div className="pointer-events-none absolute inset-0 rounded-xl ring-1 ring-inset ring-neutral-900/10" />
                    </div>
                  </div>
                  <div className="mr-auto w-44 flex-none space-y-8 sm:mr-0 sm:pt-52 lg:pt-36">
                    <div className="relative">
                      <Suspense
                        fallback={
                          <img
                            src="/hero10.png"
                            alt=""
                            className="aspect-[2/3] w-full rounded-xl bg-neutral-900/5 object-cover shadow-lg"
                          />
                        }
                      >
                        <video
                          ref={videoRef}
                          src="/hero10.mp4"
                          autoPlay
                          muted
                          loop
                          playsInline
                          poster="/hero10.png"
                          className="aspect-[2/3] w-full rounded-xl bg-neutral-900/5 object-cover  shadow-lg"
                        />
                      </Suspense>
                      <div className="pointer-events-none absolute inset-0 rounded-xl ring-1 ring-inset ring-neutral-900/10" />
                    </div>
                    <div className="relative">
                      <img
                        src="/hero4.png"
                        alt=""
                        className="aspect-[2/3] w-full rounded-xl bg-neutral-900/5 object-cover shadow-lg"
                      />
                      <div className="pointer-events-none absolute inset-0 rounded-xl ring-1 ring-inset ring-neutral-900/10" />
                    </div>
                  </div>
                  <div className="w-44 flex-none space-y-8 pt-32 sm:pt-0">
                    <div className="relative">
                      <img
                        src="/hero2.png"
                        alt=""
                        className="aspect-[2/3] w-full rounded-xl bg-neutral-900/5 object-cover shadow-lg"
                      />
                      <div className="pointer-events-none absolute inset-0 rounded-xl ring-1 ring-inset ring-neutral-900/10" />
                    </div>
                    <div className="relative">
                      <Suspense
                        fallback={
                          <img
                            src="/hero8.png"
                            alt=""
                            className="aspect-[2/3] w-full rounded-xl bg-neutral-900/5 object-cover shadow-lg"
                          />
                        }
                      >
                        <video
                          ref={videoRef}
                          src="/hero8.mp4"
                          autoPlay
                          muted
                          loop
                          playsInline
                          poster="/hero8.png"
                          className="aspect-[2/3] w-full rounded-xl bg-neutral-900/5 object-cover  shadow-lg"
                        />
                      </Suspense>

                      <div className="pointer-events-none absolute inset-0 rounded-xl ring-1 ring-inset ring-neutral-900/10" />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
