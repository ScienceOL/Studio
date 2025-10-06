import { GitHubIcon } from '@/assets/SocialIcons';
import { type NavbarFullWidthColumnsProps } from '@/types/navbar';
import {
  AdjustmentsHorizontalIcon,
  DocumentTextIcon,
  EnvelopeIcon,
  RocketLaunchIcon,
} from '@heroicons/react/24/outline';
import NavbarFullWidthColumns from './NavbarFullWidthColumns';

import {
  CubeTransparentIcon,
  NewspaperIcon,
} from '@heroicons/react/24/outline';
const resources = [
  {
    name: 'Spaces',
    description: 'Discussion to find the right',
    href: '/space',
    icon: CubeTransparentIcon,
  },
  {
    name: 'Ariticles',
    description: 'Dive into informative insights.',
    href: '/articles',
    icon: NewspaperIcon,
  },

  {
    name: 'Workflows',
    description: 'Automate your work',
    href: ``,
    icon: AdjustmentsHorizontalIcon,
  },
  {
    name: 'Tutorials',
    description: 'Learn here',
    href: '',
    icon: DocumentTextIcon,
  },
];

const recentPosts = [
  {
    id: 1,
    title: 'Anticipated Plan for Community Development',
    href: '#',
    date: 'Mar 16, 2023',
    datetime: '2023-03-16',
    category: { title: 'Development', href: '#' },
    imageUrl: '/hero4-horizen.png',
    description:
      'This is a plan for the community development of the DeePMD project. The plan includes the development of the community, the development of the project, and the development of the project. The plan is expected to be completed by the end of the year.',
  },
  {
    id: 2,
    title: 'DeePMD-PyTorch',
    href: '#',
    date: 'Mar 10, 2023',
    datetime: '2023-03-10',
    category: { title: 'Deploy', href: '#' },
    imageUrl: '/hero11.png',
    description:
      'DeePMD is a deep learning-based interatomic potential energy and force field. The original DeePMD-kit implementation is based on TensorFlow. This repository brings prior features for training, testing, and performing molecular dynamics (MD) with DeePMD to PyTorch framework. DeePMD-PyTorch also supports the multi-task pre-training of DPA-2, a large atomic model (LAM) which can be efficiently fine-tuned and distilled to downstream tasks.',
  },
];

const callsToAction = [
  { name: 'See All Projects', href: '/space', icon: RocketLaunchIcon },
  {
    name: 'Follow in Github',
    href: 'https://github.com/Protium',
    icon: GitHubIcon,
  },
  { name: 'Contact us', href: '#', icon: EnvelopeIcon },
];

export default function Community(
  props: Omit<
    NavbarFullWidthColumnsProps,
    'resources' | 'recentPosts' | 'buttonName' | 'engagements' | 'callsToAction'
  >
) {
  return (
    <NavbarFullWidthColumns
      buttonName="community"
      resources={resources}
      // engagements={engagement}
      callsToAction={callsToAction}
      recentPosts={recentPosts}
      {...props}
    />
  );
}
