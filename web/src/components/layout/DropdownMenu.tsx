import { useAuthStore } from '@/store/authStore';
import { Menu, MenuItem, MenuItems, Transition } from '@headlessui/react';
import {
  BookOpenIcon,
  Cog6ToothIcon,
  LifebuoyIcon,
  PowerIcon,
  SparklesIcon,
  UserCircleIcon,
} from '@heroicons/react/24/outline';
import { useXyzen } from '@sciol/xyzen';
import clsx from 'clsx';

import LangSwitch from '@/components/feature/LangSwitch';
import { ThemeToggle } from '@/components/feature/ThemeToggle';
import XyzenButton from '@/components/feature/XyzenButton';
import { Fragment, useState } from 'react';
import { useTranslation } from 'react-i18next';
// Removed invalid internal type import from headlessui; rely on library types

function classNames(...classes: string[]) {
  if (classes === undefined) {
    return '';
  }
  return classes.filter(Boolean).join(' ');
}

interface DropdownMenuProps<T> {
  customNavi?: T[];
  children: React.ReactNode;
  avatar: string;
  username: string;
  onOpenChange?: (isOpen: boolean) => void;
}

interface NavigationProps {
  name: string;
  href?: string;
  onClick?: () => void;
  class?: 'primary' | 'secondary' | 'third';
  icon?: React.ReactNode;
}

