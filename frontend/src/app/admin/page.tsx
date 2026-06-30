'use client';

import { useEffect, useState } from 'react';
import { adminFetch, isDevAdmin } from '@/lib/adminApi';

export default function AdminDashboard() {
  const [stats, setStats] = useState({ revenue: 0, peak: 0, spenders: 0 });
  const [role, setRole] = useState('');

  useEffect(() => {
    Promise.all([
      adminFetch<{ role: string }>('/api/v1/admin/auth/me'),
      adminFetch<{ period: string; amount: number }[]>('/api/v1/admin/kpi/revenue?period=monthly'),
      adminFetch<{ amount: number }>('/api/v1/admin/kpi/peak'),
      adminFetch<unknown[]>('/api/v1/admin/kpi/top-spenders'),
    ]).then(([me, rev, peak, spenders]) => {
      setRole(me.role);
      const monthRev = Array.isArray(rev) && rev[0] ? rev[0].amount : 0;
      setStats({
        revenue: monthRev,
        peak: peak.amount || 0,
        spenders: Array.isArray(spenders) ? spenders.length : 0,
      });
    });
  }, []);

  const resetMonthly = async () => {
    if (!confirm('Reset monthly accumulation only? (Redeem points kept)')) return;
    await adminFetch('/api/v1/admin/reset-accumulations', { method: 'POST' });
    alert('Monthly accumulation reset.');
  };

  const resetAll = async () => {
    if (!confirm('Reset monthly accumulation AND redeem points? Revenue logs will NOT be deleted.')) return;
    await adminFetch('/api/v1/admin/reset-accumulations?include_redeem=1', { method: 'POST' });
    alert('All accumulations reset.');
  };

  return (
    <div>
      <h1 className="section-title">Admin Dashboard</h1>
      <div className="grid-products" style={{ marginBottom: '2rem' }}>
        <div className="card" style={{ padding: '1.5rem' }}>
          <p style={{ color: 'var(--muted)' }}>Monthly Revenue</p>
          <h2>฿{stats.revenue.toLocaleString()}</h2>
        </div>
        <div className="card" style={{ padding: '1.5rem' }}>
          <p style={{ color: 'var(--muted)' }}>Peak Top-up</p>
          <h2>฿{stats.peak.toLocaleString()}</h2>
        </div>
        <div className="card" style={{ padding: '1.5rem' }}>
          <p style={{ color: 'var(--muted)' }}>Top Spenders Tracked</p>
          <h2>{stats.spenders}</h2>
        </div>
      </div>
      <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
        <button className="btn btn-ghost" onClick={resetMonthly}>Reset Monthly Accumulation</button>
        {isDevAdmin(role) && (
          <button className="btn btn-ghost" onClick={resetAll}>Reset All (+ Redeem Points)</button>
        )}
      </div>
    </div>
  );
}
