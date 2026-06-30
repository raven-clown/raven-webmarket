'use client';

import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';

type RedeemItem = {
  id: number;
  name: string;
  description: string;
  image_url: string;
  point_cost: number;
};

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function RedeemPage() {
  const [items, setItems] = useState<RedeemItem[]>([]);
  const [points, setPoints] = useState(0);

  const load = () => {
    fetch(`${API_URL}/api/v1/redeem/catalog`).then((r) => r.json()).then(setItems);
    const token = getToken();
    if (token) {
      fetch(`${API_URL}/api/v1/auth/me`, {
        headers: { Authorization: `Bearer ${token}` },
        credentials: 'include',
      })
        .then((r) => r.json())
        .then((u) => setPoints(u.redeem_points || 0));
    }
  };

  useEffect(load, []);

  const redeem = async (catalogId: number) => {
    const token = getToken();
    if (!token) {
      window.location.href = `${API_URL}/api/v1/auth/discord`;
      return;
    }
    const res = await fetch(`${API_URL}/api/v1/redeem`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      credentials: 'include',
      body: JSON.stringify({ catalog_id: catalogId }),
    });
    if (res.ok) {
      alert('Redeemed successfully!');
      load();
    } else {
      const err = await res.json();
      alert(err.error || 'Redeem failed');
    }
  };

  return (
    <div className="container">
      <h1 className="section-title">Redeem Points</h1>
      <p style={{ color: 'var(--muted)', marginBottom: '1.5rem' }}>
        Your redeem points: <strong>{points.toLocaleString()}</strong>
      </p>
      <div className="grid-products">
        {items.map((item) => (
          <div key={item.id} className="card" style={{ padding: '1.25rem' }}>
            <h3>{item.name}</h3>
            <p style={{ color: 'var(--muted)', fontSize: '0.9rem', margin: '0.5rem 0' }}>{item.description}</p>
            <p style={{ fontWeight: 600, color: 'var(--accent2)' }}>{item.point_cost.toLocaleString()} pts</p>
            <button className="btn btn-primary" style={{ marginTop: '0.75rem' }} onClick={() => redeem(item.id)}>
              Redeem
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
