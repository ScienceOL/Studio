import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

export type GalaxyLoadingVariant = 'small' | 'medium' | 'large';

export interface GalaxyLoadingProps {
  variant?: GalaxyLoadingVariant;
  size?: number;
  duration?: number;
  onComplete?: () => void;
  className?: string;
  /**
   * Enable interactive hover tooltips
   * @default true
   */
  interactive?: boolean;
}

type TooltipInfo = {
  title: string;
  description: string;
  color: string;
} | null;

/**
 * GalaxyLoading - Atomic Structure Animation Component
 *
 * An elegant atomic model visualization with interactive features:
 * - Central nucleus with glow effect
 * - Three electron shells rotating at different speeds
 * - Electrons orbiting along elliptical paths
 * - Smooth animations with pause-on-hover interaction
 * - Educational tooltips with standard atomic physics definitions
 */
export function GalaxyLoading({
  variant = 'medium',
  size,
  duration = 8,
  onComplete,
  className = '',
  interactive = true,
}: GalaxyLoadingProps) {
  const [cycleCount, setCycleCount] = useState(0);
  const [hoveredElement, setHoveredElement] = useState<string | null>(null);
  const [tooltipInfo, setTooltipInfo] = useState<TooltipInfo>(null);
  const [tempParticles, setTempParticles] = useState<
    Array<{
      id: number;
      x: number;
      y: number;
      dx: number;
      dy: number;
    }>
  >([]);
  const svgRef = useRef<SVGSVGElement>(null);
  const timeoutsRef = useRef<number[]>([]);

  // Generate unique IDs for gradients/filters to avoid conflicts with multiple instances
  const instanceId = useMemo(() => Math.random().toString(36).substr(2, 9), []);

  const variantSizeMap: Record<GalaxyLoadingVariant, number> = {
    small: 48,
    medium: 96,
    large: 160,
  };

  const finalSize = size ?? variantSizeMap[variant];

  // Cleanup timeouts on unmount to prevent memory leaks
  useEffect(() => {
    return () => {
      timeoutsRef.current.forEach((timeout) => clearTimeout(timeout));
      timeoutsRef.current = [];
    };
  }, []);

  useEffect(() => {
    if (cycleCount > 0 && onComplete) {
      onComplete();
    }
  }, [cycleCount, onComplete]);

  useEffect(() => {
    const timer = setInterval(() => {
      setCycleCount((prev) => prev + 1);
    }, duration * 1000);
    return () => clearInterval(timer);
  }, [duration]);

  // Handle hover events - unified tooltip position
  const handleElementHover = useCallback(
    (elementType: string, info: TooltipInfo) => {
      if (!interactive) return;

      setHoveredElement(elementType);
      setTooltipInfo(info);
    },
    [interactive]
  );

  const handleElementLeave = useCallback(() => {
    if (!interactive) return;
    setHoveredElement(null);
    setTooltipInfo(null);
  }, [interactive]);

  // Handle click on empty space to generate temporary particles
  const handleSvgClick = useCallback(
    (e: React.MouseEvent<SVGSVGElement>) => {
      if (!interactive) return;

      const svg = svgRef.current;
      if (!svg) return;

      // Check if clicking on an element
      const target = e.target as SVGElement;
      if (target.tagName !== 'svg') return;

      const rect = svg.getBoundingClientRect();
      const x = ((e.clientX - rect.left) / rect.width) * 100;
      const y = ((e.clientY - rect.top) / rect.height) * 100;

      // Generate random direction for meteor streak
      const angle = Math.random() * Math.PI * 2;
      const distance = 40 + Math.random() * 40; // Longer distance for meteor effect
      const dx = Math.cos(angle) * distance;
      const dy = Math.sin(angle) * distance;

      const newParticle = {
        id: Date.now() + Math.random(), // Ensure uniqueness
        x,
        y,
        dx,
        dy,
      };

      setTempParticles((prev) => [...prev, newParticle]);

      // Remove particle after animation completes (800ms)
      const timeout = window.setTimeout(() => {
        setTempParticles((prev) => prev.filter((p) => p.id !== newParticle.id));
        // Remove timeout from tracking
        timeoutsRef.current = timeoutsRef.current.filter((t) => t !== timeout);
      }, 800);

      // Track timeout for cleanup
      timeoutsRef.current.push(timeout);
    },
    [interactive]
  );

  return (
    <div
      className={`relative flex items-center justify-center ${className}`}
      style={{
        width: finalSize,
        height: finalSize,
      }}
    >
      <style>{`
        @keyframes galaxy-spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }

        @keyframes reverse-spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(-360deg); }
        }

        @keyframes core-glow {
          0%, 100% {
            opacity: 0.9;
            filter: blur(1px);
          }
          50% {
            opacity: 1;
            filter: blur(2px);
          }
        }

        @keyframes particle-fade {
          0%, 100% { opacity: 0.4; }
          50% { opacity: 1; }
        }

        @keyframes ring-pulse {
          0%, 100% {
            stroke-opacity: 0.2;
            stroke-width: 2;
          }
          50% {
            stroke-opacity: 0.5;
            stroke-width: 2.5;
          }
        }

        @keyframes particle-orbit {
          0%, 100% {
            opacity: 0.5;
            transform: scale(1);
          }
          50% {
            opacity: 1;
            transform: scale(1.3);
          }
        }

        @keyframes tooltip-fade-in {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }

        @keyframes meteor-streak {
          0% {
            opacity: 1;
            transform: translate(0, 0) scale(1);
          }
          100% {
            opacity: 0;
            transform: translate(var(--meteor-dx), var(--meteor-dy)) scale(0.3);
          }
        }

        @keyframes meteor-trail {
          0% {
            opacity: 0.6;
            transform: scaleX(0);
          }
          50% {
            opacity: 0.3;
            transform: scaleX(1);
          }
          100% {
            opacity: 0;
            transform: scaleX(0.5);
          }
        }

        .galaxy-element {
          cursor: pointer;
          transition: filter 0.2s ease;
        }

        .galaxy-element:hover {
          filter: brightness(1.5) drop-shadow(0 0 8px currentColor);
        }

        .galaxy-orbit {
          pointer-events: all;
        }

        .galaxy-orbit:hover,
        .galaxy-orbit:hover * {
          animation-play-state: paused !important;
        }

        .galaxy-particle {
          pointer-events: all;
        }

        .galaxy-particle:hover {
          animation-play-state: paused !important;
        }

        .tooltip-card {
          animation: tooltip-fade-in 0.2s ease-out;
          backdrop-filter: blur(12px);
          box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }
      `}</style>

      <svg
        ref={svgRef}
        width={finalSize}
        height={finalSize}
        viewBox="0 0 100 100"
        className="overflow-visible cursor-pointer"
        onClick={handleSvgClick}
        style={{ pointerEvents: 'all' }}
      >
        <defs>
          {/* Core radial gradient - brighter and more saturated */}
          <radialGradient
            id={`core-gradient-${instanceId}`}
            cx="50%"
            cy="50%"
            r="50%"
          >
            <stop offset="0%" stopColor="#818cf8" stopOpacity="1" />
            <stop offset="40%" stopColor="#a78bfa" stopOpacity="0.9" />
            <stop offset="70%" stopColor="#c084fc" stopOpacity="0.6" />
            <stop offset="100%" stopColor="#e879f9" stopOpacity="0" />
          </radialGradient>

          {/* Particle glow gradient - stronger emission */}
          <radialGradient
            id={`particle-gradient-${instanceId}`}
            cx="50%"
            cy="50%"
            r="50%"
          >
            <stop offset="0%" stopColor="#fbbf24" stopOpacity="1" />
            <stop offset="50%" stopColor="#f59e0b" stopOpacity="0.8" />
            <stop offset="100%" stopColor="#d97706" stopOpacity="0" />
          </radialGradient>

          {/* Orbital ring gradient */}
          <radialGradient
            id={`ring-gradient-${instanceId}`}
            cx="50%"
            cy="50%"
            r="50%"
          >
            <stop offset="0%" stopColor="#60a5fa" stopOpacity="0" />
            <stop offset="50%" stopColor="#818cf8" stopOpacity="0.6" />
            <stop offset="100%" stopColor="#a78bfa" stopOpacity="0.3" />
          </radialGradient>

          {/* Glow filter for enhanced visibility */}
          <filter id={`glow-${instanceId}`}>
            <feGaussianBlur stdDeviation="2" result="coloredBlur" />
            <feMerge>
              <feMergeNode in="coloredBlur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>

          {/* Intense glow filter for nucleus */}
          <filter id={`intense-glow-${instanceId}`}>
            <feGaussianBlur stdDeviation="3" result="coloredBlur" />
            <feMerge>
              <feMergeNode in="coloredBlur" />
              <feMergeNode in="coloredBlur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        {/* Outer electron shell - slowest rotation, elliptical orbit */}
        <g
          className={`galaxy-orbit ${
            hoveredElement === 'outer-orbit' ? 'paused' : ''
          }`}
          style={{
            transformOrigin: '50% 50%',
            animation: `galaxy-spin ${duration * 2}s linear infinite`,
          }}
        >
          {/* Elliptical orbital path - natural representation */}
          <ellipse
            cx="50"
            cy="50"
            rx="42"
            ry="38"
            fill="none"
            stroke={`url(#ring-gradient-${instanceId})`}
            strokeWidth="2"
            strokeDasharray="3 2"
            className="galaxy-element"
            onMouseEnter={() =>
              handleElementHover('outer-orbit', {
                title: 'Outer Electron Shell',
                description:
                  'Valence shell with higher principal quantum number (n=3). Electrons in this shell have greater energy and larger orbital radius.',
                color: '#818cf8',
              })
            }
            onMouseLeave={handleElementLeave}
            style={{
              animation: `ring-pulse ${duration}s ease-in-out infinite`,
              pointerEvents: 'stroke',
            }}
          />
          {/* Outer shell electrons - non-uniform distribution */}
          {[0, 50, 110, 180, 230, 310].map((angle, i) => {
            const rad = (angle * Math.PI) / 180;
            const rx = 42;
            const ry = 38;
            const x = 50 + rx * Math.cos(rad);
            const y = 50 + ry * Math.sin(rad);
            return (
              <circle
                key={`outer-${i}`}
                cx={x}
                cy={y}
                r={2.5 + (i % 2) * 0.5}
                fill={`url(#particle-gradient-${instanceId})`}
                filter={`url(#glow-${instanceId})`}
                className={`galaxy-element galaxy-particle ${
                  hoveredElement === `outer-particle-${i}` ? 'paused' : ''
                }`}
                onMouseEnter={() =>
                  handleElementHover(`outer-particle-${i}`, {
                    title: `Valence Electron #${i + 1}`,
                    description:
                      'Negatively charged fundamental particle orbiting the nucleus. Mass ≈ 9.109×10⁻³¹ kg, charge = -1.602×10⁻¹⁹ C.',
                    color: '#fbbf24',
                  })
                }
                onMouseLeave={handleElementLeave}
                style={{
                  animation: `particle-orbit ${
                    duration * 0.8
                  }s ease-in-out infinite`,
                  animationDelay: `${i * 0.15}s`,
                }}
              />
            );
          })}
        </g>

        {/* Middle electron shell - counter-rotating, tilted elliptical orbit */}
        <g
          className={`galaxy-orbit ${
            hoveredElement === 'middle-orbit' ? 'paused' : ''
          }`}
          style={{
            transformOrigin: '50% 50%',
            animation: `reverse-spin ${duration * 1.5}s linear infinite`,
          }}
        >
          <ellipse
            cx="50"
            cy="50"
            rx="30"
            ry="26"
            fill="none"
            stroke={`url(#ring-gradient-${instanceId})`}
            strokeWidth="2"
            strokeDasharray="2 3"
            transform="rotate(25 50 50)"
            className="galaxy-element"
            onMouseEnter={() =>
              handleElementHover('middle-orbit', {
                title: 'Middle Electron Shell',
                description:
                  'Intermediate energy level (n=2). Electrons occupy this shell with moderate binding energy and orbital stability.',
                color: '#a78bfa',
              })
            }
            onMouseLeave={handleElementLeave}
            style={{
              animation: `ring-pulse ${duration * 0.8}s ease-in-out infinite`,
              pointerEvents: 'stroke',
            }}
          />
          {/* Middle shell electrons - random distribution */}
          {[20, 80, 150, 200, 270, 340].map((angle, i) => {
            const rad = (angle * Math.PI) / 180;
            const rx = 30;
            const ry = 26;
            const x = 50 + rx * Math.cos(rad);
            const y = 50 + ry * Math.sin(rad);
            return (
              <g key={`middle-${i}`} transform="rotate(25 50 50)">
                <circle
                  cx={x}
                  cy={y}
                  r={2 + (i % 3) * 0.3}
                  fill={`url(#particle-gradient-${instanceId})`}
                  filter={`url(#glow-${instanceId})`}
                  className={`galaxy-element galaxy-particle ${
                    hoveredElement === `middle-particle-${i}` ? 'paused' : ''
                  }`}
                  onMouseEnter={() =>
                    handleElementHover(`middle-particle-${i}`, {
                      title: `Electron #${i + 1} (n=2)`,
                      description:
                        'Lepton in the second shell. Exhibits wave-particle duality with de Broglie wavelength λ = h/p.',
                      color: '#f59e0b',
                    })
                  }
                  onMouseLeave={handleElementLeave}
                  style={{
                    animation: `particle-orbit ${
                      duration * 0.6
                    }s ease-in-out infinite`,
                    animationDelay: `${i * 0.12}s`,
                  }}
                />
              </g>
            );
          })}
        </g>

        {/* Inner electron shell - fastest rotation, small elliptical orbit */}
        <g
          className={`galaxy-orbit ${
            hoveredElement === 'inner-orbit' ? 'paused' : ''
          }`}
          style={{
            transformOrigin: '50% 50%',
            animation: `galaxy-spin ${duration}s linear infinite`,
          }}
        >
          <ellipse
            cx="50"
            cy="50"
            rx="20"
            ry="18"
            fill="none"
            stroke={`url(#ring-gradient-${instanceId})`}
            strokeWidth="2"
            strokeDasharray="1.5 2.5"
            transform="rotate(-15 50 50)"
            className="galaxy-element"
            onMouseEnter={() =>
              handleElementHover('inner-orbit', {
                title: 'Inner Electron Shell',
                description:
                  'Ground state orbital (n=1). Electrons here possess the lowest energy level and strongest nuclear binding force.',
                color: '#60a5fa',
              })
            }
            onMouseLeave={handleElementLeave}
            style={{
              animation: `ring-pulse ${duration * 0.6}s ease-in-out infinite`,
              pointerEvents: 'stroke',
            }}
          />
          {/* Inner shell electrons - asymmetric distribution */}
          {[0, 65, 140, 210, 300].map((angle, i) => {
            const rad = (angle * Math.PI) / 180;
            const rx = 20;
            const ry = 18;
            const x = 50 + rx * Math.cos(rad);
            const y = 50 + ry * Math.sin(rad);
            return (
              <g key={`inner-${i}`} transform="rotate(-15 50 50)">
                <circle
                  cx={x}
                  cy={y}
                  r={1.5 + (i % 2) * 0.4}
                  fill={`url(#particle-gradient-${instanceId})`}
                  filter={`url(#glow-${instanceId})`}
                  className={`galaxy-element galaxy-particle ${
                    hoveredElement === `inner-particle-${i}` ? 'paused' : ''
                  }`}
                  onMouseEnter={() =>
                    handleElementHover(`inner-particle-${i}`, {
                      title: `Core Electron #${i + 1}`,
                      description:
                        'Ground state electron (1s orbital). Experiences maximum Coulombic attraction to the nucleus with velocity approaching c/137.',
                      color: '#d97706',
                    })
                  }
                  onMouseLeave={handleElementLeave}
                  style={{
                    animation: `particle-orbit ${
                      duration * 0.5
                    }s ease-in-out infinite`,
                    animationDelay: `${i * 0.1}s`,
                  }}
                />
              </g>
            );
          })}
        </g>

        {/* Atomic nucleus - central core */}
        <g
          className={`galaxy-element ${
            hoveredElement === 'core' ? 'paused' : ''
          }`}
          onMouseEnter={() =>
            handleElementHover('core', {
              title: 'Atomic Nucleus',
              description:
                'Dense central core composed of protons and neutrons. Contains 99.9% of atomic mass within ~10⁻¹⁵ m radius.',
              color: '#fbbf24',
            })
          }
          onMouseLeave={handleElementLeave}
        >
          {/* Outer glow layer */}
          <circle
            cx="50"
            cy="50"
            r="12"
            fill={`url(#core-gradient-${instanceId})`}
            filter={`url(#intense-glow-${instanceId})`}
            style={{
              animation: `core-glow ${duration * 0.5}s ease-in-out infinite`,
            }}
          />
          {/* Middle core layer */}
          <circle
            cx="50"
            cy="50"
            r="6"
            fill="#818cf8"
            filter={`url(#glow-${instanceId})`}
            className="opacity-95"
          />
          {/* Solid nucleus */}
          <circle
            cx="50"
            cy="50"
            r="3"
            fill="#fbbf24"
            className="opacity-100"
          />
        </g>

        {/* Temporary particles generated by user clicks */}
        {tempParticles.map((particle) => {
          return (
            <g key={particle.id}>
              {/* Meteor head */}
              <circle
                cx={particle.x}
                cy={particle.y}
                r="2.5"
                fill={`url(#particle-gradient-${instanceId})`}
                filter={`url(#glow-${instanceId})`}
                style={{
                  animation:
                    'meteor-streak 0.8s cubic-bezier(0.25, 0.1, 0.25, 1) forwards',
                  // @ts-expect-error - CSS custom properties
                  '--meteor-dx': `${particle.dx}px`,
                  '--meteor-dy': `${particle.dy}px`,
                }}
              />
              {/* Meteor trail */}
              <ellipse
                cx={particle.x}
                cy={particle.y}
                rx="15"
                ry="1.5"
                fill={`url(#particle-gradient-${instanceId})`}
                opacity="0.4"
                filter={`url(#glow-${instanceId})`}
                transform={`rotate(${
                  (Math.atan2(particle.dy, particle.dx) * 180) / Math.PI
                } ${particle.x} ${particle.y})`}
                style={{
                  transformOrigin: `${particle.x}px ${particle.y}px`,
                  animation:
                    'meteor-trail 0.8s cubic-bezier(0.25, 0.1, 0.25, 1) forwards',
                }}
              />
            </g>
          );
        })}
      </svg>

      {/* Interactive tooltip card - unified right-side position */}
      {interactive && tooltipInfo && (
        <div
          className="tooltip-card absolute pointer-events-none z-50"
          style={{
            left: '100%',
            top: '50%',
            transform: 'translate(20px, -50%)',
            marginLeft: '0px',
          }}
        >
          <div
            className="px-3 py-2 rounded-lg border"
            style={{
              background: 'rgba(17, 24, 39, 0.75)',
              borderColor: tooltipInfo.color,
              minWidth: '180px',
              maxWidth: '240px',
            }}
          >
            <div
              className="text-xs font-semibold mb-0.5"
              style={{ color: tooltipInfo.color }}
            >
              {tooltipInfo.title}
            </div>
            <div className="text-[10px] text-neutral-400 leading-relaxed">
              {tooltipInfo.description}
            </div>
            {/* Left-pointing triangular indicator */}
            <div
              className="absolute right-full top-1/2"
              style={{
                transform: 'translateY(-50%)',
                width: 0,
                height: 0,
                borderTop: '5px solid transparent',
                borderBottom: '5px solid transparent',
                borderRight: `5px solid ${tooltipInfo.color}`,
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
