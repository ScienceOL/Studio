import { type NavbarFullWidthProps } from '@/types/navbar';
import { Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/20/solid';
import clsx from 'clsx';

import { Fragment } from 'react';
import { useTranslation } from 'react-i18next';
import NavbarFullWidthFooter from './NavbarFullWidthFooter';

export default function NavbarFullWidth({
  buttonName,
  solutions,
  callsToAction,
  index,
  activeMenuItem,
  setActiveMenuItem,
  open,
  setOpen,
  numberOfCallsToAction = 3,
  numberOfSolutions = 4,
}: NavbarFullWidthProps) {
  const { t } = useTranslation();
  return (
    <div className="">
      <div
        className="bg-transparent"
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
            <div
              className={clsx(
                'mx-auto grid max-w-7xl grid-cols-1 gap-2 px-6 py-6',
                'sm:grid-cols-2 sm:gap-x-6 sm:gap-y-0 sm:py-10 ',
                `lg:grid-cols-${numberOfSolutions} lg:gap-4 md:px-8 xl:gap-8`
              )}
            >
              {solutions?.map((item) => (
                <div
                  key={item.name}
                  className="group relative -mx-3 flex gap-6 rounded-lg p-3 text-sm leading-6
                   hover:bg-neutral-50/40 hover:shadow hover:ring-2 hover:ring-inset hover:ring-indigo-600/2.5
                   dark:hover:bg-neutral-800/40 dark:hover:ring-indigo-300/2.5
                    sm:flex-col sm:p-6"
                >
                  <div
                    className="flex h-11 w-11 flex-none items-center justify-center rounded-lg
                   bg-neutral-50/40 group-hover:bg-white
                   dark:bg-neutral-900/40 dark:group-hover:bg-black"
                  >
                    <item.icon
                      className="h-6 w-6 text-neutral-600 group-hover:text-indigo-600
                       dark:text-neutral-300 dark:group-hover:text-white"
                      aria-hidden="true"
                    />
                  </div>
                  <div>
                    <a
                      href={item.href}
                      className="font-semibold text-neutral-900 dark:text-white"
                    >
                      {item.name}
                      <span className="absolute inset-0" />
                    </a>
                    <p className="mt-1 text-neutral-600 dark:text-neutral-400">
                      {item.description}
                    </p>
                  </div>
                </div>
              ))}
            </div>
            <NavbarFullWidthFooter
              callsToAction={callsToAction}
              numberOfCallsToAction={numberOfCallsToAction}
            />
          </div>
        </Transition>
      )}
    </div>
  );
}
