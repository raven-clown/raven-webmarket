'use client';

import { useEffect, useState } from 'react';
import { adminFetch } from '@/lib/adminApi';

type PurchaseLog = {
  id: number;
  order_ref: string;
  discord_id: string;
  identifier: string;
  total_amount: number;
  status: string;
  delivery_status: string;
  item_count: number;
  created_at: string;
};

export default function PurchaseLogsPage() {
  const [logs, setLogs] = useState<PurchaseLog[]>([]);
  const [discordID, setDiscordID] = useState('');
  const [status, setStatus] = useState('');
  const [detail, setDetail] = useState<object | null>(null);

  const load = () => {
    const params = new URLSearchParams();
    if (discordID) params.set('discord_id', discordID);
    if (status) params.set('status', status);
    adminFetch<PurchaseLog[]>(`/api/v1/admin/purchase-logs?${params}`).then(setLogs);
  };

  useEffect(load, []);

  const openDetail = async (orderRef: string) => {
    const data = await adminFetch<object>(`/api/v1/admin/purchase-logs/${orderRef}`);
    setDetail(data);
  };

  return (
    <div>
      <h1 className="section-title">Purchase Logs</h1>
      <div className="filters">
        <input placeholder="Discord ID" value={discordID} onChange={(e) => setDiscordID(e.target.value)} />
        <select value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">All status</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
          <option value="processing">Processing</option>
        </select>
        <button className="btn btn-primary" onClick={load}>Search</button>
      </div>
      <table>
        <thead>
          <tr><th>Time</th><th>Ref</th><th>Discord</th><th>Steam</th><th>Total</th><th>Status</th><th>Items</th><th></th></tr>
        </thead>
        <tbody>
          {logs.map((l) => (
            <tr key={l.id}>
              <td>{new Date(l.created_at).toLocaleString()}</td>
              <td>{l.order_ref}</td>
              <td>{l.discord_id}</td>
              <td>{l.identifier}</td>
              <td>฿{l.total_amount.toLocaleString()}</td>
              <td>{l.status} / {l.delivery_status}</td>
              <td>{l.item_count}</td>
              <td><button className="btn btn-ghost" onClick={() => openDetail(l.order_ref)}>Detail</button></td>
            </tr>
          ))}
        </tbody>
      </table>
      {detail && (
        <div className="modal-overlay" onClick={() => setDetail(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()} style={{ minWidth: 480 }}>
            <h2 style={{ marginBottom: '1rem' }}>Order Detail</h2>
            <pre style={{ fontSize: '0.8rem', whiteSpace: 'pre-wrap' }}>{JSON.stringify(detail, null, 2)}</pre>
          </div>
        </div>
      )}
    </div>
  );
}
