'use client';

import { useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import ProductCard from '@/components/ProductCard';
import { API_URL } from '@/lib/api';

type Category = { id: number; slug: string; name: string };
type Product = {
  id: number;
  name: string;
  image_url: string;
  regular_price: number;
  sale_price: number;
  discount_pct?: number;
  category_id: number;
};

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function ShopPageInner() {
  const searchParams = useSearchParams();
  const [shopTab, setShopTab] = useState<'products' | 'packages'>('products');
  const [categories, setCategories] = useState<Category[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [packages, setPackages] = useState<Product[]>([]);
  const [categoryId, setCategoryId] = useState(0);
  const [search, setSearch] = useState('');
  const [loginNotice, setLoginNotice] = useState('');

  useEffect(() => {
    const login = searchParams.get('login');
    const message = searchParams.get('message');
    if (login === 'success') {
      setLoginNotice('Login successful.');
    } else if (login === 'error' && message) {
      setLoginNotice(decodeURIComponent(message));
    }
  }, [searchParams]);

  useEffect(() => {
    fetch(`${API_URL}/api/v1/catalog/categories`).then((r) => r.json()).then(setCategories);
    fetch(`${API_URL}/api/v1/catalog/packages`).then((r) => r.json()).then(setPackages);
  }, []);

  useEffect(() => {
    const params = new URLSearchParams();
    if (categoryId) params.set('category_id', String(categoryId));
    if (search) params.set('search', search);
    fetch(`${API_URL}/api/v1/catalog/products?${params}`)
      .then((r) => r.json())
      .then(setProducts);
  }, [categoryId, search]);

  const addToCart = async (product: Product) => {
    const token = getToken();
    if (!token) {
      window.location.href = `${API_URL}/api/v1/auth/discord`;
      return;
    }
    const price = product.sale_price > 0 ? product.sale_price : product.regular_price;
    await fetch(`${API_URL}/api/v1/cart/items`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      credentials: 'include',
      body: JSON.stringify({ type: 'product', id: product.id, name: product.name, quantity: 1, price }),
    });
    alert('Added to cart');
  };

  const addPackageToCart = async (product: Product) => {
    const token = getToken();
    if (!token) {
      window.location.href = `${API_URL}/api/v1/auth/discord`;
      return;
    }
    const price = product.sale_price > 0 ? product.sale_price : product.regular_price;
    await fetch(`${API_URL}/api/v1/cart/items`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      credentials: 'include',
      body: JSON.stringify({ type: 'package', id: product.id, name: product.name, quantity: 1, price }),
    });
    alert('Pack added to cart');
  };

  const list = shopTab === 'products' ? products : packages;
  const onAdd = shopTab === 'products' ? addToCart : addPackageToCart;

  return (
    <div className="container">
      <h1 className="section-title">Shop</h1>
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
        <button type="button" className={shopTab === 'products' ? 'btn btn-primary btn-sm' : 'btn btn-ghost btn-sm'} onClick={() => setShopTab('products')}>Products</button>
        <button type="button" className={shopTab === 'packages' ? 'btn btn-primary btn-sm' : 'btn btn-ghost btn-sm'} onClick={() => setShopTab('packages')}>Packs & Bundles</button>
      </div>
      {loginNotice && (
        <div
          className="card"
          style={{
            padding: '1rem',
            marginBottom: '1rem',
            borderColor: loginNotice.includes('successful') ? 'var(--success)' : 'var(--danger)',
            color: loginNotice.includes('successful') ? 'var(--success)' : 'var(--danger)',
          }}
        >
          {loginNotice}
        </div>
      )}
      {shopTab === 'products' && (
      <div className="filters">
        <select value={categoryId} onChange={(e) => setCategoryId(Number(e.target.value))}>
          <option value={0}>All Categories</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </select>
        <input placeholder="Search products..." value={search} onChange={(e) => setSearch(e.target.value)} />
      </div>
      )}
      <div className="grid-products">
        {list.map((p) => (
          <ProductCard key={p.id} product={p} onAdd={() => onAdd(p)} />
        ))}
      </div>
    </div>
  );
}
