'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

const links = [
  { href: '/admin', label: 'Dashboard' },
  { href: '/admin/users', label: 'User Search' },
  { href: '/admin/kpi', label: 'KPI Analytics' },
  { href: '/admin/audit', label: 'Audit Logs' },
  { href: '/admin/cms', label: 'CMS' },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  return (
    <div className="admin-layout">
      <aside className="admin-sidebar">
        <h3 style={{ marginBottom: '1rem' }}>Admin Panel</h3>
        {links.map((l) => (
          <Link key={l.href} href={l.href} className={pathname === l.href ? 'active' : ''}>
            {l.label}
          </Link>
        ))}
      </aside>
      <div style={{ padding: '1.5rem' }}>{children}</div>
    </div>
  );
}
