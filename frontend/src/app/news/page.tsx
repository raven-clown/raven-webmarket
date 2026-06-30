import { fetchAPI } from '@/lib/api';
import NewsPageInner from './NewsPageInner';

type SitePost = {
  id: number;
  title_en: string;
  title_th?: string;
  body_en?: string;
  body_th?: string;
  publish_date?: string;
  is_pinned?: boolean;
};

export default async function NewsPage() {
  let posts: SitePost[] = [];
  try {
    posts = await fetchAPI<SitePost[]>('/api/v1/content/posts?type=daily_update&limit=30');
  } catch {
    posts = [];
  }
  return <NewsPageInner posts={posts} />;
}