export function DropdownMenu({
  customNavi,
  children,
  avatar,
  username,
  onOpenChange,
}: DropdownMenuProps<NavigationProps>) {
  const logout = useAuthStore((s) => s.logout);
  const { t } = useTranslation('userPanel');
  const [isOpen, setIsOpen] = useState(false);
  const { isXyzenOpen } = useXyzen();
  const navigations: NavigationProps[] = [
    {
      name: 'Profile',
      href: `/settings`,
      class: 'primary',
      icon: <UserCircleIcon />,
    },
    {
      name: 'Settings',
      href: `/settings`,
      class: 'primary',
      icon: <Cog6ToothIcon />,
    },
    {
      name: 'Changelog',
      href: `https://github.com/ScienceOL/PROTIUM/releases`,
      class: 'secondary',
      icon: <BookOpenIcon />,
    },
    {
      name: 'Support',
      href: `#`,
      class: 'secondary',
      icon: <LifebuoyIcon />,
    },
    {
      name: 'Signout',
      onClick: () => logout(),
      class: 'third',
      icon: <PowerIcon />,
    },
  ];

  const navigation: NavigationProps[] = customNavi ?? navigations;

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open);
    if (onOpenChange) {
      onOpenChange(open);
    }
  };

  return (
    <Menu as="div" className={'flex'}>
      {({ open }) => {
        // Call the handler when open state changes
        if (open !== isOpen) {
          handleOpenChange(open);
        }
        return (
          <>
            {children}
            <Transition
              as={Fragment}
              enter="transition ease-in-out duration-300"
              enterFrom="transform opacity-0 scale-70 origin-top-left"
              enterTo="transform opacity-100 scale-100 origin-top-left"
              leave="transition ease-in duration-150"
              leaveFrom="transform opacity-100 scale-100 origin-top-left"
              leaveTo="transform opacity-0 scale-70 origin-top-left"
              afterLeave={() => handleOpenChange(false)}
            >
              <MenuItems
                anchor={{
                  // Let TS infer types from headlessui, avoid internal type imports
                  to: 'top start',
                  gap: -3,
                  padding: -2,
                }}
                className="z-50 w-72 overflow-hidden rounded-xl bg-white shadow-lg
                  ring-1 ring-black ring-opacity-5 focus:outline-none
                  dark:border dark:border-neutral-700 dark:bg-neutral-800 dark:shadow-2xl"
              >
                {/* User Profile Section */}
                <div className="px-4 py-3">
                  <MenuItem
                    as="div"
                    className={'flex items-center justify-between'}
                  >
                    <div className="flex items-center">
                      <img
                        className="h-12 w-12 rounded-full border-2 border-white object-cover shadow-sm dark:border-neutral-700"
                        src={avatar}
                        alt="User Avatar"
                      />
                      <div className="ml-3">
                        <p className="text-base font-semibold text-neutral-900 dark:text-white">
                          {username}
                        </p>
                        <p className="text-xs text-neutral-500 dark:text-neutral-400">
                          Community
                        </p>
                      </div>
                    </div>

                    <XyzenButton>
                      <span className="font-display text-sm font-semibold transition-none">
                        Xyzen
                      </span>
                      <SparklesIcon
                        className={clsx(
                          'ml-1 h-4 w-4 group-hover:text-white',
                          isXyzenOpen
                            ? 'text-white dark:text-white'
                            : 'text-fuchsia-600 dark:text-fuchsia-400'
                        )}
                      />
                    </XyzenButton>
                  </MenuItem>
                </div>

                {/* Primary Navigation Section */}
                <div className="border-t border-neutral-200 dark:border-neutral-700">
                  <div className="px-1 py-2">
                    {navigation
                      .filter((item) => item.class === 'primary')
                      .map((item) => (
                        <MenuItem key={item.name}>
                          {({ active }) => (
                            <a
                              href={item.href || '#'}
                              onClick={item.onClick}
                              className={classNames(
                                active
                                  ? 'bg-neutral-100 dark:bg-neutral-700'
                                  : '',
                                'flex items-center rounded-md px-4 py-2.5 text-sm font-medium text-neutral-800 transition-colors dark:text-white'
                              )}
                            >
                              <div className="mr-3 h-5 w-5 text-neutral-500 dark:text-neutral-400">
                                {item.icon}
                              </div>
                              {t(item.name)}
                            </a>
                          )}
                        </MenuItem>
                      ))}
                  </div>
                </div>

                {/* Secondary Navigation Section */}
                <div className="border-t border-neutral-200 dark:border-neutral-700">
                  <div className="px-1 py-2">
                    {navigation
                      .filter((item) => item.class === 'secondary')
                      .map((item) => (
                        <MenuItem key={item.name}>
                          {({ active }) => (
                            <a
                              href={item.href || '#'}
                              onClick={item.onClick}
                              className={classNames(
                                active
                                  ? 'bg-neutral-100 dark:bg-neutral-700'
                                  : '',
                                'flex items-center rounded-md px-4 py-2.5 text-sm font-medium text-neutral-800 transition-colors dark:text-white'
                              )}
                            >
                              <div className="mr-3 h-5 w-5 text-neutral-500 dark:text-neutral-400">
                                {item.icon}
                              </div>
                              {t(item.name)}
                            </a>
                          )}
                        </MenuItem>
                      ))}
                  </div>
                </div>

                {/* Bottom Controls Section */}
                <div className="border-t border-neutral-200 dark:border-neutral-700">
                  <div className="p-3">
                    {/* Logout and Theme/Language Controls on same line */}
                    <div className="flex items-center justify-between">
                      {/* Logout Button on the left */}
                      <button
                        onClick={() => logout()}
                        className="mr-3 flex flex-1 items-center rounded-md px-3 py-2 text-sm font-medium text-red-600 transition-colors hover:bg-red-50 dark:text-red-400 dark:hover:bg-neutral-700"
                      >
                        <div className="mr-2 h-5 w-5 text-red-600 dark:text-red-400">
                          <PowerIcon />
                        </div>
                        {t('Sign out')}
                      </button>

                      {/* Theme and Language Controls on the right */}
                      <div className="flex items-center space-x-3">
                        <LangSwitch className="h-4 w-4" />
                        <ThemeToggle className="h-4 w-4" />
                      </div>
                    </div>
                  </div>
                </div>
              </MenuItems>
            </Transition>
          </>
        );
      }}
    </Menu>
  );
}

export default DropdownMenu;
