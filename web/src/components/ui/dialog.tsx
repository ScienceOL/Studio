/**
 * Dialog 组件 - 使用 HeadlessUI
 */

import { Dialog as HeadlessDialog, Transition } from '@headlessui/react';
import clsx from 'clsx';
import { Fragment, type ReactNode } from 'react';

interface DialogProps {
  size?: 'md' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl' | '5xl';
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
  className?: string;
}

export function Dialog({
  open,
  onOpenChange,
  children,
  size,
  className,
}: DialogProps) {
  return (
    <Transition show={open} as={Fragment}>
      <HeadlessDialog
        onClose={() => onOpenChange(false)}
        className="relative z-50"
      >
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black/30" aria-hidden="true" />
        </Transition.Child>

        <div className="fixed inset-0 flex items-center justify-center p-4">
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0 scale-95"
            enterTo="opacity-100 scale-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100 scale-100"
            leaveTo="opacity-0 scale-95"
          >
            <HeadlessDialog.Panel
              className={clsx(
                'w-full transform overflow-hidden rounded-2xl bg-white p-0 text-left align-middle shadow-xl transition-all dark:bg-neutral-900 dark:border dark:border-neutral-800',
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
            >
              {children}
            </HeadlessDialog.Panel>
          </Transition.Child>
        </div>
      </HeadlessDialog>
    </Transition>
  );
}

interface DialogContentProps {
  className?: string;
  children: ReactNode;
}

export function DialogContent({ className, children }: DialogContentProps) {
  return <div className={clsx('relative', className)}>{children}</div>;
}

interface DialogHeaderProps {
  className?: string;
  children: ReactNode;
}

export function DialogHeader({ className, children }: DialogHeaderProps) {
  return <div className={clsx('pt-6 pb-4', className)}>{children}</div>;
}

interface DialogTitleProps {
  className?: string;
  children: ReactNode;
}

export function DialogTitle({ className, children }: DialogTitleProps) {
  return (
    <HeadlessDialog.Title
      className={clsx(
        'text-lg font-semibold leading-6 text-neutral-900 dark:text-neutral-100',
        className
      )}
    >
      {children}
    </HeadlessDialog.Title>
  );
}

interface DialogDescriptionProps {
  className?: string;
  children: ReactNode;
}

export function DialogDescription({
  className,
  children,
}: DialogDescriptionProps) {
  return (
    <HeadlessDialog.Description
      className={clsx(
        'mt-2 text-sm text-neutral-500 dark:text-neutral-400',
        className
      )}
    >
      {children}
    </HeadlessDialog.Description>
  );
}

interface DialogFooterProps {
  className?: string;
  children: ReactNode;
}

export function DialogFooter({ className, children }: DialogFooterProps) {
  return (
    <div className={clsx('pt-6 pb-4 flex justify-end gap-3', className)}>
      {children}
    </div>
  );
}
