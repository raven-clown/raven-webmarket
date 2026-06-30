import { fetchAPI } from '@/lib/api';
import AnnouncementsPageInner from './AnnouncementsPageInner';

type SitePost = {
  id: number;
  title_en: string;
  title_th?: string;
  body_en?: string;
  body_th?: string;
  publish_date?: string;
  is_pinned?: boolean;
  link_url?: string;
};

export default async function AnnouncementsPage() {
  let posts: SitePost[] = [];
  try {
    posts = await fetchAPI<SitePost[]>('/api/v1/content/posts?type=announcement&limit=30');
  } catch {
    posts = [];
  }
  return <AnnouncementsPageInner posts={posts} />;
}
