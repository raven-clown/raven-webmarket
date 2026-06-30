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
  link_url?: string;
};

export default function AnnouncementsPageInner({ posts }: { posts: SitePost[] }) {
  const { locale, strings } = useI18n();

  return (
    <div className="container page-content">
      <header className="page-header-block">
        <h1 className="section-title">{strings.announcements.title}</h1>
        <p className="page-subtitle">{strings.announcements.subtitle}</p>
      </header>
      <div className="content-feed">
        {posts.map((post) => {
          const { title, body } = pickLocalized(locale, post);
          return (
            <article key={post.id} id={`post-${post.id}`} className="card content-feed-item announcement-item">
              {post.is_pinned && <span className="pin-badge">📌 {strings.forum.pinned}</span>}
              <time className="content-date">
                {post.publish_date ? new Date(post.publish_date).toLocaleDateString(locale === 'th' ? 'th-TH' : 'en-US') : ''}
              </time>
              <h2>{title}</h2>
              <p className="content-body">{body}</p>
              {post.link_url && (
                <a href={post.link_url} target="_blank" rel="noopener noreferrer" className="btn btn-ghost btn-sm">
                  {strings.common.readMore}
                </a>
              )}
            </article>
          );
        })}
        {posts.length === 0 && <p className="empty-state">{strings.announcements.empty}</p>}
      </div>

      <section className="card privacy-notice" style={{ marginTop: '2rem', padding: '1.5rem' }}>
        <h3>{strings.cookie.title}</h3>
        <p>{strings.cookie.body}</p>
      </section>
    </div>
  );
}
