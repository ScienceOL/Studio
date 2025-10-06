'use client';
import { RightSideStatus } from '@/components/navbar/RightSideStatus';

import clsx from 'clsx';
import { useScroll, useTransform } from 'framer-motion';
import a from 'next/link';

import { useEffect, useRef } from 'react';
import Logo from '../../@brand/Logo';

// function MobileNavItem({
//   href,
//   children,
// }: {
//   href: string;
//   children: React.ReactNode;
// }) {
//   return (
//     <li>
//       <Popover.Button as={a} href={href} className="block py-2">
//         {children}
//       </Popover.Button>
//     </li>
//   );
// }

const NavbarMenu = (props: React.ComponentPropsWithoutRef<'nav'>) => {
  return (
    <nav {...props}>
      {/* <About /> */}
      {/* <News />
      <Projects />
      <Research /> */}
      {/* <Community /> */}
    </nav>
  );
};

// function MobileNavigation(
//   props: React.ComponentPropsWithoutRef<typeof Popover>,
// ) {
//   return (
//     <Popover {...props}>
//       <Popover.Button className="group flex items-center rounded-full bg-white/90 px-4 py-2 text-sm font-medium text-zinc-800 shadow-lg shadow-zinc-800/5 ring-1 ring-zinc-900/5 backdrop-blur dark:bg-zinc-800/90 dark:text-zinc-200 dark:ring-white/10 dark:hover:ring-white/20">
//         Menu
//         <ChevronDownIcon className="ml-3 h-auto w-2 stroke-zinc-500 group-hover:stroke-zinc-700 dark:group-hover:stroke-zinc-400" />
//       </Popover.Button>
//       <Transition.Root>
//         <Transition.Child
//           as={Fragment}
//           enter="duration-150 ease-out"
//           enterFrom="opacity-0"
//           enterTo="opacity-100"
//           leave="duration-150 ease-in"
//           leaveFrom="opacity-100"
//           leaveTo="opacity-0"
//         >
//           <Popover.Overlay className="fixed inset-0 z-50 bg-zinc-800/40 backdrop-blur-sm dark:bg-black/80" />
//         </Transition.Child>
//         <Transition.Child
//           as={Fragment}
//           enter="duration-150 ease-out"
//           enterFrom="opacity-0 scale-95"
//           enterTo="opacity-100 scale-100"
//           leave="duration-150 ease-in"
//           leaveFrom="opacity-100 scale-100"
//           leaveTo="opacity-0 scale-95"
//         >
//           <Popover.Panel
//             focus
//             className="fixed inset-x-4 top-8 z-50 origin-top rounded-3xl bg-white p-8 ring-1 ring-zinc-900/5 dark:bg-zinc-900 dark:ring-zinc-800"
//           >
//             <div className="flex flex-row-reverse items-center justify-between">
//               <Popover.Button aria-label="Close menu" className="-m-1 p-1">
//                 <CloseIcon className="h-6 w-6 text-zinc-500 dark:text-zinc-400" />
//               </Popover.Button>
//               <h2 className="text-sm font-medium text-zinc-600 dark:text-zinc-400">
//                 Navigation
//               </h2>
//             </div>
//             <nav className="mt-6">
//               <ul className="-my-2 divide-y divide-zinc-100 text-base text-zinc-800 dark:divide-zinc-100/5 dark:text-zinc-300">
//                 <MobileNavItem href="/about">About</MobileNavItem>
//                 <MobileNavItem href="/articles">Articles</MobileNavItem>
//                 <MobileNavItem href="/projects">Projects</MobileNavItem>
//                 <MobileNavItem href="/speaking">Speaking</MobileNavItem>
//                 <MobileNavItem href="/uses">Uses</MobileNavItem>
//               </ul>
//             </nav>
//           </Popover.Panel>
//         </Transition.Child>
//       </Transition.Root>
//     </Popover>
//   );
// }

