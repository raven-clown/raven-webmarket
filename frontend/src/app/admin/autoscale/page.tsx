'use client';

import { useEffect, useState } from 'react';
import { adminFetch } from '@/lib/adminApi';

type AutoscaleConfig = {
  min_replicas: number;
  max_replicas: number;
  cpu_target_percent: number;
  memory_target_percent: number;
  enabled: boolean;
};

type ManifestResponse = {
  namespace: string;
  config: { api: AutoscaleConfig; frontend: AutoscaleConfig };
  yaml: { api_hpa: string; frontend_hpa: string };
  apply_commands: string[];
};

export default function AutoscalePage() {
  const [apiCfg, setApiCfg] = useState<AutoscaleConfig | null>(null);
  const [feCfg, setFeCfg] = useState<AutoscaleConfig | null>(null);
  const [manifest, setManifest] = useState<ManifestResponse | null>(null);
  const [msg, setMsg] = useState('');

  const loadManifest = () =>
    adminFetch<ManifestResponse>('/api/v1/admin/autoscale/manifest').then(setManifest);

  useEffect(() => {
    adminFetch<{ api: AutoscaleConfig; frontend: AutoscaleConfig }>('/api/v1/admin/autoscale/config').then((d) => {
      setApiCfg(d.api);
      setFeCfg(d.frontend);
    });
    loadManifest();
  }, []);

  const save = async (target: 'api' | 'frontend', config: AutoscaleConfig) => {
    await adminFetch('/api/v1/admin/autoscale/config', {
      method: 'PUT',
      body: JSON.stringify({ target, config }),
    });
    setMsg(`${target} HPA settings saved. Copy YAML below and apply to Kubernetes.`);
    loadManifest();
  };

  const copyText = async (text: string) => {
    await navigator.clipboard.writeText(text);
    setMsg('Copied to clipboard.');
  };

  const field = (label: string, cfg: AutoscaleConfig, setCfg: (c: AutoscaleConfig) => void, key: keyof AutoscaleConfig) => (
    <div style={{ marginBottom: '0.75rem' }}>
      <label style={{ display: 'block', color: 'var(--muted)', marginBottom: '0.25rem' }}>{label}</label>
      {typeof cfg[key] === 'boolean' ? (
        <input type="checkbox" checked={cfg[key] as boolean} onChange={(e) => setCfg({ ...cfg, [key]: e.target.checked })} />
      ) : (
        <input type="number" value={cfg[key] as number} onChange={(e) => setCfg({ ...cfg, [key]: Number(e.target.value) })} />
      )}
    </div>
  );

  return (
    <div>
      <h1 className="section-title">Kubernetes Pod Autoscale (HPA)</h1>
      <p style={{ color: 'var(--muted)', marginBottom: '0.5rem' }}>
        Dev Admin only. Targets scale at 60% CPU/RAM by default. Requires metrics-server on the cluster.
      </p>
      <div className="card" style={{ padding: '1rem', marginBottom: '1.5rem', fontSize: '0.9rem' }}>
        <strong>Install / apply to cluster</strong>
        <pre style={{ marginTop: '0.5rem', color: 'var(--muted)', whiteSpace: 'pre-wrap' }}>
{`# Linux / macOS / Git Bash
bash scripts/k8s-apply.sh

# Windows PowerShell
.\\scripts\\k8s-apply.ps1

# Or manually
kubectl apply -k deploy/kubernetes/
kubectl create secret generic raven-env --from-env-file=.env -n raven-webmarket`}
        </pre>
        <p style={{ color: 'var(--muted)', marginTop: '0.5rem' }}>
          Full guide: <code>DEPLOYMENT.md</code> → Kubernetes & HPA section
        </p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.5rem' }}>
        {apiCfg && (
          <div className="card" style={{ padding: '1.5rem' }}>
            <h2 style={{ marginBottom: '1rem' }}>API HPA</h2>
            {field('Min replicas', apiCfg, setApiCfg, 'min_replicas')}
            {field('Max replicas', apiCfg, setApiCfg, 'max_replicas')}
            {field('CPU target %', apiCfg, setApiCfg, 'cpu_target_percent')}
            {field('Memory target %', apiCfg, setApiCfg, 'memory_target_percent')}
            {field('Enabled', apiCfg, setApiCfg, 'enabled')}
            <button className="btn btn-primary" onClick={() => save('api', apiCfg)}>Save API HPA</button>
          </div>
        )}
        {feCfg && (
          <div className="card" style={{ padding: '1.5rem' }}>
            <h2 style={{ marginBottom: '1rem' }}>Frontend HPA</h2>
            {field('Min replicas', feCfg, setFeCfg, 'min_replicas')}
            {field('Max replicas', feCfg, setFeCfg, 'max_replicas')}
            {field('CPU target %', feCfg, setFeCfg, 'cpu_target_percent')}
            {field('Memory target %', feCfg, setFeCfg, 'memory_target_percent')}
            {field('Enabled', feCfg, setFeCfg, 'enabled')}
            <button className="btn btn-primary" onClick={() => save('frontend', feCfg)}>Save Frontend HPA</button>
          </div>
        )}
      </div>

      {msg && <p style={{ color: 'var(--success)', marginTop: '1rem' }}>{msg}</p>}

      {manifest?.yaml && (
        <>
          <h2 className="section-title" style={{ marginTop: '2rem' }}>Generated HPA YAML (apply after save)</h2>
          <div style={{ display: 'grid', gap: '1rem' }}>
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                <strong>deploy/kubernetes/hpa-api.yaml</strong>
                <button type="button" className="btn btn-ghost btn-sm" onClick={() => copyText(manifest.yaml.api_hpa)}>Copy</button>
              </div>
              <pre className="card" style={{ padding: '1rem', overflow: 'auto', fontSize: '0.78rem' }}>{manifest.yaml.api_hpa}</pre>
            </div>
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                <strong>deploy/kubernetes/hpa-frontend.yaml</strong>
                <button type="button" className="btn btn-ghost btn-sm" onClick={() => copyText(manifest.yaml.frontend_hpa)}>Copy</button>
              </div>
              <pre className="card" style={{ padding: '1rem', overflow: 'auto', fontSize: '0.78rem' }}>{manifest.yaml.frontend_hpa}</pre>
            </div>
          </div>
          {manifest.apply_commands?.length > 0 && (
            <>
              <h3 style={{ marginTop: '1.5rem', marginBottom: '0.5rem' }}>kubectl commands</h3>
              <pre className="card" style={{ padding: '1rem', overflow: 'auto', fontSize: '0.85rem' }}>
                {manifest.apply_commands.join('\n')}
              </pre>
            </>
          )}
        </>
      )}
    </div>
  );
}
