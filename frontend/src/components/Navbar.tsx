'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';

type User = {
  discord_id: string;
  display_name: string;
  is_admin: boolean;
};

function getToken() {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function Navbar() {
  const [user, setUser] = useState<User | null>(null);

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

  return (
    <nav className="navbar">
      <div className="container navbar-inner">
        <Link href="/" className="logo">Raven Market</Link>
        <div className="nav-links">
          <Link href="/shop">Shop</Link>
          <Link href="/milestones">Milestones</Link>
          <Link href="/redeem">Redeem</Link>
          <Link href="/cart">Cart</Link>
          {user?.is_admin && <Link href="/admin">Admin</Link>}
          {user ? (
            <span style={{ color: 'var(--muted)' }}>{user.display_name || user.discord_id}</span>
          ) : (
            <a href={`${API_URL}/api/v1/auth/discord`} className="btn btn-primary">Login with Discord</a>
          )}
        </div>
      </div>
    </nav>
  );
}
