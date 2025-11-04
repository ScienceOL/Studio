/**
 * Select 组件 - 简单实现
 */

import clsx from 'clsx';
import { type SelectHTMLAttributes, forwardRef } from 'react';

// 简单的 Select 组件，不需要 Radix UI
const Select = forwardRef<
  HTMLSelectElement,
  SelectHTMLAttributes<HTMLSelectElement>
>(({ className, ...props }, ref) => (
  <select
    ref={ref}
    className={clsx(
      'flex h-10 w-full items-center justify-between rounded-md border border-neutral-200 bg-white px-3 py-2 text-sm ring-offset-white focus:outline-none focus:ring-2 focus:ring-neutral-950 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:border-neutral-800 dark:bg-neutral-950 dark:ring-offset-neutral-950 dark:focus:ring-neutral-300',
      className
    )}
    {...props}
  />
));
Select.displayName = 'Select';

// 兼容组件，简化使用
const SelectTrigger = Select;
const SelectValue = Select;
const SelectContent = ({ children }: { children: React.ReactNode }) => (
  <>{children}</>
);
const SelectItem = ({
  value,
  children,
}: {
  value: string;
  children: React.ReactNode;
}) => <option value={value}>{children}</option>;

export { Select, SelectContent, SelectItem, SelectTrigger, SelectValue };
