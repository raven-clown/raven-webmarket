'use client';

import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';

type CartItem = { type: string; id: number; name: string; quantity: number; price: number };
type Cart = { items: CartItem[]; total: number };

function getToken() {
  const match = document.cookie.match(/raven_token=([^;]+)/);
  return match ? match[1] : '';
}

export default function CartPage() {
  const [cart, setCart] = useState<Cart>({ items: [], total: 0 });
  const [loading, setLoading] = useState(false);

  const load = () => {
    const token = getToken();
    if (!token) return;
    fetch(`${API_URL}/api/v1/cart`, {
      headers: { Authorization: `Bearer ${token}` },
      credentials: 'include',
    })
      .then((r) => r.json())
      .then(setCart);
  };

  useEffect(load, []);

  const updateQty = async (item: CartItem, quantity: number) => {
    const token = getToken();
    if (!token) return;
    await fetch(`${API_URL}/api/v1/cart/items`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      credentials: 'include',
      body: JSON.stringify({ type: item.type, id: item.id, quantity }),
    });
    load();
  };

  const removeItem = async (item: CartItem) => {
    const token = getToken();
    if (!token) return;
    await fetch(`${API_URL}/api/v1/cart/items`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      credentials: 'include',
      body: JSON.stringify({ type: item.type, id: item.id }),
    });
    load();
  };

  const checkout = async () => {
    const token = getToken();
    if (!token) {
      window.location.href = `${API_URL}/api/v1/auth/discord`;
      return;
    }
    setLoading(true);
    const res = await fetch(`${API_URL}/api/v1/orders/checkout`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      credentials: 'include',
    });
    setLoading(false);
    if (res.ok) {
      const data = await res.json();
      alert(`Order placed: ${data.order_ref}`);
      load();
    } else {
      const err = await res.json();
      alert(err.error || 'Checkout failed');
    }
  };

  return (
    <div className="container">
      <h1 className="section-title">Shopping Cart</h1>
      {cart.items.length === 0 ? (
        <p style={{ color: 'var(--muted)' }}>Your cart is empty.</p>
      ) : (
        <>
          <table>
            <thead>
              <tr><th>Item</th><th>Qty</th><th>Price</th><th></th></tr>
            </thead>
            <tbody>
              {cart.items.map((item) => (
                <tr key={`${item.type}-${item.id}`}>
                  <td>{item.name}</td>
                  <td>
                    <input
                      type="number"
                      min={1}
                      value={item.quantity}
                      style={{ width: 60 }}
                      onChange={(e) => updateQty(item, Number(e.target.value))}
                    />
                  </td>
                  <td>฿{(item.price * item.quantity).toLocaleString()}</td>
                  <td>
                    <button type="button" className="btn btn-ghost btn-sm" onClick={() => removeItem(item)}>Remove</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <div style={{ marginTop: '1.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <strong>Total: ฿{cart.total.toLocaleString()}</strong>
            <button className="btn btn-primary" onClick={checkout} disabled={loading}>
              {loading ? 'Processing...' : 'Checkout'}
            </button>
          </div>
        </>
      )}
    </div>
  );
}
