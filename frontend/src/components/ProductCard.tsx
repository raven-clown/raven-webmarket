'use client';

import { useI18n } from '@/lib/i18n/I18nProvider';

type Product = {
  id: number;
  name: string;
  image_url: string;
  regular_price: number;
  sale_price: number;
  discount_pct?: number;
  stock_remaining?: number;
};

export default function ProductCard({ product, onAdd }: { product: Product; onAdd?: () => void }) {
  const { strings } = useI18n();
  const price = product.sale_price > 0 ? product.sale_price : product.regular_price;

  return (
    <div className="card product-card">
      <div
        className="product-image"
        style={{
          height: 160,
          background: product.image_url ? `url(${product.image_url}) center/cover` : undefined,
          backgroundColor: 'var(--surface2)',
        }}
      />
      <div style={{ padding: '1rem' }}>
        <h3 style={{ fontSize: '1rem', fontWeight: 600 }}>{product.name}</h3>
        <div className="product-price">
          <span className="price-sale">฿{price.toLocaleString()}</span>
          {product.sale_price > 0 && product.sale_price < product.regular_price && (
            <>
              <span className="price-regular">฿{product.regular_price.toLocaleString()}</span>
              {product.discount_pct ? <span className="badge">-{Math.round(product.discount_pct)}%</span> : null}
            </>
          )}
        </div>
        {onAdd && (
          <button className="btn btn-primary" style={{ width: '100%', marginTop: '0.75rem' }} onClick={onAdd}>
            {strings.common.addToCart}
          </button>
        )}
      </div>
    </div>
  );
}
