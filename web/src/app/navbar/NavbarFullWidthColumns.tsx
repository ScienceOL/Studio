import { Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/20/solid';
import clsx from 'clsx';
import { Fragment } from 'react';
import { useTranslation } from 'react-i18next';
import NavbarFullWidthFooter from './NavbarFullWidthFooter';
import type { NavbarFullWidthColumnsProps } from './types';

export default function NavbarFullWidthColumns({
  buttonName,
  resources,

  callsToAction,
  recentPosts,
  index,
  activeMenuItem,

  setActiveMenuItem,
  open,
  setOpen,
  numberOfCallsToAction = 3,
}: NavbarFullWidthColumnsProps) {
  const { t } = useTranslation();
  return (
    <div className="">
      <div
        className=" bg-transparent"
        onMouseEnter={() => {
          setOpen(true);
          setActiveMenuItem(index);
        }}
        onClick={() => {
          setOpen(!open);
        }}
      >
        <div className="mx-auto max-w-7xl px-2">
          <button
            name={buttonName}
            className={clsx(
              'inline-flex items-center  gap-x-1 text-sm font-semibold leading-6',
              ' group-hover:text-neutral-900  ',
              'transition-transform duration-300 ease-in-out',
              open && activeMenuItem === index
                ? 'scale-110 text-indigo-600 dark:text-indigo-400'
                : 'text-neutral-900 dark:text-white'
            )}
          >
            {t(`navbar.${buttonName}`)}
            <ChevronDownIcon
              className={clsx(
                'h-5 w-5',
                'transition-transform duration-300 ease-in-out',
                open && activeMenuItem === index
                  ? 'rotate-180 text-indigo-600 dark:text-indigo-400'
                  : 'text-neutral-600 dark:text-neutral-400'
              )}
              aria-hidden="true"
            />
          </button>
        </div>
      </div>
      {activeMenuItem === index && (
        <Transition
          show={open && activeMenuItem === index}
          as={Fragment}
          enter="transition ease-out duration-500"
          enterFrom="opacity-0 -translate-y-1"
          enterTo="opacity-100 translate-y-0"
          leave="transition ease-in duration-150"
          leaveFrom="opacity-100 translate-y-0"
          leaveTo="opacity-0 -translate-y-1"
        >
          <div
            // onMouseEnter={() => setOpen(true)}
            onMouseLeave={() => setOpen(false)}
            className=" absolute inset-x-0 top-0 -z-10 w-full bg-white/40 pt-16 shadow-lg ring-1
           ring-neutral-900/5 backdrop-blur-2xl dark:bg-black/40"
          >
            <div className="mx-auto flex  max-w-8xl gap-x-8 gap-y-10 px-6 py-10 lg:px-8 2xl:max-w-screen-2xl">
              <div className=" w-1/4 gap-x-6 sm:gap-x-8">
                <div>
                  <h3
                    className="text-sm font-medium leading-6
                   text-neutral-500 dark:text-neutral-300"
                  >
                    Resources
                  </h3>
                  <div className="mt-6 flow-root">
                    <div className="-my-2">
                      {resources?.map((item) => (
                        <a
                          key={item.name}
                          href={item.href}
                          className="group flex gap-x-4 py-2 text-sm font-semibold leading-6
                           text-neutral-900 hover:text-indigo-600 dark:text-white dark:hover:text-indigo-400"
                        >
                          <item.icon
                            className="h-6 w-6 flex-none
                             text-neutral-400 group-hover:text-indigo-600 dark:text-neutral-300 dark:group-hover:text-indigo-400  "
                            aria-hidden="true"
                          />
                          {item.name}
                        </a>
                      ))}
                    </div>
                  </div>
                </div>
                {/* <div>
                  <h3 className="text-sm font-medium leading-6 text-neutral-500 dark:text-neutral-300">
                    Engagement
                  </h3>
                  <div className="mt-6 flow-root">
                    <div className="-my-2">
                      {engagements?.map((item) => (
                        <a
                          key={item.name}
                          href={item.href}
                          className="group flex gap-x-4 py-2 text-sm font-semibold leading-6
                          text-neutral-900 hover:text-indigo-600 dark:text-white dark:hover:text-indigo-400"
                        >
                          <item.icon
                            className="h-6 w-6 flex-none
                            text-neutral-400 group-hover:text-indigo-600 dark:text-neutral-300 dark:group-hover:text-indigo-400  "
                            aria-hidden="true"
                          />
                          {item.name}
                        </a>
                      ))}
                    </div>
                  </div>
                </div> */}
              </div>
              <div className="grid grid-cols-1 gap-10 sm:gap-8 lg:grid-cols-3">
                <h3 className="sr-only">Recent posts</h3>
                {recentPosts?.map((post) => (
                  <article
                    key={post.id}
                    className="relative isolate -m-4 flex max-w-2xl flex-col gap-x-8 gap-y-6 rounded-md p-4 hover:bg-indigo-200/10 sm:flex-row sm:items-start lg:flex-col lg:items-stretch"
                  >
                    <div className="relative flex-none">
                      <img
                        className="aspect-[2/1] w-full rounded-lg bg-neutral-100 object-cover dark:bg-neutral-800 sm:aspect-[16/9] sm:h-32 lg:h-auto"
                        src={post.imageUrl}
                        alt=""
                      />
                      <div className="absolute inset-0 rounded-lg ring-1 ring-inset ring-neutral-900/10 dark:ring-neutral-50/10" />
                    </div>
                    <div>
                      <div className="flex items-center gap-x-4">
                        <time
                          dateTime={post.datetime}
                          className="text-sm leading-6 text-neutral-600 dark:text-neutral-400"
                        >
                          {post.date}
                        </time>
                        <a
                          href={post.category.href}
                          className="bg-neutralral-50 hovbg-neutraleutral-100bg-neutralg-neutral-900 dabg-neutralr:bg-neutral-800 relative z-10 rounded-full px-3 py-1.5 text-xs font-medium text-neutral-600 dark:text-neutral-400"
                        >
                          {post.category.title}
                        </a>
                      </div>
                      <h4 className="mt-2 text-sm font-semibold leading-6 text-neutral-900 dark:text-white">
                        <a href={post.href}>
                          <span className="absolute inset-0" />
                          {post.title}
                        </a>
                      </h4>
                      <p className="mt-2 line-clamp-3 text-sm leading-6 text-neutral-600 dark:text-neutral-400">
                        {post.description}
                      </p>
                    </div>
                  </article>
                ))}
              </div>
            </div>
            <NavbarFullWidthFooter
              callsToAction={callsToAction!}
              numberOfCallsToAction={numberOfCallsToAction}
            />
          </div>
        </Transition>
      )}
    </div>
  );
}
