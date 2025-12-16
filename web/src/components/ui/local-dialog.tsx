import { Transition } from '@headlessui/react';
import clsx from 'clsx';
import { Fragment, type ReactNode } from 'react';

interface LocalDialogProps {
  size?: 'md' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl' | '5xl';
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
  className?: string;
}

export function LocalDialog({
  open,
  onOpenChange,
  children,
  size,
  className,
}: LocalDialogProps) {
  return (
    <Transition show={open} as={Fragment}>
      <div
        className="absolute inset-0 z-[100] flex items-center justify-center"
        aria-modal="true"
        role="dialog"
      >
        {/* Backdrop */}
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div
            className="absolute inset-0 bg-black/30 backdrop-blur-sm"
            onClick={() => onOpenChange(false)}
            aria-hidden="true"
          />
        </Transition.Child>

        {/* Panel */}
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0 scale-95"
          enterTo="opacity-100 scale-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100 scale-100"
          leaveTo="opacity-0 scale-95"
        >
          <div
            className={clsx(
              'relative w-full transform overflow-hidden rounded-2xl bg-white p-0 text-left align-middle shadow-xl transition-all dark:bg-neutral-900 dark:border dark:border-neutral-800 mx-4',
              {
                'max-w-md': size === 'md',
                'max-w-lg': size === 'lg',
                'max-w-xl': size === 'xl',
                'max-w-2xl': size === '2xl',
                'max-w-3xl': size === '3xl',
                'max-w-4xl': size === '4xl',
                'max-w-5xl': size === '5xl',
              },
              className
            )}
            onClick={(e) => e.stopPropagation()}
          >
            {children}
          </div>
        </Transition.Child>
      </div>
    </Transition>
  );
}

// 重新导出或定义子组件，使其与 Dialog 组件兼容

export function LocalDialogContent({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={clsx('relative', className)}>{children}</div>;
}

export function LocalDialogHeader({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={clsx('pt-6 pb-4', className)}>{children}</div>;
}

export function LocalDialogTitle({ className, children }: { className?: string; children: ReactNode }) {
  return (
    <h3
      className={clsx(
        'text-lg font-semibold leading-6 text-neutral-900 dark:text-neutral-100 px-6',
        className
      )}
    >
      {children}
    </h3>
  );
}

export function LocalDialogDescription({ className, children }: { className?: string; children: ReactNode }) {
  return (
    <p
      className={clsx(
        'mt-2 text-sm text-neutral-500 dark:text-neutral-400 px-6',
        className
      )}
    >
      {children}
    </p>
  );
}

export function LocalDialogFooter({ className, children }: { className?: string; children: ReactNode }) {
  return (
    <div className={clsx('pt-6 pb-4 px-6 flex justify-end gap-3', className)}>
      {children}
    </div>
  );
}
