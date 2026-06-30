'use client';

import Link from 'next/link';
import { useI18n } from '@/lib/i18n/I18nProvider';

export default function Footer() {
  const { strings } = useI18n();

  return (
    <footer className="site-footer">
      <div className="container footer-grid">
        <div>
          <div className="logo footer-logo">Raven Market</div>
          <p className="footer-tagline">{strings.footer.tagline}</p>
        </div>
        <div>
          <h4>{strings.footer.links}</h4>
          <Link href="/shop">{strings.nav.shop}</Link>
          <Link href="/forum">{strings.nav.forum}</Link>
          <Link href="/news">{strings.nav.news}</Link>
          <Link href="/announcements">{strings.nav.announcements}</Link>
        </div>
        <div>
          <h4>{strings.footer.legal}</h4>
          <Link href="/announcements">{strings.footer.privacy}</Link>
          <Link href="/admin/login">{strings.nav.admin}</Link>
        </div>
      </div>
      <div className="container footer-bottom">
        <span>© {new Date().getFullYear()} Raven Webmarket. {strings.footer.rights}</span>
      </div>
    </footer>
  );
}
