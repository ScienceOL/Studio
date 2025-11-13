// import { GitHubIcon } from '@/assets/SocialIcons';
// import { EnvelopeIcon, RocketLaunchIcon } from '@heroicons/react/24/outline';

// import {
//   ABACUSIcon,
//   AISSquareIcon,
//   DeePMDIcon,
//   DPGenIcon,
// } from '@/assets/Icons';
// import Logo from '@/assets/Logo';
// import NavbarFullWidthPreview from './NavbarFullWidthPreview';
// import type { NavbarFullWidthPreviewProps } from './types';

// const options = {
//   // 'Molecule Dynamics': [
//   //   {
//   //     name: 'DeePMD-kit',
//   //     description: 'Learn how to use DeePMD-kit',
//   //     href: `/space/DeePMD-kit`,
//   //     icon: DeePMDIcon,
//   //   },
//   //   {
//   //     name: 'DPGen',
//   //     description: 'Learn from DPGen publications',
//   //     href: `/dp-gen`,
//   //     icon: DPGenIcon,
//   //   },
//   //   {
//   //     name: 'DeePMD-kit',
//   //     description: 'Learn how to use DeePMD-kit',
//   //     href: `/DeePMD-kit`,
//   //     icon: DeePMDIcon,
//   //   },
//   //   {
//   //     name: 'DPGen',
//   //     description: 'Learn from DPGen publications',
//   //     href: `/DeePMD-kit`,
//   //     icon: DPGenIcon,
//   //   },
//   //   {
//   //     name: 'DeePMD-kit',
//   //     description: 'Learn how to use DeePMD-kit',
//   //     href: `/DeePMD-kit`,
//   //     icon: DeePMDIcon,
//   //   },
//   //   {
//   //     name: 'DPGen',
//   //     description: 'Learn from DPGen publications',
//   //     href: `/DeePMD-kit`,
//   //     icon: DPGenIcon,
//   //   },
//   //   {
//   //     name: 'DeePMD-kit',
//   //     description: 'Learn how to use DeePMD-kit',
//   //     href: `/DeePMD-kit`,
//   //     icon: DeePMDIcon,
//   //   },
//   //   {
//   //     name: 'DPGen',
//   //     description: 'Learn from DPGen publications',
//   //     href: `/DeePMD-kit`,
//   //     icon: DPGenIcon,
//   //   },
//   // ],
//   // 'Density Functional Theory': [
//   //   {
//   //     name: 'ABACUS',
//   //     description: 'Learn how to use ABACUS',
//   //     href: `/abacus`,
//   //     icon: ABACUSIcon,
//   //   },
//   // ],
//   // 'Finite Element Method': [
//   //   {
//   //     name: 'Protium',
//   //     description: 'Get all of our website information',
//   //     href: ``,
//   //     icon: Logo,
//   //   },
//   //   {
//   //     name: 'DeePMD-kit',
//   //     description: 'Learn how to use DeePMD-kit',
//   //     href: `/DeePMD-kit`,
//   //     icon: DeePMDIcon,
//   //   },
//   // ],
//   // 'Community Platform': [
//   //   {
//   //     name: 'Protium',
//   //     description: 'Get all of our website information',
//   //     href: ``,
//   //     icon: Logo,
//   //   },
//   //   {
//   //     name: 'AIS-Square',
//   //     description: 'Learn how to use DeePMD-kit',
//   //     href: `/DeePMD-kit`,
//   //     icon: AISSquareIcon,
//   //   },
//   // ],
// };

// const callsToAction = [
//   { name: 'See All Projects', href: '/space', icon: RocketLaunchIcon },
//   {
//     name: 'Follow in Github',
//     href: 'https://github.com/Protium',
//     icon: GitHubIcon,
//   },
//   { name: 'Contact us', href: '#', icon: EnvelopeIcon },
// ];

// export default function Projects(
//   props: Omit<
//     NavbarFullWidthPreviewProps,
//     'callsToAction' | 'buttonName' | 'options'
//   >
// ) {
//   return (
//     <NavbarFullWidthPreview
//       buttonName="project"
//       options={options}
//       callsToAction={callsToAction}
//       {...props}
//     />
//   );
// }


import {
  // BookmarkSquareIcon,
  EnvelopeIcon,
  // PuzzlePieceIcon,
  // RectangleGroupIcon,
} from '@heroicons/react/24/outline';

import { SiUnrealengine,SiUnity,SiProton,SiX,SiStmicroelectronics } from 'react-icons/si';
import { GitHubIcon } from '@/assets/SocialIcons';
import NavbarFullWidth from './NavbarFullWidth';
import type { NavbarFullWidthProps } from './types';
// import { color } from 'framer-motion';

const resources = [
  {
    name: 'Studio',
    description: '所有ScienceOL服务和社区的门户',
    href: ``,
    icon:SiStmicroelectronics,
    color:'text-sky-500',
  },
  {
    name: 'Xyzen',
    description: '面向实验室场景的专用智能体',
    href: `/chat`,
    icon: SiX,
    color:'text-amber-500',
  },
  {
    name: 'PROTIUM',
    description: '为科学计算设计的AI原生工作流引擎',
    href: `/deepmd-kit`,
    icon: SiProton,
    color:'text-indigo-500',
  },
  {
    name: 'Anti',
    description: '用于实验室模拟的3D数字孪生平台',
    href: `/deepmd-kit`,
    icon: SiUnity,
    color:'text-rose-500',
  },
  {
    name: 'Lab-OS',
    description: '用于模块化实验室硬件的开源操作系统',
    href: `/deepmd-kit`,
    icon: SiUnrealengine,
    color:'text-emerald-500',
  }
];

const callsToAction = [
  {
    name: 'Follow in Github',
    href: 'https://github.com/Protium',
    icon: GitHubIcon,
  },
  { name: 'Contact us', href: '#', icon: EnvelopeIcon },
];

export default function Projects(
  props: Omit<
    NavbarFullWidthProps,
    'solutions' | 'callsToAction' | 'buttonName'
  >
) {
  return (
    <NavbarFullWidth
      buttonName="project"
      solutions={resources}
      callsToAction={callsToAction}
      numberOfCallsToAction={2}
      numberOfSolutions={3}
      {...props}
    />
  );
}
