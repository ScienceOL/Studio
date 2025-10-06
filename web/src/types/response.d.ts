interface PaginatedResponseProps {
  count: number;
  next: string | null;
  previous: string | null;
  results: PostProps[];
}
