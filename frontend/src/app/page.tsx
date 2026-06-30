import Link from 'next/link';
import { fetchAPI } from '@/lib/api';
import HeroBanner from '@/components/HeroBanner';
import ProductCard from '@/components/ProductCard';

type Banner = { id: number; title: string; image_url: string; link_url: string };
type Product = {
  id: number;
  name: string;
  image_url: string;
  regular_price: number;
  sale_price: number;
  discount_pct?: number;
};

export default async function HomePage() {
  let banners: Banner[] = [];
  let featured: Product[] = [];
  try {
    [banners, featured] = await Promise.all([
      fetchAPI<Banner[]>('/api/v1/catalog/banners'),
      fetchAPI<Product[]>('/api/v1/catalog/products?featured=1'),
    ]);
  } catch {
    banners = [];
    featured = [];
  }

  return (
    <div className="container">
      <HeroBanner banners={banners} />
      <section>
        <h2 className="section-title">Recommended Promos</h2>
        <div className="grid-products">
          {featured.map((p) => (
            <ProductCard key={p.id} product={p} />
          ))}
        </div>
        {featured.length === 0 && <p style={{ color: 'var(--muted)' }}>No featured products yet.</p>}
      </section>
      <section style={{ textAlign: 'center', padding: '3rem 0' }}>
        <Link href="/shop" className="btn btn-primary">Browse All Products</Link>
      </section>
    </div>
  );
}
