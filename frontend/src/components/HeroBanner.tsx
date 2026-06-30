'use client';

import { useEffect, useState } from 'react';
import { useI18n } from '@/lib/i18n/I18nProvider';

type Banner = { id: number; title: string; image_url: string; link_url: string };

export default function HeroBanner({ banners }: { banners: Banner[] }) {
  const [index, setIndex] = useState(0);
  const { strings } = useI18n();

  useEffect(() => {
    if (banners.length <= 1) return;
    const timer = setInterval(() => setIndex((i) => (i + 1) % banners.length), 5000);
    return () => clearInterval(timer);
  }, [banners.length]);

  if (!banners.length) {
    return (
      <div className="hero-slider hero-empty">
        <div className="hero-overlay">
          <h2>Raven Webmarket</h2>
          <p style={{ color: 'var(--muted)', marginTop: '0.5rem' }}>{strings.footer.tagline}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="hero-slider">
      {banners.map((b, i) => (
        <div
          key={b.id}
          className={`hero-slide ${i === index ? 'active' : ''}`}
          style={{ backgroundImage: `url(${b.image_url})` }}
        >
          <div className="hero-overlay">
            {b.link_url ? (
              <a href={b.link_url} target="_blank" rel="noopener noreferrer">
                <h2>{b.title}</h2>
              </a>
            ) : (
              <h2>{b.title}</h2>
            )}
          </div>
        </div>
      ))}
      {banners.length > 1 && (
        <div className="hero-dots">
          {banners.map((b, i) => (
            <button
              key={b.id}
              type="button"
              className={`hero-dot ${i === index ? 'active' : ''}`}
              onClick={() => setIndex(i)}
              aria-label={`Slide ${i + 1}`}
            />
          ))}
        </div>
      )}
    </div>
  );
}
