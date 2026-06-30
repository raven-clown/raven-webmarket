'use client';

import { useI18n } from '@/lib/i18n/I18nProvider';

export default function LanguageSwitcher() {
  const { locale, setLocale } = useI18n();

  return (
    <div className="lang-switch" role="group" aria-label="Language">
      <button
        type="button"
        className={locale === 'en' ? 'active' : ''}
        onClick={() => setLocale('en')}
        aria-pressed={locale === 'en'}
      >
        EN
      </button>
      <button
        type="button"
        className={locale === 'th' ? 'active' : ''}
        onClick={() => setLocale('th')}
        aria-pressed={locale === 'th'}
      >
        TH
      </button>
    </div>
  );
}
