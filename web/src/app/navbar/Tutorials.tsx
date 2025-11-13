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

// import Link from 'next/link';

export default function Tutorial() {
  return (
    <a
      href="https://docs.sciol.ac.cn"
      target="_blank"
      rel="noopener noreferrer"
      className="text-sm font-medium text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100"
    >
      教程
    </a>
  );
}
