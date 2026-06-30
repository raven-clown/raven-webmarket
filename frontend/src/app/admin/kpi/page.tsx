'use client';

import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function AdminKPIPage() {
  const [revenue, setRevenue] = useState<{ period: string; amount: number }[]>([]);
  const [frequency, setFrequency] = useState<{ method: string; count: number }[]>([]);
  const [spenders, setSpenders] = useState<{ discord_id: string; display_name: string; total_amount: number; topup_count: number }[]>([]);
  const [period, setPeriod] = useState('daily');

  useEffect(() => {
    const token = getToken();
    const headers = { Authorization: `Bearer ${token}` };
    fetch(`${API_URL}/api/v1/admin/kpi/revenue?period=${period}`, { headers })
      .then((r) => r.json()).then(setRevenue);
    fetch(`${API_URL}/api/v1/admin/kpi/frequency`, { headers })
      .then((r) => r.json()).then(setFrequency);
    fetch(`${API_URL}/api/v1/admin/kpi/top-spenders`, { headers })
      .then((r) => r.json()).then(setSpenders);
  }, [period]);

  return (
    <div>
      <h1 className="section-title">KPI Analytics</h1>
      <select value={period} onChange={(e) => setPeriod(e.target.value)} style={{ maxWidth: 200, marginBottom: '1.5rem' }}>
        <option value="daily">Daily</option>
        <option value="weekly">Weekly</option>
        <option value="monthly">Monthly</option>
        <option value="yearly">Yearly</option>
      </select>
      <h2 style={{ marginBottom: '0.75rem' }}>Revenue Overview</h2>
      <table style={{ marginBottom: '2rem' }}>
        <thead><tr><th>Period</th><th>Amount</th></tr></thead>
        <tbody>
          {revenue.map((r) => (
            <tr key={r.period}><td>{r.period}</td><td>฿{r.amount.toLocaleString()}</td></tr>
          ))}
        </tbody>
      </table>
      <h2 style={{ marginBottom: '0.75rem' }}>Transaction Frequency by Method</h2>
      <table style={{ marginBottom: '2rem' }}>
        <thead><tr><th>Method</th><th>Count</th></tr></thead>
        <tbody>
          {frequency.map((f) => (
            <tr key={f.method}><td>{f.method}</td><td>{f.count}</td></tr>
          ))}
        </tbody>
      </table>
      <h2 style={{ marginBottom: '0.75rem' }}>Top Spenders</h2>
      <table>
        <thead><tr><th>Player</th><th>Discord</th><th>Total</th><th>Count</th></tr></thead>
        <tbody>
          {spenders.map((s) => (
            <tr key={s.discord_id}>
              <td>{s.display_name || '—'}</td>
              <td>{s.discord_id}</td>
              <td>฿{s.total_amount.toLocaleString()}</td>
              <td>{s.topup_count}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
