export type Locale = 'en' | 'th';

export const defaultLocale: Locale = 'en';

const en = {
  nav: {
    home: 'Home',
    shop: 'Shop',
    milestones: 'Milestones',
    redeem: 'Redeem',
    cart: 'Cart',
    forum: 'Forum',
    news: 'Daily Updates',
    announcements: 'Announcements',
    admin: 'Admin',
    login: 'Login with Discord',
  },
  home: {
    recommended: 'Recommended Promos',
    browseAll: 'Browse All Products',
    noFeatured: 'No featured products yet.',
    announcements: 'Latest Announcements',
    dailyUpdates: 'Daily Updates',
    viewAll: 'View All',
    quickLinks: 'Quick Links',
  },
  shop: {
    title: 'All Products',
    allCategories: 'All Categories',
    search: 'Search products...',
    addedToCart: 'Added to cart',
    loginSuccess: 'Login successful.',
  },
  forum: {
    title: 'Community Forum',
    newThread: 'New Thread',
    loginRequired: 'Login with Discord to post',
    titlePlaceholder: 'Thread title',
    bodyPlaceholder: 'Write your message...',
    post: 'Post',
    reply: 'Reply',
    replies: 'replies',
    pinned: 'Pinned',
    locked: 'Locked',
    noThreads: 'No threads yet. Be the first to post!',
    back: 'Back to Forum',
  },
  news: {
    title: 'Daily Updates',
    subtitle: 'Server patch notes and shop news',
    empty: 'No updates published yet.',
  },
  announcements: {
    title: 'Announcements',
    subtitle: 'Official notices and promotions',
    empty: 'No announcements yet.',
  },
  cookie: {
    title: 'We use cookies',
    body: 'Essential cookies keep you logged in and remember language preferences. Analytics cookies help us improve the shop experience.',
    accept: 'Accept All',
    essential: 'Essential Only',
    learnMore: 'Learn more',
  },
  footer: {
    tagline: 'FiveM ESX Web Shop — secure top-up, shop & redeem',
    links: 'Links',
    legal: 'Legal',
    privacy: 'Privacy & Cookies',
    discord: 'Join Discord',
    rights: 'All rights reserved.',
  },
  common: {
    loading: 'Loading...',
    error: 'Something went wrong',
    readMore: 'Read more',
    addToCart: 'Add to Cart',
  },
};

const th: typeof en = {
  nav: {
    home: 'หน้าแรก',
    shop: 'ร้านค้า',
    milestones: 'สะสมยอด',
    redeem: 'แลกของ',
    cart: 'ตะกร้า',
    forum: 'ฟอรั่ม',
    news: 'อัปเดตรายวัน',
    announcements: 'ประกาศ',
    admin: 'แอดมิน',
    login: 'ล็อกอิน Discord',
  },
  home: {
    recommended: 'สินค้าแนะนำ',
    browseAll: 'ดูสินค้าทั้งหมด',
    noFeatured: 'ยังไม่มีสินค้าแนะนำ',
    announcements: 'ประกาศล่าสุด',
    dailyUpdates: 'อัปเดตรายวัน',
    viewAll: 'ดูทั้งหมด',
    quickLinks: 'ลิงก์ด่วน',
  },
  shop: {
    title: 'สินค้าทั้งหมด',
    allCategories: 'ทุกหมวดหมู่',
    search: 'ค้นหาสินค้า...',
    addedToCart: 'เพิ่มในตะกร้าแล้ว',
    loginSuccess: 'ล็อกอินสำเร็จ',
  },
  forum: {
    title: 'ฟอรั่มชุมชน',
    newThread: 'สร้างกระทู้ใหม่',
    loginRequired: 'ล็อกอิน Discord เพื่อโพสต์',
    titlePlaceholder: 'หัวข้อกระทู้',
    bodyPlaceholder: 'เขียนข้อความ...',
    post: 'โพสต์',
    reply: 'ตอบกลับ',
    replies: 'ตอบกลับ',
    pinned: 'ปักหมุด',
    locked: 'ล็อก',
    noThreads: 'ยังไม่มีกระทู้ เป็นคนแรกที่โพสต์!',
    back: 'กลับไปฟอรั่ม',
  },
  news: {
    title: 'อัปเดตรายวัน',
    subtitle: 'บันทึกแพตช์เซิร์ฟเวอร์และข่าวร้านค้า',
    empty: 'ยังไม่มีอัปเดต',
  },
  announcements: {
    title: 'ประกาศ',
    subtitle: 'ประกาศอย่างเป็นทางการและโปรโมชั่น',
    empty: 'ยังไม่มีประกาศ',
  },
  cookie: {
    title: 'เราใช้คุกกี้',
    body: 'คุกกี้ที่จำเป็นช่วยให้คุณล็อกอินและจำภาษาที่เลือก คุกกี้วิเคราะห์ช่วยปรับปรุงประสบการณ์ร้านค้า',
    accept: 'ยอมรับทั้งหมด',
    essential: 'เฉพาะที่จำเป็น',
    learnMore: 'อ่านเพิ่มเติม',
  },
  footer: {
    tagline: 'ร้านค้า FiveM ESX — เติมเงิน ช้อป และแลกของอย่างปลอดภัย',
    links: 'ลิงก์',
    legal: 'กฎหมาย',
    privacy: 'ความเป็นส่วนตัว & คุกกี้',
    discord: 'เข้าร่วม Discord',
    rights: 'สงวนลิขสิทธิ์',
  },
  common: {
    loading: 'กำลังโหลด...',
    error: 'เกิดข้อผิดพลาด',
    readMore: 'อ่านเพิ่ม',
    addToCart: 'ใส่ตะกร้า',
  },
};

export const translations = { en, th };

export type TranslationKeys = typeof en;

export function t(locale: Locale): TranslationKeys {
  return translations[locale] ?? translations.en;
}

export function pickLocalized<T extends { title_en?: string; title_th?: string; body_en?: string; body_th?: string }>(
  locale: Locale,
  item: T
): { title: string; body: string } {
  const title = locale === 'th' && item.title_th ? item.title_th : (item.title_en || '');
  const body = locale === 'th' && item.body_th ? item.body_th : (item.body_en || '');
  return { title, body };
}