function NavItem({
  href,
  children,
}: {
  href: string;
  children: React.ReactNode;
}) {
  let isActive = usePathname() === href;

  return (
    <li>
      <a
        href={href}
        className={clsx(
          'relative block px-3 py-2 transition',
          isActive
            ? 'text-indigo-500 dark:text-indigo-400'
            : 'hover:text-indigo-500 dark:hover:text-indigo-400'
        )}
      >
        {children}
        {isActive && (
          <span className="absolute inset-x-1 -bottom-px h-px bg-gradient-to-r from-indigo-500/0 via-indigo-500/40 to-indigo-500/0 dark:from-indigo-400/0 dark:via-indigo-400/40 dark:to-indigo-400/0" />
        )}
      </a>
    </li>
  );
}

function DesktopNavigation(props: React.ComponentPropsWithoutRef<'nav'>) {
  return (
    <nav {...props}>
      <ul className="flex rounded-full bg-white/90 px-3 text-sm font-medium text-zinc-800 shadow-lg shadow-zinc-800/5 ring-1 ring-zinc-900/5 backdrop-blur dark:bg-zinc-800/90 dark:text-zinc-200 dark:ring-white/10">
        {/* <NavItem href="/about">About</NavItem> */}
        <NavItem href="/articles">Articles</NavItem>
        {/* <NavItem href="/projects">Projects</NavItem> */}
        {/* <NavItem href="/speaking">Speaking</NavItem> */}
        {/* <NavItem href="/uses">Uses</NavItem> */}
      </ul>
    </nav>
  );
}

function clamp(number: number, a: number, b: number) {
  let min = Math.min(a, b);
  let max = Math.max(a, b);
  return Math.min(Math.max(number, min), max);
}

function AvatarContainer({
  className,
  ...props
}: React.ComponentPropsWithoutRef<'div'>) {
  return (
    <div
      className={clsx(className, 'h-9 w-9 rounded-full  bg-transparent')}
      {...props}
    />
  );
}

function Avatar({
  large = false,
  className,
  ...props
}: Omit<React.ComponentPropsWithoutRef<typeof a>, 'href'> & {
  large?: boolean;
}) {
  return (
    <a
      href="/"
      aria-label="Home"
      className={clsx(
        className,
        'pointer-events-auto h-8 w-auto transition-opacity duration-150 ease-in-out hover:opacity-50'
      )}
      {...props}
    >
      <Logo
        className={clsx('object-cover ', large ? 'h-16 w-16' : 'h-8 w-8')}
      />
      {/* <Image
        src={}
        alt=""
        sizes={large ? '4rem' : '2.25rem'}
        className={clsx(
          'rounded-full object-cover ',
          large ? 'h-16 w-16' : 'h-8 w-8',
        )}
        priority
      /> */}
    </a>
  );
}

