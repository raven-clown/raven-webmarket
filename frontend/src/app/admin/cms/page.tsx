'use client';

import { useState } from 'react';
import { API_URL } from '@/lib/api';

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function AdminCMSPage() {
  const [product, setProduct] = useState({
    category_id: 1, sku: '', name: '', description: '', image_url: '',
    regular_price: 0, sale_price: 0, esx_item_name: '', esx_item_count: 1,
    stock_limit: 0, max_limit_per_id: 0, expiry_date: null, is_featured: 0, is_active: 1,
  });
  const [banner, setBanner] = useState({
    title: '', image_url: '', link_url: '', sort_order: 0, is_active: 1,
  });

  const saveProduct = async () => {
    const token = getToken();
    await fetch(`${API_URL}/api/v1/admin/products`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify(product),
    });
    alert('Product saved');
  };

  const saveBanner = async () => {
    const token = getToken();
    await fetch(`${API_URL}/api/v1/admin/banners`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify(banner),
    });
    alert('Banner saved');
  };

  const productField = (label: string, key: keyof typeof product) => (
    <div style={{ marginBottom: '0.75rem' }}>
      <label style={{ display: 'block', marginBottom: '0.25rem', color: 'var(--muted)', fontSize: '0.85rem' }}>{label}</label>
      <input
        value={String(product[key] ?? '')}
        onChange={(e) => setProduct({ ...product, [key]: e.target.value })}
      />
    </div>
  );

  const bannerField = (label: string, key: keyof typeof banner) => (
    <div style={{ marginBottom: '0.75rem' }}>
      <label style={{ display: 'block', marginBottom: '0.25rem', color: 'var(--muted)', fontSize: '0.85rem' }}>{label}</label>
      <input
        value={String(banner[key] ?? '')}
        onChange={(e) => setBanner({ ...banner, [key]: e.target.value })}
      />
    </div>
  );

  return (
    <div>
      <h1 className="section-title">CMS Management</h1>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
        <div className="card" style={{ padding: '1.5rem' }}>
          <h2 style={{ marginBottom: '1rem' }}>Add Product</h2>
          {productField('SKU', 'sku')}
          {productField('Name', 'name')}
          {productField('ESX Item', 'esx_item_name')}
          {productField('Image URL', 'image_url')}
          {productField('Regular Price', 'regular_price')}
          {productField('Sale Price', 'sale_price')}
          <button className="btn btn-primary" onClick={saveProduct}>Save Product</button>
        </div>
        <div className="card" style={{ padding: '1.5rem' }}>
          <h2 style={{ marginBottom: '1rem' }}>Add Banner</h2>
          {bannerField('Title', 'title')}
          {bannerField('Image URL', 'image_url')}
          {bannerField('Link URL', 'link_url')}
          <button className="btn btn-primary" onClick={saveBanner}>Save Banner</button>
        </div>
      </div>
    </div>
  );
}
