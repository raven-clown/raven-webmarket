'use client';

import { useEffect, useState } from 'react';
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

export default function ShopPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [categoryId, setCategoryId] = useState(0);
  const [search, setSearch] = useState('');

  useEffect(() => {
    fetch(`${API_URL}/api/v1/catalog/categories`).then((r) => r.json()).then(setCategories);
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

  return (
    <div className="container">
      <h1 className="section-title">All Products</h1>
      <div className="filters">
        <select value={categoryId} onChange={(e) => setCategoryId(Number(e.target.value))}>
          <option value={0}>All Categories</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </select>
        <input placeholder="Search products..." value={search} onChange={(e) => setSearch(e.target.value)} />
      </div>
      <div className="grid-products">
        {products.map((p) => (
          <ProductCard key={p.id} product={p} onAdd={() => addToCart(p)} />
        ))}
      </div>
    </div>
  );
}
