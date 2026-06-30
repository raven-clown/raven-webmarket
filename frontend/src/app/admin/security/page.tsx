'use client';

import { useEffect, useState } from 'react';
import { adminFetch } from '@/lib/adminApi';

type Account = {
  id: number;
  username: string;
  role: string;
  display_name: string;
  is_active: boolean;
  permissions?: string[];
};

export default function SecurityPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [allPerms, setAllPerms] = useState<string[]>([]);
  const [defaultPerms, setDefaultPerms] = useState<string[]>([]);
  const [form, setForm] = useState({ username: '', password: '', role: 'admin', display_name: '', permissions: [] as string[] });
  const [editId, setEditId] = useState<number | null>(null);
  const [editPerms, setEditPerms] = useState<string[]>([]);
  const [msg, setMsg] = useState('');

  const load = () => {
    adminFetch<Account[]>('/api/v1/admin/accounts').then(setAccounts);
    adminFetch<{ all: string[]; default: string[] }>('/api/v1/admin/permissions').then((d) => {
      setAllPerms(d.all);
      setDefaultPerms(d.default);
    });
  };

  useEffect(() => { load(); }, []);

  const togglePerm = (list: string[], perm: string, set: (v: string[]) => void) => {
    set(list.includes(perm) ? list.filter((p) => p !== perm) : [...list, perm]);
  };

  const permGrid = (selected: string[], set: (v: string[]) => void) => (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: '0.35rem', marginBottom: '1rem' }}>
      {allPerms.map((p) => (
        <label key={p} style={{ fontSize: '0.8rem', cursor: 'pointer' }}>
          <input type="checkbox" checked={selected.includes(p)} onChange={() => togglePerm(selected, p, set)} /> {p}
        </label>
      ))}
    </div>
  );

  const create = async () => {
    const perms = form.role === 'dev_admin' ? [] : (form.permissions.length ? form.permissions : defaultPerms);
    await adminFetch('/api/v1/admin/accounts', { method: 'POST', body: JSON.stringify({ ...form, permissions: perms }) });
    setForm({ username: '', password: '', role: 'admin', display_name: '', permissions: [] });
    setMsg('Account created — custom permissions applied for admin role.');
    load();
  };

  const savePerms = async (id: number) => {
    await adminFetch(`/api/v1/admin/accounts/${id}`, { method: 'PUT', body: JSON.stringify({ permissions: editPerms }) });
    setEditId(null);
    setMsg('Permissions updated. User must re-login.');
    load();
  };

  return (
    <div>
      <h1 className="section-title">Security & Admin Accounts</h1>
      <p style={{ color: 'var(--muted)', marginBottom: '1rem' }}>
        Dev Admin ปรับสิทธิ์แต่ละ admin ได้ไม่เหมือนกัน — เลือก permission ที่ต้องการ (dev_admin ได้ทุกอย่างเสมอ)
      </p>
      <div className="card" style={{ padding: '1.5rem', marginBottom: '2rem', maxWidth: 720 }}>
        <h2>Create Admin Account</h2>
        <input placeholder="Username" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} style={{ marginBottom: '0.75rem', width: '100%' }} />
        <input type="password" placeholder="Password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} style={{ marginBottom: '0.75rem', width: '100%' }} />
        <input placeholder="Display name" value={form.display_name} onChange={(e) => setForm({ ...form, display_name: e.target.value })} style={{ marginBottom: '0.75rem', width: '100%' }} />
        <select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })} style={{ marginBottom: '0.75rem' }}>
          <option value="admin">Admin (custom permissions below)</option>
          <option value="dev_admin">Dev Admin (full access)</option>
        </select>
        {form.role === 'admin' && (
          <>
            <p style={{ fontSize: '0.85rem', color: 'var(--muted)' }}>Permissions (empty = default set)</p>
            {permGrid(form.permissions, (p) => setForm({ ...form, permissions: p }))}
          </>
        )}
        <button className="btn btn-primary" onClick={create}>Create Account</button>
        {msg && <p style={{ color: 'var(--success)', marginTop: '0.75rem' }}>{msg}</p>}
      </div>
      <table>
        <thead><tr><th>Username</th><th>Role</th><th>Permissions</th><th>Actions</th></tr></thead>
        <tbody>
          {accounts.map((a) => (
            <tr key={a.id}>
              <td>{a.username}</td>
              <td>{a.role}</td>
              <td style={{ fontSize: '0.75rem' }}>{a.role === 'dev_admin' ? 'all' : (a.permissions?.join(', ') || 'default')}</td>
              <td>
                {a.role === 'admin' && (
                  <button type="button" className="btn btn-ghost btn-sm" onClick={() => { setEditId(a.id); setEditPerms(a.permissions?.length ? a.permissions : defaultPerms); }}>
                    Edit perms
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {editId !== null && (
        <div className="card" style={{ padding: '1.5rem', marginTop: '1rem', maxWidth: 720 }}>
          <h3>Edit permissions #{editId}</h3>
          {permGrid(editPerms, setEditPerms)}
          <button className="btn btn-primary" onClick={() => savePerms(editId)}>Save</button>
          <button className="btn btn-ghost" style={{ marginLeft: 8 }} onClick={() => setEditId(null)}>Cancel</button>
        </div>
      )}
    </div>
  );
}
