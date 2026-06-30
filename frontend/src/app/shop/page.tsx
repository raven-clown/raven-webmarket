import { Suspense } from 'react';
import ShopPageInner from './ShopPageInner';

export default function ShopPage() {
  return (
    <Suspense fallback={<div className="container"><p>Loading...</p></div>}>
      <ShopPageInner />
    </Suspense>
  );
}
