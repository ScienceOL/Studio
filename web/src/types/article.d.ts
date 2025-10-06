export interface ArticleProps {
  id: number;
  uuid: string;
  title: string;
  content: string;
  author: string;
  author_id: number;
  avatar: string;
  email: string;
  created_at: string;
  updated_at: string;
  publish: boolean;
}

export interface PageProps {
  count: number;
  next: string | null;
  previous: string | null;
  results: ArticleProps[];
}
