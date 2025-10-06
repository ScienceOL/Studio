import { useEffect, useState } from 'react';
import Logo, { GrayLogo } from '@/assets/Logo';

export interface LogoRevealProps {
  /**
   * Animation duration in seconds
   */
  duration?: number;
  /**
   * Logo size in pixels
   */
  size?: number;
  /**
   * Number of animation cycles (0 or undefined for infinite)
   */
  cycles?: number;
  /**
   * Callback function when animation completes
   */
  onComplete?: () => void;
  /**
   * Optional additional classes
   */
  className?: string;
}

export function LogoReveal({
  duration = 2,
  size = 100,
  cycles,
  onComplete,
  className = '',
}: LogoRevealProps) {
  // Track the progress of the reveal (0 to 1)
  const [revealProgress, setRevealProgress] = useState(0);
  // Track the current cycle number
  const [currentCycle, setCurrentCycle] = useState(1);
  // Track if animation is running
  const [isRunning, setIsRunning] = useState(true);

  useEffect(() => {
    if (!isRunning) return;

    const startTime = Date.now();
    const animationDuration = duration * 1000;

    // Animation frame callback
    const updateProgress = () => {
      const elapsed = Date.now() - startTime;
      const progress = (elapsed % animationDuration) / animationDuration;

      // If we're on a specific cycle, check if we need to stop
      if (cycles && progress < 0.01 && elapsed > animationDuration) {
        if (currentCycle >= cycles) {
          setRevealProgress(1); // Ensure we end at final state
          setIsRunning(false);
          if (onComplete) onComplete();
          return;
        } else {
          // Start new cycle
          setCurrentCycle(currentCycle + 1);
        }
      }

      setRevealProgress(progress);

      if (isRunning) {
        requestAnimationFrame(updateProgress);
      }
    };

    const animationId = requestAnimationFrame(updateProgress);

    // Cleanup animation frame on unmount
    return () => cancelAnimationFrame(animationId);
  }, [duration, onComplete, cycles, currentCycle, isRunning]);

  // Convert progress to a percentage for the clip-path
  const clipHeight = `${(1 - revealProgress) * 100}%`;

  return (
    <div
      className={`relative ${className}`}
      style={{ width: size, height: size }}
    >
      {/* Gray logo (background) */}
      <div className="absolute inset-0">
        <GrayLogo width={size} height={size} />
      </div>

      {/* Colored logo with clip-path animation */}
      <div
        className="absolute inset-0"
        style={{
          clipPath: `inset(${clipHeight} 0 0 0)`, // Clip from bottom, rising up
        }}
      >
        <Logo width={size} height={size} />
      </div>
    </div>
  );
}

export default LogoReveal;
