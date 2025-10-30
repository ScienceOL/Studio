import { useTheme } from '@/hooks/useUI';
import {
  ComputerDesktopIcon,
  MoonIcon,
  SunIcon,
} from '@heroicons/react/24/outline';
import clsx from 'clsx';

export type ThemeToggleProps = {
  className?: string;
  title?: string;
};

export const ThemeToggle = ({
  className,
  title = '切换主题',
}: ThemeToggleProps) => {
  const { theme, cycleTheme } = useTheme();

  return (
    <button
      className={clsx(
        'group relative flex items-center justify-center',
        'rounded-lg p-2',
        'text-neutral-500 dark:text-neutral-400',
        'transition-all duration-200 ease-in-out',
        'hover:bg-neutral-100 hover:text-neutral-700',
        'dark:hover:bg-neutral-700 dark:hover:text-neutral-200',
        'focus:outline-none focus:ring-2 focus:ring-indigo-500/20',
        'active:scale-95',
        className
      )}
      title={title}
      onClick={cycleTheme}
      aria-label={title}
      type="button"
    >
      {/* 图标容器，添加旋转动画 */}
      <div className="relative h-5 w-5">
        {theme === 'light' && (
          <SunIcon
            className="absolute inset-0 h-5 w-5 animate-in spin-in-180 fade-in duration-300"
            key="sun"
          />
        )}
        {theme === 'dark' && (
          <MoonIcon
            className="absolute inset-0 h-5 w-5 animate-in spin-in-180 fade-in duration-300"
            key="moon"
          />
        )}
        {theme === 'system' && (
          <ComputerDesktopIcon
            className="absolute inset-0 h-5 w-5 animate-in spin-in-180 fade-in duration-300"
            key="system"
          />
        )}
      </div>

      {/* Hover 效果光晕 */}
      <div
        className="absolute inset-0 rounded-lg bg-gradient-to-r from-indigo-500/0
        via-purple-500/0 to-pink-500/0 opacity-0 transition-opacity duration-300
        group-hover:opacity-10"
      />
    </button>
  );
};
