'use client';

import { useEffect, useState } from 'react';
import { adminFetch } from '@/lib/adminApi';

type ActivityLog = {
  id: number;
  category: string;
  actor_type: string;
  actor_id: string;
  action: string;
  target_type: string;
  target_id: string;
  detail: string;
  created_at: string;
};

export default function ActivityLogsPage() {
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [category, setCategory] = useState('');
  const [actorID, setActorID] = useState('');

  const load = () => {
    const params = new URLSearchParams();
    if (category) params.set('category', category);
    if (actorID) params.set('actor_id', actorID);
    adminFetch<ActivityLog[]>(`/api/v1/admin/activity-logs?${params}`).then(setLogs);
  };

  useEffect(load, []);

  return (
    <div>
      <h1 className="section-title">Activity Logs</h1>
      <div className="filters">
        <select value={category} onChange={(e) => setCategory(e.target.value)}>
          <option value="">All categories</option>
          <option value="purchase">Purchase</option>
          <option value="topup">Top-up</option>
          <option value="redeem">Redeem</option>
          <option value="milestone">Milestone</option>
          <option value="delivery">Delivery</option>
          <option value="security">Security</option>
          <option value="admin">Admin</option>
        </select>
        <input placeholder="Actor ID / Discord" value={actorID} onChange={(e) => setActorID(e.target.value)} />
        <button className="btn btn-primary" onClick={load}>Filter</button>
      </div>
      <table>
        <thead>
          <tr><th>Time</th><th>Category</th><th>Actor</th><th>Action</th><th>Target</th><th>Detail</th></tr>
        </thead>
        <tbody>
          {logs.map((l) => (
            <tr key={l.id}>
              <td>{new Date(l.created_at).toLocaleString()}</td>
              <td>{l.category}</td>
              <td>{l.actor_type}:{l.actor_id}</td>
              <td>{l.action}</td>
              <td>{l.target_type}:{l.target_id}</td>
              <td style={{ maxWidth: 280, overflow: 'hidden', textOverflow: 'ellipsis' }}>{l.detail}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
