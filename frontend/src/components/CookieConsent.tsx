'use client';

import Link from 'next/link';
import { useI18n } from '@/lib/i18n/I18nProvider';

export default function CookieConsent() {
  const { strings, cookieConsent, setCookieConsent } = useI18n();

  if (cookieConsent !== 'none') return null;

  return (
    <div className="cookie-banner" role="dialog" aria-label={strings.cookie.title}>
      <div className="cookie-banner-inner">
        <div>
          <strong>{strings.cookie.title}</strong>
          <p>{strings.cookie.body}</p>
        </div>
        <div className="cookie-actions">
          <Link href="/announcements" className="btn btn-ghost btn-sm">
            {strings.cookie.learnMore}
          </Link>
          <button type="button" className="btn btn-ghost btn-sm" onClick={() => setCookieConsent('essential')}>
            {strings.cookie.essential}
          </button>
          <button type="button" className="btn btn-primary btn-sm" onClick={() => setCookieConsent('all')}>
            {strings.cookie.accept}
          </button>
        </div>
      </div>
    </div>
  );
}
