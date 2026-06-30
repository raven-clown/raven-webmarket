'use client';

import { useEffect, useState } from 'react';
import { adminFetch } from '@/lib/adminApi';

type Tab = 'products' | 'packages' | 'promotions' | 'milestones' | 'redeem' | 'banners' | 'content';

const emptyProduct = {
  category_id: 1, sku: '', name: '', description: '', image_url: '',
  regular_price: 0, sale_price: 0, esx_item_name: '', esx_item_count: 1,
  stock_limit: 0, max_limit_per_id: 1, expiry_date: '', sale_start_date: '',
  is_featured: 0, is_active: 1, sort_order: 0,
};

const emptyPackage = {
  sku: '', name: '', description: '', image_url: '',
  regular_price: 0, sale_price: 0, stock_limit: 0, max_limit_per_id: 1,
  expiry_date: '', sale_start_date: '', is_featured: 0, is_active: 1,
};

const emptyPromo = {
  id: 0, name: '', description: '', target_type: 'product', target_id: 0,
  banner_image_url: '', regular_price: 0, sale_price: 0, max_limit_per_id: 1,
  start_date: '', end_date: '', is_active: true, sort_order: 0,
};

const emptyTier = { tier_level: 1, threshold_amount: 0, reward_name: '', esx_item_name: '', esx_item_count: 1 };

