import FeatureOfAbout from '@/app/landscape/FeatureOfAbout';
import Footer from '@/app/landscape/Footer';
import Hero from '@/app/landscape/Hero';
import Navbar from '@/app/navbar/Navbar';
import Sponsor from '@/app/landscape/Sponsor';

export default function LandscapePage() {
  return (
    <div className="flex flex-col">
      <Navbar />
      <Hero />
      <FeatureOfAbout />
      <Sponsor />
      <Footer />
    </div>
  );
}
