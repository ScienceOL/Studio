import Logo, { GrayLogo } from '@/assets/Logo';
import { useEffect, useState } from 'react';
import { AtomicLogoLoading } from './AdvLoading';
import { GalaxyLoading } from './GalaxyLoading';

export type LogoLoadingVariant = 'small' | 'medium' | 'large';
export type LogoLoadingAnimationType = 'fade' | 'atomic' | 'galaxy';

export interface LogoLoadingProps {
  /**
   * Size variant of the logo loading animation
   * - small: 24px (适用于按钮、列表项、小卡片等局部加载场景)
   * - medium: 56px (适用于卡片内容、对话框、侧边栏等中等区域加载场景)
   * - large: 100px (适用于全屏、页面级加载场景)
   * @default 'medium'
   */
  variant?: LogoLoadingVariant;
  /**
   * Custom logo size in pixels (overrides variant size)
   * 优先级最高，可覆盖 variant 的默认尺寸
   */
  size?: number;
  /**
   * Animation type
   * - fade: 简单的旋转 + 填色动画（默认）
   * - atomic: 高级原子轨道动画（电子公转 + 整体自转）
   * - galaxy: 优雅的星系旋转动画（简约大气，不眩晕）
   * @default 'fade'
   */
  animationType?: LogoLoadingAnimationType;
  /**
   * Animation duration in seconds
   * @default 2
   */
  duration?: number;
  /**
   * Callback function when animation completes
   */
  onComplete?: () => void;
  /**
   * Optional additional classes
   */
  className?: string;
  /**
   * Enable glow effect (only for atomic animation)
   * @default true
   */
  glowEffect?: boolean;
}

/**
 * LogoLoading - Logo 加载动画组件
 *
 * 提供三种动画类型：
 * 1. fade - 简单的旋转 + 填色动画（默认，轻量级）
 * 2. atomic - 高级原子轨道动画（电子公转 + 整体自转，视觉效果更酷炫）
 * 3. galaxy - 优雅的星系旋转动画（简约大气，缓慢优雅，不会让人眩晕）
 *
 * 提供三种尺寸变体：
 *
 * 1. small (24px) - 小尺寸
 *    适用场景：
 *    - 按钮内加载状态（Button loading state）
 *    - 列表项加载指示器（List item loading indicator）
 *    - 小卡片加载状态（Small card loading）
 *    - 图标位置的加载提示（Icon loading placeholder）
 *    - 内联加载提示（Inline loading indicator）
 *
 * 2. medium (56px) - 中尺寸（默认）
 *    适用场景：
 *    - 卡片内容加载（Card content loading）
 *    - 模态对话框加载（Modal dialog loading）
 *    - 侧边栏加载（Sidebar loading）
 *    - 表单区域加载（Form section loading）
 *    - 面板加载（Panel loading）
 *
 * 3. large (100px) - 大尺寸
 *    适用场景：
 *    - 全屏加载页面（Full-page loading）
 *    - 应用初始化加载（App initialization）
 *    - 页面级路由切换（Page-level route transition）
 *    - Splash screen（启动画面）
 *
 * @example
 * // 使用预设尺寸 + 默认动画
 * <LogoLoading variant="small" />
 * <LogoLoading variant="medium" />
 * <LogoLoading variant="large" />
 *
 * // 使用原子动画
 * <LogoLoading animationType="atomic" />
 * <LogoLoading variant="large" animationType="atomic" />
 *
 * // 自定义尺寸（优先级最高）
 * <LogoLoading size={80} />
 * <LogoLoading variant="medium" size={72} animationType="atomic" />
 */
export function LogoLoading({
  variant = 'medium',
  size,
  animationType = 'fade',
  duration = 2,
  onComplete,
  className = '',
  glowEffect = true,
}: LogoLoadingProps) {
  const [showColorLogo, setShowColorLogo] = useState(false);

  // 尺寸映射：优先使用 size，否则根据 variant 使用预设尺寸
  const variantSizeMap: Record<LogoLoadingVariant, number> = {
    small: 24,
    medium: 56,
    large: 100,
  };

  const logoSize = size ?? variantSizeMap[variant];

  useEffect(() => {
    // 只在 fade 动画模式下执行颜色过渡
    if (animationType !== 'fade') return;

    // Transition from gray to color halfway through the animation
    const timer = setTimeout(() => {
      setShowColorLogo(true);

      // Call onComplete after full animation
      if (onComplete) {
        setTimeout(onComplete, duration * 1000 * 0.5);
      }
    }, duration * 1000 * 0.5);

    return () => clearTimeout(timer);
  }, [duration, onComplete, animationType]);

  // 如果选择星系动画，使用 GalaxyLoading 组件
  if (animationType === 'galaxy') {
    return (
      <GalaxyLoading
        variant={variant}
        size={size}
        duration={duration}
        onComplete={onComplete}
        className={className}
      />
    );
  }

  // 如果选择原子动画，使用 AtomicLogoLoading 组件
  if (animationType === 'atomic') {
    return (
      <AtomicLogoLoading
        variant={variant}
        size={size}
        duration={duration}
        onComplete={onComplete}
        className={className}
        glowEffect={glowEffect}
      />
    );
  }

  // 默认使用简单的填色动画
  return (
    <div
      className={`relative ${className}`}
      style={{
        width: logoSize,
        height: logoSize,
        transformOrigin: 'center',
        animation: `spin-logo ${duration}s ease-in-out forwards`,
      }}
    >
      <style>{`
        @keyframes spin-logo {
          0% {
            transform: rotate(0deg);
          }
          100% {
            transform: rotate(360deg);
          }
        }
      `}</style>

      {/* Grayscale Logo - Fades Out */}
      <div
        className="absolute inset-0 h-full w-full"
        style={{
          opacity: showColorLogo ? 0 : 1,
          transition: `opacity ${duration * 0.5}s ease-in-out`,
        }}
      >
        <GrayLogo width={logoSize} height={logoSize} />
      </div>

      {/* Color Logo - Fades In */}
      <div
        className="absolute inset-0 h-full w-full"
        style={{
          opacity: showColorLogo ? 1 : 0,
          transition: `opacity ${duration * 0.5}s ease-in-out`,
        }}
      >
        <Logo width={logoSize} height={logoSize} />
      </div>
    </div>
  );
}

export default LogoLoading;
