import clsx from 'clsx';
import React from 'react';

/**
 * DepthWordmark
 * A modern, elegant, minimal 3D / depth wordmark using:
 *  - Multi-layer gradient fill (animated) via existing tailwind utilities
 *  - CSS mask & highlight for subtle bevel
 *  - Ambient glow (behind) + soft inner shadow
 *  - Slight parallax tilt on hover (reduced-motion aware)
 */
export interface DepthWordmarkProps {
  text?: string;
  className?: string;
  sizeClassName?: string; // allow overriding size (e.g., text-7xl)
  glow?: boolean;
  bevel?: boolean; // show bevel highlight
  flat?: boolean; // disable extrusion depth
}

const DepthWordmark: React.FC<DepthWordmarkProps> = ({
  text = 'Science OL',
  className,
  sizeClassName = 'text-7xl',
  glow = true,
  bevel = true,
  flat = false,
}) => {
  const gradientClass = 'animate-gradient-flow bg-gradient-flow';

  return (
    <span
      className={clsx(
        'depth-wordmark relative inline-block select-none font-extrabold leading-[0.95] tracking-tight',
        'will-change-transform motion-safe:transition-transform motion-safe:duration-700 motion-safe:ease-[cubic-bezier(.19,1,.22,1)]',
        sizeClassName,
        className
      )}
      data-glow={glow ? 'on' : 'off'}
    >
      <span
        className={clsx(
          'relative z-20 block bg-[length:400%_400%] bg-clip-text text-transparent',
          gradientClass
        )}
      >
        {text}
      </span>
      {bevel && (
        <span
          aria-hidden
          className={clsx(
            'pointer-events-none absolute inset-0 z-30 mix-blend-overlay opacity-60 dark:opacity-40 [mask:linear-gradient(180deg,white_0%,transparent_55%)]'
          )}
        />
      )}
      {glow && (
        <span
          aria-hidden
          className={clsx(
            'pointer-events-none absolute -inset-2 z-0 rounded-[2rem] blur-2xl opacity-60 dark:opacity-40 mix-blend-plus-lighter bg-gradient-to-br from-indigo-400/40 via-sky-400/25 to-purple-500/30'
          )}
        />
      )}
      {!flat && (
        <span
          aria-hidden
          className="extrusion-layers pointer-events-none absolute inset-0 z-0 select-none"
          data-text={text}
        />
      )}
    </span>
  );
};

export default DepthWordmark;
