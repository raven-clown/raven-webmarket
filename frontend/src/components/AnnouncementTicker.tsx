'use client';

import Link from 'next/link';
import { useI18n } from '@/lib/i18n/I18nProvider';
import { pickLocalized } from '@/lib/i18n/translations';

type Post = {
  id: number;
  title_en: string;
  title_th?: string;
  link_url?: string;
};

export default function AnnouncementTicker({ items }: { items: Post[] }) {
  const { locale } = useI18n();
  if (!items.length) return null;

  return (
    <div className="announcement-ticker">
      <div className="container ticker-inner">
        <span className="ticker-label">📢</span>
        <div className="ticker-track">
          {items.map((item) => {
            const { title } = pickLocalized(locale, item);
            const content = item.link_url ? (
              <a key={item.id} href={item.link_url} target="_blank" rel="noopener noreferrer">{title}</a>
            ) : (
              <Link key={item.id} href="/announcements">{title}</Link>
            );
            return <span key={item.id} className="ticker-item">{content}</span>;
          })}
        </div>
      </div>
    </div>
  );
}
