import Footer from '@/app/landscape/Footer';
import Hero from '@/app/landscape/Hero';
import Navbar from '@/app/navbar/Navbar';
import Ecosystem from './Ecosystem';
import ProductShowcase from './ProductShowcase';

export default function LandscapePage() {
  return (
    <div className="flex flex-col custom-scrollbar bg-black">
      <Navbar />
      <Hero />
      <ProductShowcase />
      <Ecosystem />
      <Footer />
    </div>
  );
}
