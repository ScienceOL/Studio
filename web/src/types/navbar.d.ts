interface solutionsProps {
  name: string;
  description: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface callsToActionProps {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface resourcesProps {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface recentPostsProps {
  id: number;
  title: string;
  href: string;
  date: string;
  datetime: string;
  category: { title: string; href: string };
  imageUrl: string;
  description: string;
}

interface OptionsProps {
  [key: string]: solutionsProps[];
}

interface NavbarFullWidthProps {
  buttonName: string;

  solutions: solutionsProps[];
  callsToAction: callsToActionProps[];
  index: number;
  activeMenuItem: number | null;
  setActiveMenuItem: (value: number | null) => void;
  open: boolean;
  setOpen: (value: boolean) => void;
  numberOfCallsToAction?: number;
  numberOfSolutions?: number;
}

interface NavbarFullWidthColumnsProps {
  buttonName: string;
  resources: resourcesProps[];
  engagements?: resourcesProps[];
  callsToAction: callsToActionProps[];

  recentPosts: recentPostsProps[];
  index: number;
  activeMenuItem: number | null;
  setActiveMenuItem: (value: number | null) => void;
  open: boolean;
  setOpen: (value: boolean) => void;
  numberOfCallsToAction?: number;
}

interface NavbarFullWidthPreviewProps {
  buttonName: string;
  options: OptionsProps;
  callsToAction: callsToActionProps[];
  index: number;
  activeMenuItem: number | null;
  setActiveMenuItem: (value: number | null) => void;
  open: boolean;
  setOpen: (value: boolean) => void;
  numberOfCallsToAction?: number;
  numberOfSolutions?: number;
}
