import { Laptop } from 'lucide-react';
import { useEffect, useState } from 'react';

const MobileOverlay = () => {
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const checkDevice = () => {
      // 检查屏幕宽度是否小于 768px (常见的平板/移动端断点)
      const isSmallScreen = window.innerWidth < 768;

      // 检查是否为移动设备用户代理
      const isMobileDevice =
        /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
          navigator.userAgent
        );

      setIsMobile(isSmallScreen || isMobileDevice);
    };

    // 初始检查
    checkDevice();

    // 监听窗口大小变化
    window.addEventListener('resize', checkDevice);

    return () => {
      window.removeEventListener('resize', checkDevice);
    };
  }, []);

  if (!isMobile) return null;

  return (
    <div className="fixed inset-0 z-[9999] flex flex-col items-center justify-center bg-background/95 backdrop-blur-sm p-6 text-center">
      <div className="max-w-md space-y-6">
        <div className="flex justify-center">
          <div className="relative">
            <div className="absolute -inset-1 rounded-full bg-gradient-to-r from-primary to-purple-600 blur opacity-75 animate-pulse"></div>
            <div className="relative bg-background rounded-full p-4 border border-border">
              <Laptop className="w-12 h-12 text-primary" />
            </div>
          </div>
        </div>

        <div className="space-y-2">
          <h2 className="text-2xl font-bold tracking-tight text-foreground">
            请使用桌面端访问
          </h2>
          <p className="text-muted-foreground text-sm leading-relaxed">
            为了获得最佳的使用体验，当前应用仅支持在平板/桌面端设备（PC/Mac）上运行。
            <br />
            请切换至大屏设备继续操作。
          </p>
        </div>

        <div className="pt-4">
          <p className="text-xs text-muted-foreground/50">
            建议分辨率 &gt; 768px
          </p>
        </div>
      </div>
    </div>
  );
};

export default MobileOverlay;
