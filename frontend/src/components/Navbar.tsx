'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';
import { useI18n } from '@/lib/i18n/I18nProvider';
import LanguageSwitcher from '@/components/LanguageSwitcher';

type User = {
  discord_id: string;
  display_name: string;
};

function getToken() {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function Navbar() {
  const { strings } = useI18n();
  const [user, setUser] = useState<User | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    const token = getToken();
    if (!token) return;
    fetch(`${API_URL}/api/v1/auth/me`, {
      headers: { Authorization: `Bearer ${token}` },
      credentials: 'include',
    })
      .then((r) => (r.ok ? r.json() : null))
      .then(setUser)
      .catch(() => null);
  }, []);

  const links = [
    { href: '/', label: strings.nav.home },
    { href: '/shop', label: strings.nav.shop },
    { href: '/milestones', label: strings.nav.milestones },
    { href: '/redeem', label: strings.nav.redeem },
    { href: '/forum', label: strings.nav.forum },
    { href: '/news', label: strings.nav.news },
    { href: '/announcements', label: strings.nav.announcements },
    { href: '/cart', label: strings.nav.cart },
  ];

  return (
    <nav className="navbar">
      <div className="container navbar-inner">
        <Link href="/" className="logo">Raven Market</Link>
        <button
          type="button"
          className="nav-toggle"
          aria-label="Menu"
          onClick={() => setMenuOpen((v) => !v)}
        >
          ☰
        </button>
        <div className={`nav-links ${menuOpen ? 'open' : ''}`}>
          {links.map((l) => (
            <Link key={l.href} href={l.href} onClick={() => setMenuOpen(false)}>{l.label}</Link>
          ))}
          <LanguageSwitcher />
          <Link href="/admin/login" className="nav-admin">{strings.nav.admin}</Link>
          {user ? (
            <span className="nav-user">{user.display_name || user.discord_id}</span>
          ) : (
            <a href={`${API_URL}/api/v1/auth/discord`} className="btn btn-primary btn-sm">
              {strings.nav.login}
            </a>
          )}
        </div>
      </div>
    </nav>
  );
}
