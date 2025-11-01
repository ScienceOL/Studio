/**
 * Badge 组件 - 使用 TailwindCSS
 */

import clsx from 'clsx';
import { type HTMLAttributes, forwardRef } from 'react';

interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'outline' | 'secondary' | 'destructive';
}

const Badge = forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = 'default', ...props }, ref) => {
    const variants = {
      default: 'bg-indigo-600 text-white dark:bg-indigo-500',
      outline:
        'border border-neutral-300 bg-white text-neutral-700 dark:border-neutral-700 dark:bg-neutral-800 dark:text-neutral-300',
      secondary:
        'bg-neutral-100 text-neutral-900 dark:bg-neutral-800 dark:text-neutral-200',
      destructive: 'bg-red-600 text-white dark:bg-red-500',
    };

    return (
      <div
        ref={ref}
        className={clsx(
          'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold transition-colors',
          variants[variant],
          className
        )}
        {...props}
      />
    );
  }
);

Badge.displayName = 'Badge';

export { Badge };
