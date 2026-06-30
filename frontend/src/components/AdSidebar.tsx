'use client';

import Link from 'next/link';
import { useI18n } from '@/lib/i18n/I18nProvider';
import { pickLocalized } from '@/lib/i18n/translations';

type Ad = {
  id: number;
  title_en: string;
  title_th?: string;
  body_en?: string;
  body_th?: string;
  image_url?: string;
  link_url?: string;
};

export default function AdSidebar({ ads }: { ads: Ad[] }) {
  const { locale } = useI18n();
  if (!ads.length) return null;

  return (
    <aside className="ad-sidebar">
      {ads.map((ad) => {
        const { title, body } = pickLocalized(locale, ad);
        const inner = (
          <div className="ad-card card">
            {ad.image_url && (
              <div className="ad-image" style={{ backgroundImage: `url(${ad.image_url})` }} />
            )}
            <div className="ad-body">
              <span className="ad-badge">AD</span>
              <h4>{title}</h4>
              {body && <p>{body}</p>}
            </div>
          </div>
        );
        return ad.link_url ? (
          <a key={ad.id} href={ad.link_url} target="_blank" rel="noopener noreferrer sponsored" className="ad-link">
            {inner}
          </a>
        ) : (
          <div key={ad.id}>{inner}</div>
        );
      })}
    </aside>
  );
}
