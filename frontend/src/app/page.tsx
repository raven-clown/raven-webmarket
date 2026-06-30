import { fetchAPI } from '@/lib/api';
import HeroBanner from '@/components/HeroBanner';
import AnnouncementTicker from '@/components/AnnouncementTicker';
import AdSidebar from '@/components/AdSidebar';
import HomeSections from '@/components/HomeSections';

type Banner = { id: number; title: string; image_url: string; link_url: string };
type Product = {
  id: number;
  name: string;
  image_url: string;
  regular_price: number;
  sale_price: number;
  discount_pct?: number;
};
type SitePost = {
  id: number;
  post_type: string;
  title_en: string;
  title_th?: string;
  body_en?: string;
  body_th?: string;
  image_url?: string;
  link_url?: string;
  is_pinned?: boolean;
  publish_date?: string;
};

export default async function HomePage() {
  let banners: Banner[] = [];
  let featured: Product[] = [];
  let announcements: SitePost[] = [];
  let updates: SitePost[] = [];
  let ads: SitePost[] = [];

  try {
    [banners, featured, announcements, updates, ads] = await Promise.all([
      fetchAPI<Banner[]>('/api/v1/catalog/banners'),
      fetchAPI<Product[]>('/api/v1/catalog/products?featured=1'),
      fetchAPI<SitePost[]>('/api/v1/content/posts?type=announcement&placement=home&limit=5'),
      fetchAPI<SitePost[]>('/api/v1/content/posts?type=daily_update&limit=3'),
      fetchAPI<SitePost[]>('/api/v1/content/posts?type=ad&placement=sidebar&limit=3'),
    ]);
  } catch {
    banners = [];
    featured = [];
    announcements = [];
    updates = [];
    ads = [];
  }

  return (
    <>
      <AnnouncementTicker items={announcements} />
      <div className="container home-layout">
        <div className="home-main">
          <HeroBanner banners={banners} />
          <HomeSections
            featured={featured}
            announcements={announcements}
            updates={updates}
          />
        </div>
        <AdSidebar ads={ads} />
      </div>
    </>
  );
}
