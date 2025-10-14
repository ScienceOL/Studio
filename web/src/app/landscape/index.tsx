import FeatureOfAbout from '@/app/landscape/FeatureOfAbout';
import Footer from '@/app/landscape/Footer';
import Hero from '@/app/landscape/Hero';
import Sponsor from '@/app/landscape/Sponsor';
import Navbar from '@/app/navbar/Navbar';
import FeatureOfChat from './FeatureOfChat';
import FeatureOfServer from './FeatureOfServer';

export default function LandscapePage() {
  return (
    <div className="flex flex-col">
      <Navbar />
      <Hero />
      <FeatureOfAbout />
      <FeatureOfChat />
      <FeatureOfServer />
      <Sponsor />
      <Footer />
    </div>
  );
}
