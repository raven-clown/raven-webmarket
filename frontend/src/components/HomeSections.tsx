'use client';

import Link from 'next/link';
import ProductCard from '@/components/ProductCard';
import ContentCard from '@/components/ContentCard';
import { useI18n } from '@/lib/i18n/I18nProvider';

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
  title_en: string;
  title_th?: string;
  body_en?: string;
  body_th?: string;
  is_pinned?: boolean;
  publish_date?: string;
};

export default function HomeSections({
  featured,
  announcements,
  updates,
}: {
  featured: Product[];
  announcements: SitePost[];
  updates: SitePost[];
}) {
  const { strings } = useI18n();

  return (
    <>
      <section className="home-section">
        <div className="section-header">
          <h2 className="section-title">{strings.home.recommended}</h2>
          <Link href="/shop" className="section-link">{strings.home.browseAll}</Link>
        </div>
        <div className="grid-products">
          {featured.map((p) => (
            <ProductCard key={p.id} product={p} />
          ))}
        </div>
        {featured.length === 0 && <p className="empty-state">{strings.home.noFeatured}</p>}
      </section>

      <div className="home-grid-2">
        <section className="home-section">
          <div className="section-header">
            <h2 className="section-title">{strings.home.announcements}</h2>
            <Link href="/announcements" className="section-link">{strings.home.viewAll}</Link>
          </div>
          <div className="content-list">
            {announcements.slice(0, 3).map((p) => (
              <ContentCard key={p.id} post={p} href={`/announcements#post-${p.id}`} />
            ))}
            {announcements.length === 0 && <p className="empty-state">{strings.announcements.empty}</p>}
          </div>
        </section>

        <section className="home-section">
          <div className="section-header">
            <h2 className="section-title">{strings.home.dailyUpdates}</h2>
            <Link href="/news" className="section-link">{strings.home.viewAll}</Link>
          </div>
          <div className="content-list">
            {updates.map((p) => (
              <ContentCard key={p.id} post={p} href={`/news#post-${p.id}`} />
            ))}
            {updates.length === 0 && <p className="empty-state">{strings.news.empty}</p>}
          </div>
        </section>
      </div>

      <section className="home-cta">
        <Link href="/shop" className="btn btn-primary btn-lg">{strings.home.browseAll}</Link>
      </section>
    </>
  );
}
