const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export type AdminSession = {
  username: string;
  role: 'admin' | 'dev_admin';
  display_name?: string;
};

export function getAdminToken() {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(/raven_admin_token=([^;]+)/);
  return match ? match[1] : '';
}

export async function adminFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const token = getAdminToken();
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      ...options?.headers,
    },
  });
  if (res.status === 401) {
    if (typeof window !== 'undefined') window.location.href = '/admin/login';
    throw new Error('unauthorized');
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || 'request failed');
  }
  return res.json();
}

export async function adminLogin(username: string, password: string) {
  const res = await fetch(`${API_URL}/api/v1/admin/auth/login`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) throw new Error('invalid credentials');
  return res.json();
}

export function isDevAdmin(role?: string) {
  return role === 'dev_admin';
}
