import { type NavbarFullWidthPreviewProps } from '@/types/navbar';
import { Tab, Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/20/solid';
import clsx from 'clsx';

import { Fragment, useState } from 'react';
import { useTranslation } from 'react-i18next';
import NavbarFullWidthFooter from './NavbarFullWidthFooter';

export default function NavbarFullWidthPreview({
  buttonName,
  options,
  callsToAction,
  index,
  activeMenuItem,
  setActiveMenuItem,
  open,
  setOpen,
  numberOfCallsToAction = 3,
  numberOfSolutions = 3,
}: NavbarFullWidthPreviewProps) {
  const { t } = useTranslation();

  const [categories] = useState(options);
  const [selectedIndex, setSelectedIndex] = useState(0);
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
              'inline-flex items-center gap-x-1 text-sm font-semibold leading-6',
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
            className={clsx(
              ' absolute inset-x-0 top-0 -z-10 w-full bg-white/40 pt-16 shadow-lg ring-1',
              'ring-neutral-900/5 backdrop-blur-3xl dark:bg-black/40'
            )}
          >
            <Tab.Group
              selectedIndex={selectedIndex}
              onChange={setSelectedIndex}
              as="div"
              className="flex w-full flex-row items-start justify-between space-x-8 px-8 py-6 sm:py-10"
            >
              <Tab.List className="flex w-1/5 flex-col items-center justify-start space-y-4">
                {Object.keys(categories).map((category, idx) => (
                  <Tab
                    onMouseEnter={() => {
                      setSelectedIndex(idx);
                    }}
                    key={category}
                    className={({ selected }) =>
                      clsx(
                        'w-full rounded-lg py-2.5 text-sm font-medium leading-5',
                        'focus:outline-none',
                        selected
                          ? 'bg-neutral-100/80 text-indigo-700 shadow dark:bg-neutral-800/80 dark:text-white'
                          : 'text-neutral-800 hover:text-indigo-600 dark:text-indigo-100  dark:hover:text-white'
                      )
                    }
                  >
                    <span className="text-sm">{category}</span>
                  </Tab>
                ))}
              </Tab.List>
              <Tab.Panels className="flex w-full flex-1 justify-start">
                {Object.values(categories).map((posts, idx) => (
                  <Tab.Panel
                    key={idx}
                    className={clsx(
                      'grid w-full grid-cols-1 gap-2 px-6 ',
                      'sm:grid-cols-2 sm:gap-x-6 sm:gap-y-0 ',
                      `lg:grid-cols-3 lg:gap-4 lg:px-8 xl:gap-8`
                    )}
                    // className={clsx(
                    //   'w-full rounded-xl bg-transparent',
                    //   'ring-white/60 ring-offset-2 ring-offset-blue-400 focus:outline-none focus:ring-2',
                    // )}
                  >
                    {posts.map((item) => (
                      <div
                        key={item.name}
                        className="group relative -mx-3 flex gap-6 rounded-lg p-3 text-sm leading-6
                   hover:bg-neutral-50/40 hover:shadow hover:ring-2 hover:ring-inset hover:ring-indigo-600/2.5
                   dark:hover:bg-neutral-800/40 dark:hover:ring-indigo-300/2.5
                    sm:flex-col sm:p-6"
                      >
                        {/* <div
                          className="flex h-11 w-11 flex-none items-center justify-center rounded-lg
                   bg-neutral-50/40 group-hover:bg-white
                   dark:bg-neutral-900/40 dark:group-hover:bg-black"
                        >

                        </div> */}
                        <div>
                          <a
                            href={item.href}
                            className="flex items-center gap-4 font-semibold text-neutral-900 dark:text-white"
                          >
                            <item.icon
                              className="h-6 w-6 text-neutral-600 group-hover:text-indigo-600
                       dark:text-neutral-300 dark:group-hover:text-white"
                              aria-hidden="true"
                            />
                            {item.name}
                            <span className="absolute inset-0" />
                          </a>
                          <p className="mt-1 line-clamp-3 text-neutral-700 dark:text-neutral-200">
                            {item.description}
                          </p>
                          {item.name === 'DeePMD-kit' ? (
                            <span className="isolate z-[60] mt-4 inline-flex gap-x-4 rounded-md ">
                              <button
                                type="button"
                                className="rounded bg-indigo-50 px-2 py-1 text-xs font-semibold text-indigo-600 shadow-sm transition-all
                              duration-300 ease-in-out hover:bg-indigo-100 dark:bg-white/10
                              dark:text-indigo-100 dark:hover:bg-indigo-800 dark:hover:text-indigo-50
                              "
                              >
                                Docs
                              </button>
                              <button
                                type="button"
                                className="rounded bg-indigo-50 px-2 py-1 text-xs font-semibold text-indigo-600 shadow-sm transition-all
                              duration-300 ease-in-out hover:bg-indigo-100 dark:bg-white/10
                              dark:text-indigo-100 dark:hover:bg-indigo-800 dark:hover:text-indigo-50
                              "
                              >
                                Publications
                              </button>
                            </span>
                          ) : (
                            <span className="isolate z-[60] mt-4 inline-flex gap-x-4 rounded-md group-hover:animate-bounce-x group-hover:text-indigo-600 dark:group-hover:text-white ">
                              →
                            </span>
                          )}
                        </div>
                      </div>
                    ))}
                  </Tab.Panel>
                ))}
              </Tab.Panels>
            </Tab.Group>
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