export default function AdminCMSPage() {
  const [tab, setTab] = useState<Tab>('products');
  const [categories, setCategories] = useState<{ id: number; name: string }[]>([]);
  const [products, setProducts] = useState<Record<string, unknown>[]>([]);
  const [packages, setPackages] = useState<Record<string, unknown>[]>([]);
  const [promotions, setPromotions] = useState<Record<string, unknown>[]>([]);
  const [product, setProduct] = useState(emptyProduct);
  const [pkg, setPkg] = useState(emptyPackage);
  const [pkgItems, setPkgItems] = useState([{ esx_item_name: '', esx_item_count: 1 }]);
  const [promo, setPromo] = useState(emptyPromo);
  const [milestone, setMilestone] = useState({ name: 'Monthly Top-up', month_year: new Date().toISOString().slice(0, 7), is_active: true, tiers: [emptyTier] });
  const [redeem, setRedeem] = useState({ name: '', description: '', image_url: '', point_cost: 0, esx_item_name: '', esx_item_count: 1, stock_limit: 0, is_active: 1 });
  const [banner, setBanner] = useState({ title: '', image_url: '', link_url: '', sort_order: 0, is_active: 1 });
  const [post, setPost] = useState({
    post_type: 'announcement', title_en: '', title_th: '', body_en: '', body_th: '',
    image_url: '', link_url: '', placement: 'home', sort_order: 0, is_pinned: 0, is_active: 1,
    publish_date: new Date().toISOString().slice(0, 10),
  });
  const [msg, setMsg] = useState('');

  const load = () => {
    adminFetch<{ id: number; name: string }[]>('/api/v1/admin/catalog/categories').then(setCategories).catch(() => {});
    adminFetch<Record<string, unknown>[]>('/api/v1/admin/catalog/products').then(setProducts).catch(() => {});
    adminFetch<Record<string, unknown>[]>('/api/v1/admin/catalog/packages').then(setPackages).catch(() => {});
    adminFetch<Record<string, unknown>[]>('/api/v1/admin/promotions').then(setPromotions).catch(() => {});
  };

  useEffect(() => { load(); }, []);

  const field = (label: string, value: string | number, set: (v: string) => void, type = 'text') => (
    <div style={{ marginBottom: '0.75rem' }}>
      <label style={{ display: 'block', marginBottom: '0.25rem', color: 'var(--muted)', fontSize: '0.85rem' }}>{label}</label>
      <input type={type} value={value} onChange={(e) => set(e.target.value)} />
    </div>
  );

  const saveProduct = async () => {
    await adminFetch('/api/v1/admin/products', { method: 'POST', body: JSON.stringify(product) });
    setMsg('Product saved'); load();
  };

  const savePackage = async () => {
    await adminFetch('/api/v1/admin/packages', { method: 'POST', body: JSON.stringify({ package: pkg, items: pkgItems }) });
    setMsg('Package saved'); load();
  };

  const savePromo = async () => {
    await adminFetch('/api/v1/admin/promotions', { method: 'POST', body: JSON.stringify(promo) });
    setMsg('Promotion saved'); load();
  };

  const saveMilestone = async () => {
    await adminFetch('/api/v1/admin/milestones/events', { method: 'POST', body: JSON.stringify(milestone) });
    setMsg('Milestone event saved');
  };

  const saveRedeem = async () => {
    await adminFetch('/api/v1/admin/redeem/catalog', { method: 'POST', body: JSON.stringify(redeem) });
    setMsg('Redeem item saved');
  };

  const saveBanner = async () => {
    await adminFetch('/api/v1/admin/banners', { method: 'POST', body: JSON.stringify(banner) });
    setMsg('Banner saved');
  };

  const savePost = async () => {
    await adminFetch('/api/v1/admin/content/posts', { method: 'POST', body: JSON.stringify(post) });
    setMsg('Content saved');
  };

  const tabs: { id: Tab; label: string }[] = [
    { id: 'products', label: 'Products' },
    { id: 'packages', label: 'Packs' },
    { id: 'promotions', label: 'Promotions' },
    { id: 'milestones', label: 'Milestones' },
    { id: 'redeem', label: 'Redeem' },
    { id: 'banners', label: 'Banners' },
    { id: 'content', label: 'News/Ads' },
  ];

  return (
    <div>
      <h1 className="section-title">Shop & CMS Management</h1>
      <p style={{ color: 'var(--muted)', marginBottom: '1rem' }}>
        ตั้งสินค้า แพ็ก โปรโมชัน ยอดสะสม แลกของ ราคา/วันหมดอายุ จำกัด 1 ไอดี 1 สิทธิ์ (max_limit_per_id=1)
      </p>
      {msg && <p style={{ color: 'var(--success)', marginBottom: '1rem' }}>{msg}</p>}
      <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', marginBottom: '1.5rem' }}>
        {tabs.map((t) => (
          <button key={t.id} type="button" className={tab === t.id ? 'btn btn-primary btn-sm' : 'btn btn-ghost btn-sm'} onClick={() => { setTab(t.id); setMsg(''); }}>
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'products' && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
          <div className="card" style={{ padding: '1.5rem' }}>
            <h2>Add / Edit Product</h2>
            <select value={product.category_id} onChange={(e) => setProduct({ ...product, category_id: Number(e.target.value) })} style={{ marginBottom: '0.75rem' }}>
              {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
            {field('SKU (unique)', product.sku, (v) => setProduct({ ...product, sku: v }))}
            {field('Name', product.name, (v) => setProduct({ ...product, name: v }))}
            {field('Description', product.description, (v) => setProduct({ ...product, description: v }))}
            {field('Image URL', product.image_url, (v) => setProduct({ ...product, image_url: v }))}
            {field('ESX Item Name', product.esx_item_name, (v) => setProduct({ ...product, esx_item_name: v }))}
            {field('ESX Item Count', product.esx_item_count, (v) => setProduct({ ...product, esx_item_count: Number(v) }), 'number')}
            {field('Regular Price', product.regular_price, (v) => setProduct({ ...product, regular_price: Number(v) }), 'number')}
            {field('Sale Price (0 = no sale)', product.sale_price, (v) => setProduct({ ...product, sale_price: Number(v) }), 'number')}
            {field('Stock Limit (0 = unlimited)', product.stock_limit, (v) => setProduct({ ...product, stock_limit: Number(v) }), 'number')}
            {field('Max per Discord ID (1 = one purchase)', product.max_limit_per_id, (v) => setProduct({ ...product, max_limit_per_id: Number(v) }), 'number')}
            {field('Sale Start (datetime)', product.sale_start_date, (v) => setProduct({ ...product, sale_start_date: v }))}
            {field('Expiry (datetime)', product.expiry_date, (v) => setProduct({ ...product, expiry_date: v }))}
            {field('Featured (0/1)', product.is_featured, (v) => setProduct({ ...product, is_featured: Number(v) }), 'number')}
            {field('Active (0/1)', product.is_active, (v) => setProduct({ ...product, is_active: Number(v) }), 'number')}
            <button className="btn btn-primary" onClick={saveProduct}>Save Product</button>
          </div>
          <div className="card" style={{ padding: '1rem', overflow: 'auto', maxHeight: 600 }}>
            <h3>Existing ({products.length})</h3>
            {products.map((p) => (
              <div key={String(p.id)} style={{ borderBottom: '1px solid var(--border)', padding: '0.5rem 0', fontSize: '0.85rem' }}>
                <strong>{String(p.name)}</strong> · {String(p.sku)} · ฿{String(p.sale_price || p.regular_price)}
                <button type="button" className="btn btn-ghost btn-sm" style={{ marginLeft: 8 }} onClick={() => setProduct({ ...emptyProduct, ...p } as typeof product)}>Edit</button>
              </div>
            ))}
          </div>
        </div>
      )}

      {tab === 'packages' && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
          <div className="card" style={{ padding: '1.5rem' }}>
            <h2>Pack / Bundle</h2>
            {field('SKU', pkg.sku, (v) => setPkg({ ...pkg, sku: v }))}
            {field('Name', pkg.name, (v) => setPkg({ ...pkg, name: v }))}
            {field('Description', pkg.description, (v) => setPkg({ ...pkg, description: v }))}
            {field('Image URL', pkg.image_url, (v) => setPkg({ ...pkg, image_url: v }))}
            {field('Regular Price', pkg.regular_price, (v) => setPkg({ ...pkg, regular_price: Number(v) }), 'number')}
            {field('Sale Price', pkg.sale_price, (v) => setPkg({ ...pkg, sale_price: Number(v) }), 'number')}
            {field('Max per ID', pkg.max_limit_per_id, (v) => setPkg({ ...pkg, max_limit_per_id: Number(v) }), 'number')}
            {field('Sale Start', pkg.sale_start_date, (v) => setPkg({ ...pkg, sale_start_date: v }))}
            {field('Expiry', pkg.expiry_date, (v) => setPkg({ ...pkg, expiry_date: v }))}
            <h3 style={{ marginTop: '1rem' }}>Items in pack</h3>
            {pkgItems.map((it, i) => (
              <div key={i} style={{ display: 'flex', gap: '0.5rem', marginBottom: '0.5rem' }}>
                <input placeholder="ESX item" value={it.esx_item_name} onChange={(e) => { const n = [...pkgItems]; n[i].esx_item_name = e.target.value; setPkgItems(n); }} />
                <input type="number" placeholder="Qty" value={it.esx_item_count} onChange={(e) => { const n = [...pkgItems]; n[i].esx_item_count = Number(e.target.value); setPkgItems(n); }} />
              </div>
            ))}
            <button type="button" className="btn btn-ghost btn-sm" onClick={() => setPkgItems([...pkgItems, { esx_item_name: '', esx_item_count: 1 }])}>+ Item</button>
            <button className="btn btn-primary" style={{ display: 'block', marginTop: '1rem' }} onClick={savePackage}>Save Pack</button>
          </div>
          <div className="card" style={{ padding: '1rem' }}>
            <h3>Packs ({packages.length})</h3>
            {packages.map((p) => (
              <div key={String(p.id)} style={{ padding: '0.5rem 0', borderBottom: '1px solid var(--border)', fontSize: '0.85rem' }}>
                <strong>{String(p.name)}</strong> · {String((p.items as unknown[])?.length || 0)} items
              </div>
            ))}
          </div>
        </div>
      )}

      {tab === 'promotions' && (
        <div className="card" style={{ padding: '1.5rem', maxWidth: 560 }}>
          <h2>Promotion Campaign</h2>
          {field('Name', promo.name, (v) => setPromo({ ...promo, name: v }))}
          {field('Description', promo.description, (v) => setPromo({ ...promo, description: v }))}
          <select value={promo.target_type} onChange={(e) => setPromo({ ...promo, target_type: e.target.value })} style={{ marginBottom: '0.75rem' }}>
            <option value="product">Product</option>
            <option value="package">Package</option>
          </select>
          {field('Target ID', promo.target_id, (v) => setPromo({ ...promo, target_id: Number(v) }), 'number')}
          {field('Banner Image URL', promo.banner_image_url, (v) => setPromo({ ...promo, banner_image_url: v }))}
          {field('Show Regular Price', promo.regular_price, (v) => setPromo({ ...promo, regular_price: Number(v) }), 'number')}
          {field('Promo Sale Price', promo.sale_price, (v) => setPromo({ ...promo, sale_price: Number(v) }), 'number')}
          {field('Start (datetime)', promo.start_date, (v) => setPromo({ ...promo, start_date: v }))}
          {field('End (datetime)', promo.end_date, (v) => setPromo({ ...promo, end_date: v }))}
          {field('Max per ID', promo.max_limit_per_id, (v) => setPromo({ ...promo, max_limit_per_id: Number(v) }), 'number')}
          <button className="btn btn-primary" onClick={savePromo}>Save Promotion</button>
          <div style={{ marginTop: '1rem' }}>{promotions.map((p) => <div key={String(p.id)}>{String(p.name)} · until {String(p.end_date)}</div>)}</div>
        </div>
      )}

      {tab === 'milestones' && (
        <div className="card" style={{ padding: '1.5rem', maxWidth: 640 }}>
          <h2>Monthly Milestone (ยอดสะสมไม่ลด — รีเซ็ตทุกเดือนอัตโนมัติ)</h2>
          {field('Event Name', milestone.name, (v) => setMilestone({ ...milestone, name: v }))}
          {field('Month (YYYY-MM)', milestone.month_year, (v) => setMilestone({ ...milestone, month_year: v }))}
          {milestone.tiers.map((t, i) => (
            <div key={i} className="card" style={{ padding: '0.75rem', marginBottom: '0.5rem' }}>
              Tier {i + 1}: threshold ฿
              <input type="number" value={t.threshold_amount} onChange={(e) => { const n = [...milestone.tiers]; n[i].threshold_amount = Number(e.target.value); setMilestone({ ...milestone, tiers: n }); }} />
              reward: <input value={t.reward_name} onChange={(e) => { const n = [...milestone.tiers]; n[i].reward_name = e.target.value; setMilestone({ ...milestone, tiers: n }); }} />
              item: <input value={t.esx_item_name} onChange={(e) => { const n = [...milestone.tiers]; n[i].esx_item_name = e.target.value; setMilestone({ ...milestone, tiers: n }); }} />
              x<input type="number" value={t.esx_item_count} onChange={(e) => { const n = [...milestone.tiers]; n[i].esx_item_count = Number(e.target.value); setMilestone({ ...milestone, tiers: n }); }} />
            </div>
          ))}
          <button type="button" className="btn btn-ghost btn-sm" onClick={() => setMilestone({ ...milestone, tiers: [...milestone.tiers, { ...emptyTier, tier_level: milestone.tiers.length + 1 }] })}>+ Tier</button>
          <button className="btn btn-primary" style={{ marginTop: '1rem' }} onClick={saveMilestone}>Save Milestone</button>
        </div>
      )}

      {tab === 'redeem' && (
        <div className="card" style={{ padding: '1.5rem', maxWidth: 560 }}>
          <h2>Redeem Catalog (แต้มลดเมื่อแลก)</h2>
          {field('Name', redeem.name, (v) => setRedeem({ ...redeem, name: v }))}
          {field('Point Cost', redeem.point_cost, (v) => setRedeem({ ...redeem, point_cost: Number(v) }), 'number')}
          {field('ESX Item', redeem.esx_item_name, (v) => setRedeem({ ...redeem, esx_item_name: v }))}
          {field('Count', redeem.esx_item_count, (v) => setRedeem({ ...redeem, esx_item_count: Number(v) }), 'number')}
          {field('Image URL', redeem.image_url, (v) => setRedeem({ ...redeem, image_url: v }))}
          <button className="btn btn-primary" onClick={saveRedeem}>Save Redeem Item</button>
        </div>
      )}

      {tab === 'banners' && (
        <div className="card" style={{ padding: '1.5rem', maxWidth: 480 }}>
          {field('Title', banner.title, (v) => setBanner({ ...banner, title: v }))}
          {field('Image URL', banner.image_url, (v) => setBanner({ ...banner, image_url: v }))}
          {field('Link URL', banner.link_url, (v) => setBanner({ ...banner, link_url: v }))}
          <button className="btn btn-primary" onClick={saveBanner}>Save Banner</button>
        </div>
      )}

      {tab === 'content' && (
        <div className="card" style={{ padding: '1.5rem', maxWidth: 560 }}>
          <select value={post.post_type} onChange={(e) => setPost({ ...post, post_type: e.target.value })} style={{ marginBottom: '0.75rem' }}>
            <option value="announcement">announcement</option>
            <option value="daily_update">daily_update</option>
            <option value="ad">ad</option>
          </select>
          {field('Title EN', post.title_en, (v) => setPost({ ...post, title_en: v }))}
          {field('Title TH', post.title_th, (v) => setPost({ ...post, title_th: v }))}
          {field('Publish Date', post.publish_date, (v) => setPost({ ...post, publish_date: v }))}
          <button className="btn btn-primary" onClick={savePost}>Save Content</button>
        </div>
      )}
    </div>
  );
}
