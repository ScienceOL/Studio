'use client';

import { useUI } from '@/hooks/useUI';
import { useEffect, useState } from 'react';
import DeviceDetailModal from './DeviceDetailModal';
import { getAllDeviceIds, getDeviceInfo } from './deviceInfo';
import InteractiveLabScene from './InteractiveLabScene';
import './styles.css';

export default function Lab3DPage() {
  // 初始化主题系统，跟随主页设置
  useUI();

  const [selectedDevice, setSelectedDevice] = useState<string | null>(null);
  const [showDeviceList, setShowDeviceList] = useState(false);
  const [highlightedDevice, setHighlightedDevice] = useState<string | null>(
    null
  );
  const [animatingDevice, setAnimatingDevice] = useState<string | null>(null);

  const deviceIds = getAllDeviceIds();

  // 处理设备点击
  const handleDeviceClick = (deviceId: string) => {
    setSelectedDevice(deviceId);
    setHighlightedDevice(deviceId);
    // 自动开始动画演示
    setAnimatingDevice(deviceId);
  };

  // 当模态框关闭时，停止高亮和动画
  useEffect(() => {
    if (!selectedDevice) {
      setHighlightedDevice(null);
      setAnimatingDevice(null);
    }
  }, [selectedDevice]);

  // 处理动画控制
  const handleToggleAnimation = (deviceId: string) => {
    if (animatingDevice === deviceId) {
      setAnimatingDevice(null);
    } else {
      setAnimatingDevice(deviceId);
    }
  };

  return (
    <div className="relative h-screen w-full bg-gradient-to-br from-neutral-50 to-neutral-100 dark:from-neutral-900 dark:to-black overflow-hidden">
      {/* 3D 交互式场景 */}
      <InteractiveLabScene
        onDeviceClick={handleDeviceClick}
        highlightedDevice={highlightedDevice}
        animatingDevice={animatingDevice}
        disabled={!!selectedDevice}
      />

      {/* 页面标题 */}
      <div className="absolute top-6 left-6 z-10">
        <div className="bg-white/95 dark:bg-neutral-800/95 backdrop-blur-md px-6 py-4 rounded-2xl shadow-2xl border border-neutral-200/50 dark:border-neutral-700/50">
          <h1 className="text-2xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent mb-1">
            3D 智能实验室
          </h1>
          <p className="text-sm text-neutral-600 dark:text-neutral-400">
            Interactive Laboratory Visualization
          </p>
        </div>
      </div>

      {/* 设备列表按钮 */}
      <div className="absolute top-6 right-6 z-10">
        <button
          onClick={() => setShowDeviceList(!showDeviceList)}
          className="bg-indigo-600 hover:bg-indigo-700 text-white px-6 py-3 rounded-xl shadow-lg transition-all duration-300 hover:scale-105 flex items-center gap-2 "
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 6h16M4 12h16m-16 6h16"
            />
          </svg>
          查看设备列表
        </button>
      </div>

      {/* 设备列表侧边栏 */}
      {showDeviceList && (
        <div
          className="absolute top-20 right-6 z-10 w-80"
          style={{ maxHeight: 'calc(100vh - 170px)' }}
        >
          <div
            className="bg-white/95 dark:bg-neutral-800/95 backdrop-blur-md rounded-2xl shadow-2xl border border-neutral-200/50 dark:border-neutral-700/50 overflow-hidden flex flex-col"
            style={{ height: '100%', maxHeight: 'calc(100vh - 170px)' }}
          >
            {/* 头部 */}
            <div className="flex items-center justify-between p-4 pb-3 border-b border-neutral-200/50 dark:border-neutral-700/50 flex-shrink-0">
              <h3 className="text-lg font-bold text-neutral-900 dark:text-white">
                实验室设备
              </h3>
              <button
                onClick={() => setShowDeviceList(false)}
                className="text-neutral-500 hover:text-neutral-700 dark:hover:text-neutral-300"
              >
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            {/* 滚动内容区域 */}
            <div
              className="flex-1 overflow-y-auto custom-scrollbar p-4 pt-3"
              style={{ minHeight: 0, maxHeight: '100%' }}
            >
              <div className="space-y-2">
                {deviceIds.map((deviceId) => {
                  const info = getDeviceInfo(deviceId);
                  if (!info) return null;

                  return (
                    <button
                      key={deviceId}
                      onClick={() => {
                        setSelectedDevice(deviceId);
                        setShowDeviceList(false);
                      }}
                      className="w-full text-left p-3 rounded-lg bg-neutral-50 hover:bg-indigo-50 dark:bg-neutral-700/50 dark:hover:bg-indigo-900/30 transition-colors group"
                    >
                      <div className="font-semibold text-neutral-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400">
                        {info.name}
                      </div>
                      <div className="text-xs text-neutral-500 dark:text-neutral-400 mt-1">
                        {info.nameEn}
                      </div>
                    </button>
                  );
                })}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 操作指南 */}
      <div className="absolute bottom-6 left-6 z-10">
        <div className="bg-white/95 dark:bg-neutral-800/95 backdrop-blur-md px-5 py-3 rounded-xl shadow-lg border border-neutral-200/50 dark:border-neutral-700/50">
          <div className="flex items-center gap-6 text-sm text-neutral-600 dark:text-neutral-300">
            <div className="flex items-center gap-2">
              <span className="text-xl">🖱️</span>
              <span className="font-medium">拖动旋转</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xl">⚙️</span>
              <span className="font-medium">滚轮缩放</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xl">👆</span>
              <span className="font-medium">点击设备查看</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xl">📋</span>
              <span className="font-medium">右上查看列表</span>
            </div>
          </div>
        </div>
      </div>

      {/* 快速访问设备卡片（底部） */}
      <div className="absolute bottom-6 right-6 z-10">
        <div className="bg-white/95 dark:bg-neutral-800/95 backdrop-blur-md px-4 py-3 rounded-xl shadow-lg border border-neutral-200/50 dark:border-neutral-700/50">
          <div className="flex items-center gap-2">
            <div className="flex gap-2">
              {['liquid-handler', 'microscope', 'agv-robot', 'centrifuge'].map(
                (deviceId) => {
                  const info = getDeviceInfo(deviceId);
                  if (!info) return null;

                  return (
                    <button
                      key={deviceId}
                      onClick={() => setSelectedDevice(deviceId)}
                      className="px-3 py-2 text-xs font-medium bg-indigo-100 hover:bg-indigo-200 dark:bg-indigo-900/30 dark:hover:bg-indigo-800/50 text-indigo-700 dark:text-indigo-300 rounded-lg transition-colors"
                      title={info.name}
                    >
                      {info.name.slice(0, 4)}
                    </button>
                  );
                }
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 设备详情模态框 */}
      {selectedDevice && (
        <DeviceDetailModal
          deviceId={selectedDevice}
          onClose={() => setSelectedDevice(null)}
          isAnimating={animatingDevice === selectedDevice}
          onToggleAnimation={() => handleToggleAnimation(selectedDevice)}
        />
      )}
    </div>
  );
}
