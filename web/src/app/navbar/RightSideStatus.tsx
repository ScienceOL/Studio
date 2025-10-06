'use client';
import { Dialog, Transition } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';

import { Fragment } from 'react';

import Logo from '@/assets/Logo';

import { ThemeToggle } from '@/components/feature/ThemeToggle';
import { useAuthStore } from '@/store/authStore';
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
  const isLogged = useAuthStore((s) => s.isAuthenticated);
  const userInfo = useAuthStore((s) => s.user);
  const checkAuthStatus = useAuthStore((s) => s.checkAuthStatus);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  useEffect(() => {
    checkAuthStatus();
  }, [checkAuthStatus]);

  return (
    <div className="w-fit">
      {/* 桌面端 */}
      <div className="hidden items-center gap-x-3 lg:flex">
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
          <div className="fixed inset-0 z-50" />

          <Transition.Child
            as={Fragment}
            enter="transform transition ease-out duration-300 sm:duration-500"
            enterFrom="translate-x-6 opacity-0"
            enterTo="translate-x-0 opacity-100"
            leave="transform transition ease-out duration-300 sm:duration-500"
            leaveFrom="translate-x-0 opacity-100"
            leaveTo="translate-x-6 opacity-0"
          >
            <Dialog.Panel className="fixed inset-y-0 right-0 z-50 w-full overflow-y-auto bg-white px-6 py-6 ring-neutral-900/10 dark:bg-neutral-900 dark:ring-neutral-50/10 sm:max-w-sm sm:ring-1">
              <div className="flex items-center justify-between">
                <a href="/" className="-m-1.5 p-1.5">
                  <span className="sr-only">Protium</span>
                  <Logo className="h-8 w-8 fill-indigo-800 dark:fill-white" />
                </a>
                <div>
                  <ThemeToggle className="mr-2" />
                  <button
                    type="button"
                    className="-m-2.5 rounded-md p-2.5 text-neutral-700 dark:text-neutral-200"
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
              <div className="mt-6 flow-root">
                <div className="-my-6 divide-y divide-neutral-500/10 dark:divide-neutral-500/70">
                  <div className="space-y-2 py-6">
                    {navigation.map((item) => (
                      <a
                        key={item.name}
                        href={item.href}
                        className="-mx-3 block rounded-lg px-3 py-2 text-base font-semibold leading-7 text-neutral-900 hover:bg-neutral-50 dark:text-neutral-50 dark:hover:bg-neutral-800"
                      >
                        {item.name}
                      </a>
                    ))}
                  </div>
                  <div className="py-6">
                    {!isLogged && (
                      <a
                        href="/login"
                        className="-mx-3 block rounded-lg px-3 py-2.5 text-base font-semibold leading-7 text-neutral-900 hover:bg-neutral-50 dark:text-neutral-50 dark:hover:bg-neutral-800"
                      >
                        Log in
                      </a>
                    )}
                  </div>
                </div>
              </div>
            </Dialog.Panel>
          </Transition.Child>
        </Dialog>
      </Transition.Root>
    </div>
  );
};
