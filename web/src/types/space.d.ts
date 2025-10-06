import { ArticleProps } from './article';

export interface SpaceProps {
  uuid: string;
  name: string;
  description: string;
  owner: string;
  owner_avatar: string;
  icon: string;
  admins: string[];
  members: string[];
  readme: ArticleProps;
  pinned_manuscript: {
    id: number;
    document: ArticleProps;
    server: string;
    order: number;
  }[];
  groups: string[];
  banner: string;
  channel_server: DiscussionProps[];
  github_url: string; // 空间对应的 GitHub 仓库地址，由用户填写
  document_url: string; // 空间对应的 Github 文档地址，由用户填写
  created_at: string;
  enable_releases: boolean;
  enable_discussion: boolean;
}

export interface DiscussionProps {
  uuid: string;
  name: string;
  description: string;
  owner: string;
  admins: string[];
  privacy: 'public' | 'apply' | 'memberOnly' | 'private';
  icon?: string;
  created_at?: string;
  members?: string[];
  progress: 'archived' | 'inProgress' | 'completed';
  server: uuid;
  latest_message: MessageProps;
  // commenters: string[];
  // status: string;
}

export interface MessageProps {
  id: number;
  sender: string;
  avatar: string;
  content: string;
  timestamp: string;
}
