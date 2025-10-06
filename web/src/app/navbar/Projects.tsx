import { GitHubIcon } from '@/assets/SocialIcons';
import { type NavbarFullWidthPreviewProps } from '@/types/navbar';
import { EnvelopeIcon, RocketLaunchIcon } from '@heroicons/react/24/outline';

import {
  ABACUSIcon,
  AISSquareIcon,
  DeePMDIcon,
  DPGenIcon,
} from '@/assets/Icons';
import Logo from '@/assets/Logo';
import NavbarFullWidthPreview from './NavbarFullWidthPreview';

const options = {
  'Molecule Dynamics': [
    {
      name: 'DeePMD-kit',
      description: 'Learn how to use DeePMD-kit',
      href: `/space/DeePMD-kit`,
      icon: DeePMDIcon,
    },
    {
      name: 'DPGen',
      description: 'Learn from DPGen publications',
      href: `/dp-gen`,
      icon: DPGenIcon,
    },
    {
      name: 'DeePMD-kit',
      description: 'Learn how to use DeePMD-kit',
      href: `/DeePMD-kit`,
      icon: DeePMDIcon,
    },
    {
      name: 'DPGen',
      description: 'Learn from DPGen publications',
      href: `/DeePMD-kit`,
      icon: DPGenIcon,
    },
    {
      name: 'DeePMD-kit',
      description: 'Learn how to use DeePMD-kit',
      href: `/DeePMD-kit`,
      icon: DeePMDIcon,
    },
    {
      name: 'DPGen',
      description: 'Learn from DPGen publications',
      href: `/DeePMD-kit`,
      icon: DPGenIcon,
    },
    {
      name: 'DeePMD-kit',
      description: 'Learn how to use DeePMD-kit',
      href: `/DeePMD-kit`,
      icon: DeePMDIcon,
    },
    {
      name: 'DPGen',
      description: 'Learn from DPGen publications',
      href: `/DeePMD-kit`,
      icon: DPGenIcon,
    },
  ],
  'Density Functional Theory': [
    {
      name: 'ABACUS',
      description: 'Learn how to use ABACUS',
      href: `/abacus`,
      icon: ABACUSIcon,
    },
  ],
  'Finite Element Method': [
    {
      name: 'Protium',
      description: 'Get all of our website information',
      href: ``,
      icon: Logo,
    },
    {
      name: 'DeePMD-kit',
      description: 'Learn how to use DeePMD-kit',
      href: `/DeePMD-kit`,
      icon: DeePMDIcon,
    },
  ],
  'Community Platform': [
    {
      name: 'Protium',
      description: 'Get all of our website information',
      href: ``,
      icon: Logo,
    },
    {
      name: 'AIS-Square',
      description: 'Learn how to use DeePMD-kit',
      href: `/DeePMD-kit`,
      icon: AISSquareIcon,
    },
  ],
};

const callsToAction = [
  { name: 'See All Projects', href: '/space', icon: RocketLaunchIcon },
  {
    name: 'Follow in Github',
    href: 'https://github.com/Protium',
    icon: GitHubIcon,
  },
  { name: 'Contact us', href: '#', icon: EnvelopeIcon },
];

export default function Projects(
  props: Omit<
    NavbarFullWidthPreviewProps,
    'callsToAction' | 'buttonName' | 'options'
  >
) {
  return (
    <NavbarFullWidthPreview
      buttonName="project"
      options={options}
      callsToAction={callsToAction}
      {...props}
    />
  );
}
