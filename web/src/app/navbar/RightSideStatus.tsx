'use client';
import { Dialog, Transition } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';

import { Fragment } from 'react';

import Logo from '@/assets/Logo';

import LangSwitch from '@/components/feature/LangSwitch';
import { ThemeToggle } from '@/components/feature/ThemeToggle';
import { useAuth } from '@/hooks/useAuth';
import { Bars3Icon } from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';

interface RightSideStatusProps {
  isMobile?: boolean;
  dropdownMenuPosition?: string;
}

const navigation = [
  { name: 'About', href: '/about' },
  { name: 'Manifest', href: '/manifesto' },
  { name: 'Project', href: '/space/DeePMD-kit' },
  // { name: 'Research', href: '#' },
  { name: 'Articles', href: '/articles' },
];

export const RightSideStatus: React.FC<RightSideStatusProps> = () => {
  const {
    isAuthenticated: isLogged,
    user: userInfo,
    checkAuthStatus,
  } = useAuth();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  useEffect(() => {
    checkAuthStatus();
  }, [checkAuthStatus]);

  return (
    <div className="w-fit">
      {/* 桌面端 */}
      <div className="hidden items-center gap-x-3 lg:flex">
        <LangSwitch />
        <ThemeToggle />
        {isLogged ? (
          <img
            className="bg-neutral-5 h-8 w-8 rounded-full"
            src={userInfo?.avatar || ''}
            alt="avatar"
          />
        ) : (
          <a
            href="/login"
            className="text-sm font-semibold leading-6 text-neutral-900 opacity-75 hover:text-indigo-600 hover:opacity-100 dark:text-neutral-100 dark:hover:text-white dark:hover:opacity-100"
          >
            Log in <span aria-hidden="true">&rarr;</span>
          </a>
        )}
      </div>

      {/* 移动端 */}
      <div className="flex lg:hidden">
        {isLogged && (
          <img
            className="bg-neutral-5 h-7 w-7 rounded-full"
            src={userInfo?.avatar || ''}
            alt="avatar"
          />
        )}
        <button
          type="button"
          className="-m-2.5 ml-2 inline-flex items-center justify-center rounded-md p-2.5 text-neutral-700 dark:text-neutral-200"
          onClick={() => setMobileMenuOpen(true)}
        >
          <span className="sr-only">Open main menu</span>
          <Bars3Icon
            className="h-7 w-7 dark:text-neutral-50"
            aria-hidden="true"
          />
        </button>
      </div>

      <Transition.Root show={mobileMenuOpen} as={Fragment}>
        <Dialog as="div" className="lg:hidden" onClose={setMobileMenuOpen}>
          <Transition.Child
            as={Fragment}
            enter="transition-opacity ease-out duration-200"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="transition-opacity ease-out duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 z-40 bg-neutral-950/40 backdrop-blur-sm" />
          </Transition.Child>

          <Transition.Child
            as={Fragment}
            enter="transform transition ease-out duration-300 sm:duration-500"
            enterFrom="translate-x-6 opacity-0"
            enterTo="translate-x-0 opacity-100"
            leave="transform transition ease-out duration-300 sm:duration-500"
            leaveFrom="translate-x-0 opacity-100"
            leaveTo="translate-x-6 opacity-0"
          >
            <Dialog.Panel className="fixed inset-0 z-50 flex flex-col overflow-y-auto bg-white px-6 py-6 shadow-xl ring-neutral-900/10 dark:bg-neutral-900 dark:ring-neutral-50/10">
              <div className="flex items-center justify-between">
                <a
                  href="/"
                  className="-m-1.5 rounded-md p-1.5"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  <span className="sr-only">Protium</span>
                  <Logo className="h-8 w-8 fill-indigo-800 dark:fill-white" />
                </a>
                <div className="flex items-center">
                  <LangSwitch className="mr-1.5" />
                  <ThemeToggle className="mr-1.5" />
                  <button
                    type="button"
                    className="-m-2.5 rounded-md p-2.5 text-neutral-700 transition hover:text-neutral-900 focus:outline-none focus:ring-2 focus:ring-indigo-500/40 dark:text-neutral-200 dark:hover:text-neutral-50"
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    <span className="sr-only">Close menu</span>
                    <XMarkIcon
                      className="h-6 w-6 dark:text-neutral-50"
                      aria-hidden="true"
                    />
                  </button>
                </div>
              </div>
              <div className="mt-8 flex-1 overflow-y-auto">
                <nav className="flex flex-col gap-4">
                  {navigation.map((item) => (
                    <a
                      key={item.name}
                      href={item.href}
                      onClick={() => setMobileMenuOpen(false)}
                      className="rounded-2xl border border-neutral-200/0 bg-neutral-100/0 px-4 py-4 text-lg font-semibold leading-7 text-neutral-900 transition hover:border-neutral-200/80 hover:bg-neutral-50 hover:text-indigo-600 dark:border-neutral-700/0 dark:text-neutral-50 dark:hover:border-neutral-700 dark:hover:bg-neutral-800/60 dark:hover:text-indigo-400"
                    >
                      {item.name}
                    </a>
                  ))}
                </nav>
              </div>
              <div className="mt-6 border-t border-neutral-200 pt-6 dark:border-neutral-800">
                {isLogged ? (
                  <div className="flex items-center gap-3 rounded-2xl bg-neutral-50 px-4 py-3 dark:bg-neutral-800/60">
                    <img
                      className="h-10 w-10 rounded-full object-cover"
                      src={userInfo?.avatar || ''}
                      alt="avatar"
                    />
                    <div className="text-sm font-semibold text-neutral-900 dark:text-neutral-100">
                      {userInfo?.displayName || userInfo?.name || 'Explorer'}
                    </div>
                  </div>
                ) : (
                  <a
                    href="/login"
                    onClick={() => setMobileMenuOpen(false)}
                    className="flex items-center justify-center rounded-xl bg-indigo-600 px-4 py-3 text-base font-semibold text-white transition hover:bg-indigo-500"
                  >
                    Log in
                  </a>
                )}
              </div>
            </Dialog.Panel>
          </Transition.Child>
        </Dialog>
      </Transition.Root>
    </div>
  );
};
