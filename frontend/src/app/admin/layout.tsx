'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { adminFetch, getAdminToken, isDevAdmin } from '@/lib/adminApi';

type AdminMe = { username: string; role: string; display_name: string };

const allLinks = [
  { href: '/admin', label: 'Dashboard', roles: ['admin', 'dev_admin'] },
  { href: '/admin/monitoring', label: 'Health & Monitoring', roles: ['admin', 'dev_admin'] },
  { href: '/admin/autoscale', label: 'Pod Autoscale', roles: ['dev_admin'] },
  { href: '/admin/security', label: 'Security & Accounts', roles: ['dev_admin'] },
  { href: '/admin/users', label: 'User Search', roles: ['admin', 'dev_admin'] },
  { href: '/admin/purchases', label: 'Purchase Logs', roles: ['admin', 'dev_admin'] },
  { href: '/admin/activity', label: 'Activity Logs', roles: ['admin', 'dev_admin'] },
  { href: '/admin/kpi', label: 'KPI Analytics', roles: ['admin', 'dev_admin'] },
  { href: '/admin/audit', label: 'Admin Audit', roles: ['admin', 'dev_admin'] },
  { href: '/admin/cms', label: 'Shop & CMS', roles: ['admin', 'dev_admin'] },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [me, setMe] = useState<AdminMe | null>(null);

  useEffect(() => {
    if (pathname === '/admin/login') return;
    if (!getAdminToken()) {
      router.replace('/admin/login');
      return;
    }
    adminFetch<AdminMe>('/api/v1/admin/auth/me').then(setMe).catch(() => router.replace('/admin/login'));
  }, [pathname, router]);

  if (pathname === '/admin/login') {
    return <>{children}</>;
  }

  const links = allLinks.filter((l) => me && l.roles.includes(me.role));

  return (
    <div className="admin-layout">
      <aside className="admin-sidebar">
        <h3 style={{ marginBottom: '0.25rem' }}>Admin Panel</h3>
        {me && (
          <p style={{ color: 'var(--muted)', fontSize: '0.8rem', marginBottom: '1rem' }}>
            {me.display_name || me.username} · <strong>{me.role}</strong>
          </p>
        )}
        {links.map((l) => (
          <Link key={l.href} href={l.href} className={pathname === l.href ? 'active' : ''}>
            {l.label}
          </Link>
        ))}
        <button
          className="btn btn-ghost"
          style={{ marginTop: '1.5rem', width: '100%' }}
          onClick={() => {
            document.cookie = 'raven_admin_token=; Path=/; Max-Age=0';
            router.push('/admin/login');
          }}
        >
          Logout
        </button>
      </aside>
      <div style={{ padding: '1.5rem' }}>{children}</div>
    </div>
  );
}
