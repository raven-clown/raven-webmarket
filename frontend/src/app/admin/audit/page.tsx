'use client';

import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';

type Log = {
  id: number;
  admin_discord_id: string;
  action: string;
  target_type: string;
  target_id: string;
  created_at: string;
};

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function AdminAuditPage() {
  const [logs, setLogs] = useState<Log[]>([]);

  useEffect(() => {
    const token = getToken();
    fetch(`${API_URL}/api/v1/admin/audit-logs`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.json())
      .then(setLogs);
  }, []);

  return (
    <div>
      <h1 className="section-title">Admin Audit Logs</h1>
      <table>
        <thead>
          <tr><th>Time</th><th>Admin</th><th>Action</th><th>Target</th></tr>
        </thead>
        <tbody>
          {logs.map((l) => (
            <tr key={l.id}>
              <td>{new Date(l.created_at).toLocaleString()}</td>
              <td>{l.admin_discord_id}</td>
              <td>{l.action}</td>
              <td>{l.target_type}:{l.target_id}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
