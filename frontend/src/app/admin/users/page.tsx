'use client';

import { useState } from 'react';
import { adminFetch } from '@/lib/adminApi';

type User = {
  discord_id: string;
  identifier: string;
  display_name: string;
  monthly_accumulation: number;
  redeem_points: number;
  total_topup_amount: number;
  topup_count: number;
};

type Topup = {
  id: number;
  tx_ref: string;
  amount: number;
  payment_method: string;
  slip_url: string;
  status: string;
  created_at: string;
};


export default function AdminUsersPage() {
  const [query, setQuery] = useState('');
  const [users, setUsers] = useState<User[]>([]);
  const [topups, setTopups] = useState<Topup[]>([]);
  const [selected, setSelected] = useState('');
  const [slipUrl, setSlipUrl] = useState('');

  const search = async () => {
    const res = await adminFetch<User[]>(`/api/v1/admin/users/search?q=${encodeURIComponent(query)}`);
    setUsers(res);
  };

  const loadTopups = async (discordId: string) => {
    setSelected(discordId);
    const res = await adminFetch<Topup[]>(`/api/v1/admin/users/${discordId}/topups`);
    setTopups(res);
  };

  return (
    <div>
      <h1 className="section-title">User Billing Search</h1>
      <div className="filters">
        <input placeholder="Discord ID, identifier, or name..." value={query} onChange={(e) => setQuery(e.target.value)} />
        <button className="btn btn-primary" onClick={search}>Search</button>
      </div>
      <table>
        <thead>
          <tr><th>Discord</th><th>Identifier</th><th>Monthly</th><th>Points</th><th>Total Top-up</th><th></th></tr>
        </thead>
        <tbody>
          {users.map((u) => (
            <tr key={u.discord_id}>
              <td>{u.discord_id}</td>
              <td>{u.identifier}</td>
              <td>฿{u.monthly_accumulation.toLocaleString()}</td>
              <td>{u.redeem_points.toLocaleString()}</td>
              <td>฿{u.total_topup_amount.toLocaleString()} ({u.topup_count}x)</td>
              <td><button className="btn btn-ghost" onClick={() => loadTopups(u.discord_id)}>History</button></td>
            </tr>
          ))}
        </tbody>
      </table>
      {selected && (
        <>
          <h2 className="section-title">Top-up History — {selected}</h2>
          <table>
            <thead>
              <tr><th>Ref</th><th>Amount</th><th>Method</th><th>Status</th><th>Slip</th></tr>
            </thead>
            <tbody>
              {topups.map((t) => (
                <tr key={t.id}>
                  <td>{t.tx_ref}</td>
                  <td>฿{t.amount.toLocaleString()}</td>
                  <td>{t.payment_method}</td>
                  <td>{t.status}</td>
                  <td>
                    {t.slip_url ? (
                      <button className="btn btn-ghost" onClick={() => setSlipUrl(t.slip_url)}>View Slip</button>
                    ) : '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
      {slipUrl && (
        <div className="modal-overlay" onClick={() => setSlipUrl('')}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <img src={slipUrl} alt="Payment slip" />
          </div>
        </div>
      )}
    </div>
  );
}
