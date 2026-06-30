'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { API_URL } from '@/lib/api';
import { useI18n } from '@/lib/i18n/I18nProvider';

type Reply = { id: number; author_name: string; body: string; created_at: string };
type Thread = {
  id: number;
  author_name: string;
  title: string;
  body: string;
  is_locked: boolean;
  created_at: string;
};

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function ForumThreadPage() {
  const params = useParams();
  const id = params.id as string;
  const { strings } = useI18n();
  const [thread, setThread] = useState<Thread | null>(null);
  const [replies, setReplies] = useState<Reply[]>([]);
  const [replyBody, setReplyBody] = useState('');

  const load = () => {
    fetch(`${API_URL}/api/v1/forum/threads/${id}`)
      .then((r) => r.json())
      .then((data) => {
        setThread(data.thread);
        setReplies(data.replies || []);
      })
      .catch(() => {
        setThread(null);
        setReplies([]);
      });
  };

  useEffect(() => { load(); }, [id]);

  const postReply = async () => {
    const token = getToken();
    if (!token) {
      window.location.href = `${API_URL}/api/v1/auth/discord`;
      return;
    }
    await fetch(`${API_URL}/api/v1/forum/threads/${id}/replies`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      credentials: 'include',
      body: JSON.stringify({ body: replyBody }),
    });
    setReplyBody('');
    load();
  };

  if (!thread) return <div className="container"><p className="empty-state">{strings.common.loading}</p></div>;

  return (
    <div className="container page-forum">
      <Link href="/forum" className="back-link">← {strings.forum.back}</Link>
      <article className="card forum-thread-detail">
        <h1>{thread.title}</h1>
        <div className="forum-meta">
          <span>{thread.author_name}</span>
          <time>{new Date(thread.created_at).toLocaleString()}</time>
        </div>
        <div className="forum-body">{thread.body}</div>
      </article>

      <h2 className="section-subtitle">{replies.length} {strings.forum.replies}</h2>
      <div className="forum-replies">
        {replies.map((r) => (
          <div key={r.id} className="card forum-reply">
            <div className="forum-meta">
              <strong>{r.author_name}</strong>
              <time>{new Date(r.created_at).toLocaleString()}</time>
            </div>
            <p>{r.body}</p>
          </div>
        ))}
      </div>

      {!thread.is_locked && (
        <div className="card forum-form">
          <textarea className="forum-textarea" placeholder={strings.forum.bodyPlaceholder} value={replyBody} onChange={(e) => setReplyBody(e.target.value)} rows={4} />
          <button type="button" className="btn btn-primary" onClick={postReply}>{strings.forum.reply}</button>
        </div>
      )}
    </div>
  );
}
