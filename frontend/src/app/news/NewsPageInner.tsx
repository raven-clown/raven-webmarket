'use client';

import { useI18n } from '@/lib/i18n/I18nProvider';
import { pickLocalized } from '@/lib/i18n/translations';

type SitePost = {
  id: number;
  title_en: string;
  title_th?: string;
  body_en?: string;
  body_th?: string;
  publish_date?: string;
  is_pinned?: boolean;
};

export default function NewsPageInner({ posts }: { posts: SitePost[] }) {
  const { locale, strings } = useI18n();

  return (
    <div className="container page-content">
      <header className="page-header-block">
        <h1 className="section-title">{strings.news.title}</h1>
        <p className="page-subtitle">{strings.news.subtitle}</p>
      </header>
      <div className="content-feed">
        {posts.map((post) => {
          const { title, body } = pickLocalized(locale, post);
          return (
            <article key={post.id} id={`post-${post.id}`} className="card content-feed-item">
              {post.is_pinned && <span className="pin-badge">📌</span>}
              <time className="content-date">
                {post.publish_date ? new Date(post.publish_date).toLocaleDateString(locale === 'th' ? 'th-TH' : 'en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' }) : ''}
              </time>
              <h2>{title}</h2>
              <p className="content-body">{body}</p>
            </article>
          );
        })}
        {posts.length === 0 && <p className="empty-state">{strings.news.empty}</p>}
      </div>
    </div>
  );
}
