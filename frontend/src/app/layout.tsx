import type { Metadata } from 'next';
import './globals.css';
import Navbar from '@/components/Navbar';
import Footer from '@/components/Footer';
import AppProviders from '@/components/AppProviders';

export const metadata: Metadata = {
  title: 'Raven Webmarket',
  description: 'FiveM ESX Web Shop — Top-up, Shop, Milestones & Redeem',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AppProviders>
          <Navbar />
          <main className="site-main">{children}</main>
          <Footer />
        </AppProviders>
      </body>
    </html>
  );
}
