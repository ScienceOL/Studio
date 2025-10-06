import { useXyzen } from '@sciol/xyzen';
import clsx from 'clsx';
import { useState } from 'react';

export default function XyzenButton({
  children,
  className,
  ...props
}: {
  children: React.ReactNode;
} & React.HTMLAttributes<HTMLButtonElement>) {
  const { toggleXyzen, isXyzenOpen } = useXyzen();
  const [isHovering, setIsHovering] = useState(false);

  // 基础类名
  const defaultClass =
    'group relative z-[1] flex items-center justify-center rounded px-3 py-1.5 outline-none transition-all duration-300';

  // 根据侧边栏状态应用不同的文本颜色样式
  const textColorClass = isXyzenOpen
    ? 'text-white' // 当侧边栏打开时，文字为白色
    : 'bg-gradient-to-br from-violet-600 to-fuchsia-600 bg-clip-text text-transparent hover:text-white dark:from-violet-400 dark:to-fuchsia-400';

  // 添加console.log帮助调试
  console.log('XyzenButton rendering, isXyzenOpen:', isXyzenOpen);

  return (
    <>
      <button
        rel="xyzen"
        className={clsx(defaultClass, textColorClass, className)}
        onClick={(e) => {
          console.log('XyzenButton clicked, toggling state');
          e.preventDefault(); // 防止事件冒泡
          e.stopPropagation(); // 防止事件冒泡
          toggleXyzen();
          if (props.onClick) {
            props.onClick(e);
          }
        }}
        onMouseEnter={() => setIsHovering(true)}
        onMouseLeave={() => setIsHovering(false)}
        {...props}
      >
        {children}
        <div
          className={clsx(
            'absolute inset-0 z-[-1] rounded transition-all duration-300',
            // 未激活状态：hover时显示渐变，非hover时隐藏
            !isXyzenOpen && 'opacity-0 group-hover:opacity-100',
            // 激活状态：始终显示背景
            isXyzenOpen && 'opacity-100',
            'animate-gradient-flow'
          )}
          style={{
            backgroundSize: '400% 400%',
            backgroundImage:
              isXyzenOpen && isHovering
                ? // 激活 + 悬停：使用灰色渐变
                  'linear-gradient(90deg, rgba(139, 92, 246, 0.7), rgba(236, 72, 153, 0.7), rgba(245, 158, 11, 0.7))'
                : // 激活但未悬停 或 未激活但悬停：使用彩色渐变
                  'linear-gradient(90deg, #8B5CF6, #EC4899, #F59E0B)',
            animation: 'gradient-flow 8s linear infinite',
          }}
        ></div>
      </button>

      <style
        dangerouslySetInnerHTML={{
          __html: `
        @keyframes gradient-flow {
          0% {
            background-position: 0% 50%;
          }
          50% {
            background-position: 100% 50%;
          }
          100% {
            background-position: 0% 50%;
          }
        }
        .button:hover .icon {
          transform: translateX(4px);
        }
      `,
        }}
      />
    </>
  );
}
