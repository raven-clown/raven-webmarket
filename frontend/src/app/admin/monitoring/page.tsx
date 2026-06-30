'use client';

import { useEffect, useState } from 'react';
import { adminFetch, isDevAdmin } from '@/lib/adminApi';

type HealthReport = {
  status: string;
  checked_at: string;
  services: Record<string, { status: string; message?: string; latency?: string }>;
  config: {
    enabled: boolean;
    check_interval_sec: number;
    cpu_alert_threshold: number;
    memory_alert_threshold: number;
    alert_webhook_url: string;
    prometheus_url: string;
  };
};

export default function MonitoringPage() {
  const [report, setReport] = useState<HealthReport | null>(null);
  const [config, setConfig] = useState<HealthReport['config'] | null>(null);
  const [role, setRole] = useState('');
  const [saved, setSaved] = useState('');

  const load = () => {
    adminFetch<HealthReport>('/api/v1/admin/health').then((r) => {
      setReport(r);
      setConfig(r.config);
    });
    adminFetch<{ role: string }>('/api/v1/admin/auth/me').then((m: { role: string }) => setRole(m.role));
  };

  useEffect(load, []);

  const saveConfig = async () => {
    if (!config) return;
    await adminFetch('/api/v1/admin/monitoring/config', { method: 'PUT', body: JSON.stringify(config) });
    setSaved('Monitoring configuration saved.');
    load();
  };

  return (
    <div>
      <h1 className="section-title">Health & Monitoring</h1>
      {report && (
        <>
          <div className="card" style={{ padding: '1.25rem', marginBottom: '1.5rem' }}>
            <p>Overall status: <strong style={{ color: report.status === 'healthy' ? 'var(--success)' : 'var(--danger)' }}>{report.status}</strong></p>
            <p style={{ color: 'var(--muted)', fontSize: '0.85rem' }}>Checked: {new Date(report.checked_at).toLocaleString()}</p>
          </div>
          <table style={{ marginBottom: '2rem' }}>
            <thead><tr><th>Service</th><th>Status</th><th>Details</th></tr></thead>
            <tbody>
              {Object.entries(report.services).map(([name, svc]) => (
                <tr key={name}>
                  <td>{name}</td>
                  <td style={{ color: svc.status === 'up' ? 'var(--success)' : 'var(--danger)' }}>{svc.status}</td>
                  <td>{svc.latency || svc.message || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
      <h2 style={{ marginBottom: '1rem' }}>Monitoring Configuration</h2>
      {config && (
        <div className="card" style={{ padding: '1.5rem', maxWidth: 520 }}>
          <div style={{ marginBottom: '0.75rem' }}>
            <label><input type="checkbox" checked={config.enabled} disabled={!isDevAdmin(role)} onChange={(e) => setConfig({ ...config, enabled: e.target.checked })} /> Enabled</label>
          </div>
          <div style={{ marginBottom: '0.75rem' }}>
            <label style={{ display: 'block', color: 'var(--muted)', marginBottom: '0.25rem' }}>Check interval (sec)</label>
            <input type="number" value={config.check_interval_sec} disabled={!isDevAdmin(role)} onChange={(e) => setConfig({ ...config, check_interval_sec: Number(e.target.value) })} />
          </div>
          <div style={{ marginBottom: '0.75rem' }}>
            <label style={{ display: 'block', color: 'var(--muted)', marginBottom: '0.25rem' }}>CPU alert threshold (%)</label>
            <input type="number" value={config.cpu_alert_threshold} disabled={!isDevAdmin(role)} onChange={(e) => setConfig({ ...config, cpu_alert_threshold: Number(e.target.value) })} />
          </div>
          <div style={{ marginBottom: '0.75rem' }}>
            <label style={{ display: 'block', color: 'var(--muted)', marginBottom: '0.25rem' }}>Memory alert threshold (%)</label>
            <input type="number" value={config.memory_alert_threshold} disabled={!isDevAdmin(role)} onChange={(e) => setConfig({ ...config, memory_alert_threshold: Number(e.target.value) })} />
          </div>
          <div style={{ marginBottom: '0.75rem' }}>
            <label style={{ display: 'block', color: 'var(--muted)', marginBottom: '0.25rem' }}>Alert webhook URL</label>
            <input value={config.alert_webhook_url} disabled={!isDevAdmin(role)} onChange={(e) => setConfig({ ...config, alert_webhook_url: e.target.value })} />
          </div>
          <div style={{ marginBottom: '0.75rem' }}>
            <label style={{ display: 'block', color: 'var(--muted)', marginBottom: '0.25rem' }}>Prometheus metrics path</label>
            <input value={config.prometheus_url} disabled={!isDevAdmin(role)} onChange={(e) => setConfig({ ...config, prometheus_url: e.target.value })} />
          </div>
          {isDevAdmin(role) && <button className="btn btn-primary" onClick={saveConfig}>Save Configuration</button>}
          {!isDevAdmin(role) && <p style={{ color: 'var(--muted)', fontSize: '0.85rem' }}>View only — dev admin required to edit.</p>}
          {saved && <p style={{ color: 'var(--success)', marginTop: '0.75rem' }}>{saved}</p>}
        </div>
      )}
    </div>
  );
}
