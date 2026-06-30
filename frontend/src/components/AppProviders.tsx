'use client';

import { I18nProvider } from '@/lib/i18n/I18nProvider';
import CookieConsent from '@/components/CookieConsent';

export default function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <I18nProvider>
      {children}
      <CookieConsent />
    </I18nProvider>
  );
}
