import { useEffect, useRef } from 'react';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';

// Geometric constellation background
function ConstellationCanvas() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animationId: number;
    let particles: { x: number; y: number; vx: number; vy: number; r: number }[] = [];

    const resize = () => {
      canvas.width = canvas.offsetWidth * window.devicePixelRatio;
      canvas.height = canvas.offsetHeight * window.devicePixelRatio;
      ctx.scale(window.devicePixelRatio, window.devicePixelRatio);
    };

    const init = () => {
      resize();
      const w = canvas.offsetWidth;
      const h = canvas.offsetHeight;
      particles = Array.from({ length: 50 }, () => ({
        x: Math.random() * w,
        y: Math.random() * h,
        vx: (Math.random() - 0.5) * 0.25,
        vy: (Math.random() - 0.5) * 0.25,
        r: Math.random() * 1.2 + 0.4,
      }));
    };

    const draw = () => {
      const w = canvas.offsetWidth;
      const h = canvas.offsetHeight;
      ctx.clearRect(0, 0, w, h);

      for (const p of particles) {
        p.x += p.vx;
        p.y += p.vy;
        if (p.x < 0) p.x = w;
        if (p.x > w) p.x = 0;
        if (p.y < 0) p.y = h;
        if (p.y > h) p.y = 0;

        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(255,255,255,0.35)';
        ctx.fill();
      }

      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const dx = particles[i].x - particles[j].x;
          const dy = particles[i].y - particles[j].y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < 140) {
            ctx.beginPath();
            ctx.moveTo(particles[i].x, particles[i].y);
            ctx.lineTo(particles[j].x, particles[j].y);
            ctx.strokeStyle = `rgba(255,255,255,${0.05 * (1 - dist / 140)})`;
            ctx.lineWidth = 0.5;
            ctx.stroke();
          }
        }
      }

      animationId = requestAnimationFrame(draw);
    };

    init();
    draw();
    window.addEventListener('resize', init);
    return () => {
      window.removeEventListener('resize', init);
      cancelAnimationFrame(animationId);
    };
  }, []);

  return <canvas ref={canvasRef} className="absolute inset-0 w-full h-full pointer-events-none z-[1]" />;
}

export default function Hero() {
  const { t } = useTranslation();
  const videoRef = useRef<HTMLVideoElement>(null);

  return (
    <section className="relative min-h-screen bg-black overflow-hidden flex flex-col justify-center">
      <video
        ref={videoRef}
        autoPlay
        muted
        loop
        playsInline
        preload="metadata"
        className="absolute inset-0 w-full h-full object-cover opacity-15 z-0 will-change-transform"
        src="https://storage.sciol.ac.cn/library/hero/automation-web.mp4"
      />

      <ConstellationCanvas />

      <div className="relative z-10 mx-auto w-full max-w-screen-2xl px-6 lg:px-8 py-32 sm:py-40">
        <motion.h1
          initial={{ opacity: 0, y: 40 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1, ease: [0.16, 1, 0.3, 1] }}
          className="text-7xl sm:text-9xl lg:text-[10rem] font-black text-white tracking-tighter leading-[0.85] select-none"
        >
          Science
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-white/90 to-white/30">
            OL
          </span>
        </motion.h1>

        <motion.div
          initial={{ opacity: 0, y: 25 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2, ease: [0.16, 1, 0.3, 1] }}
          className="mt-10 max-w-2xl space-y-2"
        >
          <p className="text-base sm:text-lg font-mono text-white/80 tracking-wide">
            {t('landing.hero.tagline1')}
          </p>
          <p className="text-sm sm:text-base font-mono text-white/35 tracking-wide">
            {t('landing.hero.tagline2')}
          </p>
          <p className="text-sm sm:text-base font-mono text-white/35 tracking-wide">
            {t('landing.hero.tagline3')}
          </p>
        </motion.div>

        <motion.div
          initial={{ scaleX: 0 }}
          animate={{ scaleX: 1 }}
          transition={{ duration: 0.7, delay: 0.45, ease: [0.16, 1, 0.3, 1] }}
          className="mt-12 h-px w-24 bg-gradient-to-r from-white/25 to-transparent origin-left"
        />

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, delay: 0.55, ease: [0.16, 1, 0.3, 1] }}
          className="mt-10 flex flex-wrap items-center gap-4"
        >
          <a
            href="https://docs.sciol.ac.cn"
            target="_blank"
            className="rounded-full bg-white px-7 py-3 text-sm font-semibold text-black transition-all hover:bg-white/90 active:scale-[0.97]"
          >
            {t('landing.hero.cta_start')}
          </a>
          <a
            href="https://xyzen.cc/login"
            target="_blank"
            className="rounded-full px-6 py-3 text-sm font-medium text-white/50 ring-1 ring-white/[0.1] transition-all hover:text-white/80 hover:ring-white/[0.2] active:scale-[0.97]"
          >
            {t('landing.hero.cta_xyzen')}
          </a>
        </motion.div>
      </div>

      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1.5, duration: 1 }}
        className="absolute bottom-8 left-1/2 -translate-x-1/2 z-10"
      >
        <motion.div
          animate={{ y: [0, 6, 0] }}
          transition={{ duration: 2, repeat: Infinity }}
          className="h-8 w-5 rounded-full ring-1 ring-white/15 flex items-start justify-center pt-1.5"
        >
          <div className="h-1.5 w-0.5 rounded-full bg-white/30" />
        </motion.div>
      </motion.div>
    </section>
  );
}
