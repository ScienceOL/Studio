'use client';

import { HeroParallax } from '@/components/ui/hero-parallax';
import { useTheme } from '@/hooks/useTheme';
import { products } from './products';

export default function ChatPage() {
  const { isDark } = useTheme();

  return (
    <div className="relative w-full">
      <HeroParallax products={products} />

      <div
        className={`relative py-16 px-4 pb-48 transition-colors duration-300 ${
          isDark ? 'bg-black' : 'bg-white'
        }`}
      >
        <div className="max-w-7xl mx-auto flex flex-col items-center justify-center gap-8">
          <div className="text-center">
            <h2
              className={`text-3xl md:text-4xl font-bold mb-4 transition-colors duration-300 ${
                isDark ? 'text-white' : 'text-gray-900'
              }`}
            >
              准备好开始了吗？
            </h2>
            <p
              className={`text-lg mb-8 transition-colors duration-300 ${
                isDark ? 'text-gray-300' : 'text-gray-600'
              }`}
            >
              点击下方按钮，立即体验 Xyzen AI Agent
            </p>
          </div>

          <button
            onClick={() => {
              window.open(
                'https://www.bohrium.com/apps/xyzen',
                '_blank'
              );
            }}
            className="group relative px-8 py-3 text-lg font-semibold rounded bg-gradient-to-br from-violet-600 to-fuchsia-600 text-white hover:shadow-2xl transition-all duration-300 hover:scale-105"
          >
            开始使用 Xyzen
          </button>
        </div>
      </div>
    </div>
  );
}
