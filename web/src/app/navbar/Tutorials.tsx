// import {
//   BookmarkSquareIcon,
//   EnvelopeIcon,
//   PuzzlePieceIcon,
//   RectangleGroupIcon,
// } from '@heroicons/react/24/outline';

// import { GitHubIcon } from '@/assets/SocialIcons';
// import NavbarFullWidth from './NavbarFullWidth';
// import type { NavbarFullWidthProps } from './types';

// const resources = [
//   {
//     name: 'Overview',
//     description: 'Get review of our projects',
//     href: ``,
//     icon: BookmarkSquareIcon,
//   },
//   {
//     name: 'Workflow',
//     description: 'Learn from DeePMD-kit publications',
//     href: `/deepmd-kit`,
//     icon: RectangleGroupIcon,
//   },
//   {
//     name: 'Flociety',
//     description: 'Learn from DPGen publications',
//     href: `/deepmd-kit`,
//     icon: PuzzlePieceIcon,
//   },
// ];

// const callsToAction = [
//   {
//     name: 'Follow in Github',
//     href: 'https://github.com/Protium',
//     icon: GitHubIcon,
//   },
//   { name: 'Contact us', href: '#', icon: EnvelopeIcon },
// ];

// export default function Tutorial(
//   props: Omit<
//     NavbarFullWidthProps,
//     'solutions' | 'callsToAction' | 'buttonName'
//   >
// ) {
//   return (
//     <NavbarFullWidth
//       buttonName="tutorial"
//       solutions={resources}
//       callsToAction={callsToAction}
//       numberOfCallsToAction={2}
//       numberOfSolutions={3}
//       {...props}
//     />
//   );
// }

'use client';

// import Link from 'next/link'
import { useTranslation } from "react-i18next";

export default function Tutorial() {
  const { t } = useTranslation();


  return (
    <a
      href="https://docs.sciol.ac.cn"
      target="_blank"
      rel="noopener noreferrer"
      className="inline-flex items-center gap-x-1 text-sm font-semibold leading-6 text-neutral-900 focus:outline-none dark:text-neutral-100 hover:text-indigo-600 dark:hover:text-indigo-500"
      >
      {t('navbar.tutorial','Tutorial')}
      </a>
  );
}
