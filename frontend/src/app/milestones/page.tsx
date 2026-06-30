'use client';

import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';

type Tier = {
  id: number;
  tier_level: number;
  threshold_amount: number;
  reward_name: string;
  claimed: boolean;
  eligible: boolean;
};

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function MilestonesPage() {
  const [accumulation, setAccumulation] = useState(0);
  const [tiers, setTiers] = useState<Tier[]>([]);

  const load = () => {
    const token = getToken();
    if (!token) return;
    fetch(`${API_URL}/api/v1/milestones`, {
      headers: { Authorization: `Bearer ${token}` },
      credentials: 'include',
    })
      .then((r) => r.json())
      .then((data) => {
        setAccumulation(data.accumulation || 0);
        setTiers(data.tiers || []);
      });
  };

  useEffect(load, []);

  const claim = async (tierId: number) => {
    const token = getToken();
    const res = await fetch(`${API_URL}/api/v1/milestones/claim`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      credentials: 'include',
      body: JSON.stringify({ tier_id: tierId }),
    });
    if (res.ok) {
      alert('Reward claimed!');
      load();
    } else {
      const err = await res.json();
      alert(err.error || 'Claim failed');
    }
  };

  const maxThreshold = tiers.length ? Math.max(...tiers.map((t) => t.threshold_amount)) : 1;
  const progress = Math.min(100, (accumulation / maxThreshold) * 100);

  return (
    <div className="container">
      <h1 className="section-title">Monthly Milestone Event</h1>
      <p style={{ color: 'var(--muted)', marginBottom: '1rem' }}>
        Accumulated top-up this month: <strong>฿{accumulation.toLocaleString()}</strong>
      </p>
      <div className="progress-bar">
        <div className="progress-fill" style={{ width: `${progress}%` }} />
      </div>
      <div className="grid-products">
        {tiers.map((t) => (
          <div key={t.id} className="card" style={{ padding: '1.25rem' }}>
            <span className="badge" style={{ background: 'var(--accent)' }}>Tier {t.tier_level}</span>
            <h3 style={{ margin: '0.75rem 0' }}>{t.reward_name}</h3>
            <p style={{ color: 'var(--muted)', fontSize: '0.9rem' }}>
              Required: ฿{t.threshold_amount.toLocaleString()}
            </p>
            {t.claimed ? (
              <span style={{ color: 'var(--success)' }}>Claimed</span>
            ) : t.eligible ? (
              <button className="btn btn-primary" onClick={() => claim(t.id)}>Claim Reward</button>
            ) : (
              <span style={{ color: 'var(--muted)' }}>Not eligible yet</span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
