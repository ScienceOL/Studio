'use client';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import About from './About';
import Community from './Community';
import Projects from './Projects';
import { RightSideStatus } from './RightSideStatus';
// import Tutorial from './Tutorials';

// const Tutorial = Tutorial as React.ComponentType<{
//   index: number;
//   activeMenuItem: number | null;
//   setActiveMenuItem: React.Dispatch<React.SetStateAction<number | null>>;
//   open: boolean;
//   setOpen: React.Dispatch<React.SetStateAction<boolean>>;
// }>;

const NavbarMenu = () => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [activeMenuItem, setActiveMenuItem] = useState<number | null>(null);
  return (
    <div className="hidden lg:flex lg:items-center lg:space-x-4 lg:pl-6">
      <About
        index={1}
        activeMenuItem={activeMenuItem}
        setActiveMenuItem={setActiveMenuItem}
        open={open}
        setOpen={setOpen}
      />
      <Projects
        index={2}
        activeMenuItem={activeMenuItem}
        setActiveMenuItem={setActiveMenuItem}
        open={open}
        setOpen={setOpen}
      />
      <a
        className="inline-flex items-center gap-x-1 text-sm font-semibold leading-6 text-white/80 hover:text-white focus:outline-none"
        href="https://docs.sciol.ac.cn"
      >
        <span>{t('navbar.tutorial')}</span>
      </a>
      {/* <Tutorial
        index={4}
        activeMenuItem={activeMenuItem}
        setActiveMenuItem={setActiveMenuItem}
        open={open}
        setOpen={setOpen}
      /> */}
      <Community
        index={3}
        activeMenuItem={activeMenuItem}
        setActiveMenuItem={setActiveMenuItem}
        open={open}
        setOpen={setOpen}
      />
      <a
        className="inline-flex items-center gap-x-1 text-sm font-semibold leading-6 text-white/80 hover:text-white focus:outline-none"
        href="/certification"
      >
        <span>{t('navbar.certification')}</span>
      </a>
    </div>
  );
};

const Navbar = () => {
  return (
    <div className="relative flex w-full justify-center">
      <header className="absolute top-0 z-50 mx-auto w-full max-w-full">
        <nav
          className=" flex items-center justify-between p-6 lg:px-8"
          aria-label="Global"
        >
          <div className="flex">
            <a
              href="/"
              className="z-50 -m-1.5 rounded-md p-1.5 transition-opacity duration-150 ease-in-out hover:bg-neutral-300/20 dark:hover:bg-neutral-800"
            >
              <span className="sr-only">ScienceOL</span>
              <img
                src="https://storage.sciol.ac.cn/library/BLogo-dark.svg"
                alt="ScienceOL"
                className="h-8 w-8"
              />
            </a>
            <NavbarMenu />
          </div>

          <RightSideStatus dropdownMenuPosition="-mt-2" />
        </nav>
      </header>
    </div>
  );
};

export default Navbar;
