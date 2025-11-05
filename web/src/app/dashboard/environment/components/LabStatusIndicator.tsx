/**
 * 实验室在线状态指示器组件
 */

import { Circle, Wifi, WifiOff } from 'lucide-react';

// 简单的 className 合并函数
function cn(...classes: (string | undefined | false)[]) {
  return classes.filter(Boolean).join(' ');
}

interface LabStatusIndicatorProps {
  isOnline?: boolean;
  lastConnectedAt?: string;
  showText?: boolean;
  showTime?: boolean;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function LabStatusIndicator({
  isOnline = false,
  lastConnectedAt,
  showText = true,
  showTime = false,
  size = 'md',
  className,
}: LabStatusIndicatorProps) {
  const sizeClasses = {
    sm: 'h-2 w-2',
    md: 'h-3 w-3',
    lg: 'h-4 w-4',
  };

  const iconSizeClasses = {
    sm: 'h-3 w-3',
    md: 'h-4 w-4',
    lg: 'h-5 w-5',
  };

  const textSizeClasses = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-base',
  };

  const formatTime = (time: string) => {
    const date = new Date(time);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return '刚刚';
    if (minutes < 60) return `${minutes}分钟前`;
    if (hours < 24) return `${hours}小时前`;
    if (days < 7) return `${days}天前`;
    return date.toLocaleDateString('zh-CN');
  };

  return (
    <div className={cn('flex items-center gap-2', className)}>
      {/* 状态指示器 */}
      <div className="relative flex items-center">
        {isOnline ? (
          <>
            {/* 在线：带脉冲动画 */}
            <Circle
              className={cn(
                sizeClasses[size],
                'fill-green-500 text-green-500 dark:fill-green-400 dark:text-green-400'
              )}
            />
            <Circle
              className={cn(
                sizeClasses[size],
                'absolute fill-green-500 text-green-500 dark:fill-green-400 dark:text-green-400 animate-ping opacity-75'
              )}
            />
          </>
        ) : (
          // 离线：灰色圆点
          <Circle
            className={cn(
              sizeClasses[size],
              'fill-gray-400 text-gray-400 dark:fill-gray-600 dark:text-gray-600'
            )}
          />
        )}
      </div>

      {/* 状态文本 */}
      {showText && (
        <div className="flex items-center gap-1.5">
          {isOnline ? (
            <Wifi
              className={cn(
                iconSizeClasses[size],
                'text-green-600 dark:text-green-400'
              )}
            />
          ) : (
            <WifiOff
              className={cn(
                iconSizeClasses[size],
                'text-gray-500 dark:text-gray-400'
              )}
            />
          )}
          <span
            className={cn(
              textSizeClasses[size],
              'font-medium',
              isOnline
                ? 'text-green-700 dark:text-green-400'
                : 'text-gray-600 dark:text-gray-400'
            )}
          >
            {isOnline ? '在线' : '离线'}
          </span>
        </div>
      )}

      {/* 时间信息 */}
      {showTime && lastConnectedAt && (
        <span
          className={cn(
            textSizeClasses[size],
            'text-gray-500 dark:text-gray-400'
          )}
        >
          {isOnline ? '已连接' : formatTime(lastConnectedAt)}
        </span>
      )}
    </div>
  );
}
