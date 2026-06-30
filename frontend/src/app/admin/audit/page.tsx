'use client';

import { useEffect, useState } from 'react';
import { adminFetch } from '@/lib/adminApi';

type Log = {
  id: number;
  admin_username: string;
  admin_discord_id: string;
  action: string;
  target_type: string;
  target_id: string;
  detail: string;
  created_at: string;
};

export default function AdminAuditPage() {
  const [logs, setLogs] = useState<Log[]>([]);
  const [category, setCategory] = useState('');

  const load = () => {
    const params = category ? `?category=${category}` : '';
    adminFetch<Log[]>(`/api/v1/admin/audit-logs${params}`).then(setLogs);
  };

  useEffect(load, [category]);

  return (
    <div>
      <h1 className="section-title">Admin Audit Logs</h1>
      <select value={category} onChange={(e) => setCategory(e.target.value)} style={{ maxWidth: 220, marginBottom: '1rem' }}>
        <option value="">All categories</option>
        <option value="security">Security</option>
        <option value="cms">CMS</option>
        <option value="monitoring">Monitoring</option>
        <option value="system">System</option>
        <option value="user">User</option>
      </select>
      <table>
        <thead>
          <tr><th>Time</th><th>Admin</th><th>Action</th><th>Target</th><th>Detail</th></tr>
        </thead>
        <tbody>
          {logs.map((l) => (
            <tr key={l.id}>
              <td>{new Date(l.created_at).toLocaleString()}</td>
              <td>{l.admin_username || l.admin_discord_id}</td>
              <td>{l.action}</td>
              <td>{l.target_type}:{l.target_id}</td>
              <td style={{ maxWidth: 240, overflow: 'hidden', textOverflow: 'ellipsis' }}>{l.detail}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
