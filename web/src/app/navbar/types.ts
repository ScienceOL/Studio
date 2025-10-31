export interface solutionsProps {
  name: string;
  description: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

export interface callsToActionProps {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

export interface resourcesProps {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

export interface recentPostsProps {
  id: number;
  title: string;
  href: string;
  date: string;
  datetime: string;
  category: { title: string; href: string };
  imageUrl: string;
  description: string;
}

export interface OptionsProps {
  [key: string]: solutionsProps[];
}

export interface NavbarFullWidthProps {
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

export interface NavbarFullWidthColumnsProps {
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

export interface NavbarFullWidthPreviewProps {
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
