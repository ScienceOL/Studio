/**
 * Label 组件 - 使用 TailwindCSS
 */

import clsx from 'clsx';
import { type LabelHTMLAttributes, forwardRef } from 'react';

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
interface LabelProps extends LabelHTMLAttributes<HTMLLabelElement> {}

const Label = forwardRef<HTMLLabelElement, LabelProps>(
  ({ className, ...props }, ref) => {
    return (
      <label
        ref={ref}
        className={clsx(
          'text-sm font-medium text-neutral-700 leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 dark:text-neutral-300',
          className
        )}
        {...props}
      />
    );
  }
);

Label.displayName = 'Label';

export { Label };
