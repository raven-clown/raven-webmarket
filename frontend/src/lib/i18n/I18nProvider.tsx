'use client';

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { defaultLocale, t, type Locale, type TranslationKeys } from './translations';

const STORAGE_KEY = 'raven_lang';
const COOKIE_KEY = 'raven_cookie_consent';

type I18nContextValue = {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  strings: TranslationKeys;
  cookieConsent: 'none' | 'essential' | 'all';
  setCookieConsent: (value: 'essential' | 'all') => void;
};

const I18nContext = createContext<I18nContextValue | null>(null);

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(defaultLocale);
  const [cookieConsent, setCookieConsentState] = useState<'none' | 'essential' | 'all'>('none');
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY) as Locale | null;
    if (stored === 'en' || stored === 'th') setLocaleState(stored);
    const consent = document.cookie.match(/raven_cookie_consent=([^;]+)/)?.[1];
    if (consent === 'essential' || consent === 'all') setCookieConsentState(consent);
    setReady(true);
  }, []);

  const setLocale = useCallback((next: Locale) => {
    setLocaleState(next);
    localStorage.setItem(STORAGE_KEY, next);
    document.documentElement.lang = next;
  }, []);

  const setCookieConsent = useCallback((value: 'essential' | 'all') => {
    setCookieConsentState(value);
    const maxAge = 365 * 24 * 60 * 60;
    document.cookie = `${COOKIE_KEY}=${value}; path=/; max-age=${maxAge}; SameSite=Lax`;
  }, []);

  const value = useMemo(
    () => ({
      locale,
      setLocale,
      strings: t(locale),
      cookieConsent,
      setCookieConsent,
    }),
    [locale, setLocale, cookieConsent, setCookieConsent]
  );

  if (!ready) return null;

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error('useI18n must be used within I18nProvider');
  return ctx;
}