export default function Header() {
  let isHomePage = usePathname() === '/';

  let headerRef = useRef<React.ElementRef<'div'>>(null);
  let avatarRef = useRef<React.ElementRef<'div'>>(null);
  let isInitial = useRef(true);

  useEffect(() => {
    let downDelay = avatarRef.current?.offsetTop ?? 0;
    let upDelay = 64;

    function setProperty(property: string, value: string) {
      document.documentElement.style.setProperty(property, value);
    }

    function removeProperty(property: string) {
      document.documentElement.style.removeProperty(property);
    }

    function updateHeaderStyles() {
      if (!headerRef.current) {
        return;
      }

      let { top, height } = headerRef.current.getBoundingClientRect();
      let scrollY = clamp(
        window.scrollY,
        0,
        document.body.scrollHeight - window.innerHeight
      );

      if (isInitial.current) {
        setProperty('--header-position', 'sticky');
      }

      setProperty('--content-offset', `${downDelay}px`);

      if (isInitial.current || scrollY < downDelay) {
        setProperty('--header-height', `${downDelay + height}px`);
        setProperty('--header-mb', `${-downDelay}px`);
      } else if (top + height < -upDelay) {
        let offset = Math.max(height, scrollY - upDelay);
        setProperty('--header-height', `${offset}px`);
        setProperty('--header-mb', `${height - offset}px`);
      } else if (top === 0) {
        setProperty('--header-height', `${scrollY + height}px`);
        setProperty('--header-mb', `${-scrollY}px`);
      }

      if (top === 0 && scrollY > 0 && scrollY >= downDelay) {
        setProperty('--header-inner-position', 'fixed');
        removeProperty('--header-top');
        removeProperty('--avatar-top');
      } else {
        removeProperty('--header-inner-position');
        setProperty('--header-top', '0px');
        setProperty('--avatar-top', '0px');
      }
    }

    function updateAvatarStyles() {
      if (!isHomePage) {
        return;
      }

      let fromScale = 1;
      let toScale = 36 / 64;
      let fromX = 0;
      let toX = 2 / 16;

      let scrollY = downDelay - window.scrollY;

      let scale = (scrollY * (fromScale - toScale)) / downDelay + toScale;
      scale = clamp(scale, fromScale, toScale);

      let x = (scrollY * (fromX - toX)) / downDelay + toX;
      x = clamp(x, fromX, toX);

      setProperty(
        '--avatar-image-transform',
        `translate3d(${x}rem, 0, 0) scale(${scale})`
      );

      let borderScale = 1 / (toScale / scale);
      let borderX = (-toX + x) * borderScale;
      let borderTransform = `translate3d(${borderX}rem, 0, 0) scale(${borderScale})`;

      setProperty('--avatar-border-transform', borderTransform);
      setProperty('--avatar-border-opacity', scale === toScale ? '1' : '0');
    }

    function updateStyles() {
      updateHeaderStyles();
      updateAvatarStyles();
      isInitial.current = false;
    }

    updateStyles();
    window.addEventListener('scroll', updateStyles, { passive: true });
    window.addEventListener('resize', updateStyles);

    return () => {
      window.removeEventListener('scroll', updateStyles);
      window.removeEventListener('resize', updateStyles);
    };
  }, [isHomePage]);

  let { scrollY } = useScroll();
  let bgOpacityLight = useTransform(scrollY, [0, 72], [0.5, 0.9]);
  let bgOpacityDark = useTransform(scrollY, [0, 72], [0.2, 0.8]);

  return (
    <header className={isHomePage ? 'absolute inset-x-0 top-0 z-50' : ''}>
      <div
        className="pointer-events-none z-50 flex flex-none flex-col"
        style={{
          height: 'var(--header-height)',
          marginBottom: 'var(--header-mb)',
        }}
      >
        <div
          ref={headerRef}
          className="top-0 z-10 h-16 pt-6"
          style={{
            position:
              'var(--header-position)' as React.CSSProperties['position'],
          }}
        >
          <div
            className={clsx(
              'sm:px-8',
              'top-[var(--header-top,theme(spacing.6))] w-full'
            )}
            style={{
              position:
                'var(--header-inner-position)' as React.CSSProperties['position'],
            }}
          >
            <div
              className={clsx(
                'mx-auto ',
                isHomePage ? ' max-w-screen-2xl' : 'w-full max-w-7xl lg:px-8'
              )}
            >
              <div
                className={clsx(
                  'relative px-4 ',
                  isHomePage ? 'sm:px-0' : 'sm:px-8 lg:px-12'
                )}
              >
                <div
                  className={clsx(
                    'mx-auto',
                    isHomePage ? '' : ' max-w-2xl lg:max-w-5xl'
                  )}
                >
                  <div className="relative flex items-center gap-4">
                    <div className="flex flex-1 items-center justify-start gap-x-8">
                      <AvatarContainer>
                        <Avatar />
                      </AvatarContainer>
                      <div className="flex justify-start">
                        {isHomePage ? (
                          <NavbarMenu className=" pointer-events-auto hidden lg:flex lg:gap-x-9 " />
                        ) : (
                          <DesktopNavigation className="pointer-events-auto hidden lg:block" />
                        )}
                      </div>
                    </div>

                    <div className="flex justify-end md:flex-1">
                      <div className="pointer-events-auto flex items-center">
                        <RightSideStatus dropdownMenuPosition="mt-3" />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
