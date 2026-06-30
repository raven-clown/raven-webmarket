'use client';

import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function AdminDashboard() {
  const [stats, setStats] = useState({ revenue: 0, peak: 0, spenders: 0 });

  useEffect(() => {
    const token = getToken();
    const headers = { Authorization: `Bearer ${token}` };
    Promise.all([
      fetch(`${API_URL}/api/v1/admin/kpi/revenue?period=monthly`, { headers }).then((r) => r.json()),
      fetch(`${API_URL}/api/v1/admin/kpi/peak`, { headers }).then((r) => r.json()),
      fetch(`${API_URL}/api/v1/admin/kpi/top-spenders`, { headers }).then((r) => r.json()),
    ]).then(([rev, peak, spenders]) => {
      const monthRev = Array.isArray(rev) && rev[0] ? rev[0].amount : 0;
      setStats({
        revenue: monthRev,
        peak: peak.amount || 0,
        spenders: Array.isArray(spenders) ? spenders.length : 0,
      });
    });
  }, []);

  const resetAll = async () => {
    if (!confirm('Reset all monthly accumulation and redeem points? Revenue logs will NOT be deleted.')) return;
    const token = getToken();
    await fetch(`${API_URL}/api/v1/admin/reset-accumulations`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    });
    alert('Accumulations reset.');
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
      <button className="btn btn-ghost" onClick={resetAll}>Reset All Accumulations</button>
    </div>
  );
}
