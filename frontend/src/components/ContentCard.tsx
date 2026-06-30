'use client';

import Link from 'next/link';
import { useI18n } from '@/lib/i18n/I18nProvider';
import { pickLocalized } from '@/lib/i18n/translations';

type Post = {
  id: number;
  title_en: string;
  title_th?: string;
  body_en?: string;
  body_th?: string;
  publish_date?: string;
  is_pinned?: boolean;
};

export default function ContentCard({ post, href }: { post: Post; href?: string }) {
  const { locale } = useI18n();
  const { title, body } = pickLocalized(locale, post);
  const link = href ?? `/news#post-${post.id}`;

  return (
    <article className="content-card card">
      {post.is_pinned && <span className="pin-badge">📌</span>}
      <div className="content-card-body">
        <time className="content-date">
          {post.publish_date ? new Date(post.publish_date).toLocaleDateString(locale === 'th' ? 'th-TH' : 'en-US') : ''}
        </time>
        <h3>{title}</h3>
        <p>{body?.slice(0, 180)}{body && body.length > 180 ? '…' : ''}</p>
        <Link href={link} className="read-more">→</Link>
      </div>
    </article>
  );
}
