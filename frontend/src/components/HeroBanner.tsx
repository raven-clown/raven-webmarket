'use client';

import { useEffect, useState } from 'react';

type Banner = { id: number; title: string; image_url: string; link_url: string };

export default function HeroBanner({ banners }: { banners: Banner[] }) {
  const [index, setIndex] = useState(0);

  useEffect(() => {
    if (banners.length <= 1) return;
    const timer = setInterval(() => setIndex((i) => (i + 1) % banners.length), 5000);
    return () => clearInterval(timer);
  }, [banners.length]);

  if (!banners.length) {
    return (
      <div className="hero-slider" style={{ background: 'var(--surface2)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <span style={{ color: 'var(--muted)' }}>Welcome to Raven Webmarket</span>
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
            <h2>{b.title}</h2>
          </div>
        </div>
      ))}
    </div>
  );
}
