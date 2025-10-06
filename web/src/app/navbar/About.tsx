import { GitHubIcon } from '@/assets/SocialIcons';
import { type NavbarFullWidthProps } from '@/types/navbar';
import {
  EnvelopeIcon,
  InformationCircleIcon,
  TrophyIcon,
  UserGroupIcon,
} from '@heroicons/react/24/outline';
import NavbarFullWidth from './NavbarFullWidth';

const solutions = [
  {
    name: 'About',
    description: 'Get a better understanding of us',
    href: '/about',
    icon: InformationCircleIcon,
  },
  {
    name: 'Manifesto',
    description: 'Our vision and mission',
    href: '/manifesto',
    icon: TrophyIcon,
  },
  {
    name: 'Join us',
    description: "We're always looking for new talents",
    href: '/about#joinus',
    icon: UserGroupIcon,
  },
];
const callsToAction = [
  {
    name: 'Follow in Github',
    href: 'https://github.com/Protium',
    icon: GitHubIcon,
  },
  { name: 'Contact us', href: '#', icon: EnvelopeIcon },
];

export default function About(
  props: Omit<
    NavbarFullWidthProps,
    'solutions' | 'callsToAction' | 'buttonName'
  >
) {
  return (
    <NavbarFullWidth
      buttonName="about.title"
      solutions={solutions}
      callsToAction={callsToAction}
      numberOfCallsToAction={2}
      numberOfSolutions={3}
      {...props}
    />
  );
}
