'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { API_URL } from '@/lib/api';
import { useI18n } from '@/lib/i18n/I18nProvider';

type Thread = {
  id: number;
  author_name: string;
  title: string;
  body: string;
  reply_count: number;
  is_pinned: boolean;
  is_locked: boolean;
  created_at: string;
};

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function ForumPage() {
  const { strings } = useI18n();
  const [threads, setThreads] = useState<Thread[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [loggedIn, setLoggedIn] = useState(false);

  const load = () => {
    fetch(`${API_URL}/api/v1/forum/threads`)
      .then((r) => r.json())
      .then(setThreads)
      .catch(() => setThreads([]));
  };

  useEffect(() => {
    load();
    setLoggedIn(!!getToken());
  }, []);

  const createThread = async () => {
    const token = getToken();
    if (!token) {
      window.location.href = `${API_URL}/api/v1/auth/discord`;
      return;
    }
    await fetch(`${API_URL}/api/v1/forum/threads`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      credentials: 'include',
      body: JSON.stringify({ title, body }),
    });
    setTitle('');
    setBody('');
    setShowForm(false);
    load();
  };

  return (
    <div className="container page-forum">
      <div className="page-header">
        <h1 className="section-title">{strings.forum.title}</h1>
        {loggedIn ? (
          <button type="button" className="btn btn-primary" onClick={() => setShowForm((v) => !v)}>
            {strings.forum.newThread}
          </button>
        ) : (
          <a href={`${API_URL}/api/v1/auth/discord`} className="btn btn-primary">{strings.nav.login}</a>
        )}
      </div>

      {showForm && (
        <div className="card forum-form">
          <input placeholder={strings.forum.titlePlaceholder} value={title} onChange={(e) => setTitle(e.target.value)} />
          <textarea className="forum-textarea" placeholder={strings.forum.bodyPlaceholder} value={body} onChange={(e) => setBody(e.target.value)} rows={5} />
          <button type="button" className="btn btn-primary" onClick={createThread}>{strings.forum.post}</button>
        </div>
      )}

      <div className="forum-list">
        {threads.map((t) => (
          <Link key={t.id} href={`/forum/${t.id}`} className="card forum-thread">
            {t.is_pinned && <span className="pin-badge">{strings.forum.pinned}</span>}
            <h3>{t.title}</h3>
            <p className="forum-preview">{t.body.slice(0, 120)}…</p>
            <div className="forum-meta">
              <span>{t.author_name}</span>
              <span>{t.reply_count} {strings.forum.replies}</span>
              <time>{new Date(t.created_at).toLocaleDateString()}</time>
            </div>
          </Link>
        ))}
        {threads.length === 0 && <p className="empty-state">{strings.forum.noThreads}</p>}
      </div>
    </div>
  );
}
