'use client';

import LogoLoading from '@/components/basic/loading';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { OrbitControls, PerspectiveCamera } from '@react-three/drei';
import { Canvas } from '@react-three/fiber';
import { Suspense } from 'react';
import { getDeviceInfo } from './deviceInfo';

// 导入设备组件
import {
  AGVRobot,
  Beaker,
  Centrifuge,
  LiquidHandlerModel,
  Microscope,
  Monitor,
  PetriDishStack,
  PipetteRack,
  ReagentBottle,
  ReagentRack,
  SampleRack,
  StorageCabinet,
} from './deviceComponents';

interface DeviceDetailModalProps {
  deviceId: string;
  onClose: () => void;
  isAnimating?: boolean;
  onToggleAnimation?: () => void;
}

// 设备渲染映射
function DeviceRenderer({
  deviceId,
  isAnimating = false,
}: {
  deviceId: string;
  isAnimating?: boolean;
}) {
  const position: [number, number, number] = [0, 0, 0];

  switch (deviceId) {
    case 'liquid-handler':
      return (
        <LiquidHandlerModel position={position} isAnimating={isAnimating} />
      );
    case 'microscope':
      return <Microscope position={position} isAnimating={isAnimating} />;
    case 'monitor':
      return <Monitor position={position} />;
    case 'agv-robot':
      return (
        <AGVRobot
          position={position}
          rotation={[0, Math.PI / 4, 0]}
          isAnimating={isAnimating}
        />
      );
    case 'centrifuge':
      return <Centrifuge position={position} isAnimating={isAnimating} />;
    case 'pipette-rack':
      return <PipetteRack position={position} />;
    case 'beaker':
      return <Beaker position={position} color="#3b82f6" />;
    case 'storage-cabinet':
      return <StorageCabinet position={position} />;
    case 'reagent-rack':
      return <ReagentRack position={position} />;
    case 'reagent-bottle':
      return <ReagentBottle position={position} color="#ef4444" size="large" />;
    case 'petri-dish':
      return <PetriDishStack position={position} />;
    case 'sample-rack':
      return <SampleRack position={position} />;
    default:
      return null;
  }
}

// 根据设备类型调整相机位置
function getCameraPosition(deviceId: string): [number, number, number] {
  const positions: Record<string, [number, number, number]> = {
    'liquid-handler': [0, 2, 3],
    microscope: [0.5, 0.8, 1.2],
    monitor: [0, 0.8, 1.5],
    'agv-robot': [2, 2, 3],
    centrifuge: [0.4, 0.4, 0.8],
    'pipette-rack': [0.3, 0.3, 0.6],
    beaker: [0.3, 0.3, 0.5],
    'storage-cabinet': [0, 2, 3],
    'reagent-rack': [0.5, 0.5, 1],
    'reagent-bottle': [0.3, 0.3, 0.5],
    'petri-dish': [0.2, 0.2, 0.4],
    'sample-rack': [0.5, 0.5, 1],
  };
  return positions[deviceId] || [0, 1, 2];
}

export default function DeviceDetailModal({
  deviceId,
  onClose,
  isAnimating = false,
  onToggleAnimation,
}: DeviceDetailModalProps) {
  const deviceInfo = getDeviceInfo(deviceId);
  const cameraPos = getCameraPosition(deviceId);

  // 支持动画的设备列表
  const animatableDevices = [
    'liquid-handler',
    'microscope',
    'agv-robot',
    'centrifuge',
  ];
  const canAnimate = animatableDevices.includes(deviceId);

  if (!deviceInfo) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
      <div className="relative h-[90vh] w-[90vw] rounded-2xl bg-white dark:bg-neutral-900 shadow-2xl overflow-hidden">
        {/* 关闭按钮 */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 z-10 rounded-full bg-white/90 dark:bg-neutral-800/90 p-2 shadow-lg hover:bg-white dark:hover:bg-neutral-700 transition-colors"
        >
          <XMarkIcon className="h-6 w-6 text-neutral-900 dark:text-neutral-100" />
        </button>

        <div className="flex h-full">
          {/* 左侧：3D 视图 */}
          <div className="flex-1 relative">
            <Suspense
              fallback={
                <div className="flex h-full w-full items-center justify-center">
                  <LogoLoading variant="large" animationType="galaxy" />
                </div>
              }
            >
              <Canvas shadows dpr={[1, 2]} gl={{ antialias: true }}>
                <PerspectiveCamera makeDefault position={cameraPos} fov={50} />
                <OrbitControls
                  enableZoom
                  enablePan
                  autoRotate
                  autoRotateSpeed={2}
                  minDistance={0.5}
                  maxDistance={5}
                />

                {/* 光照 */}
                <ambientLight intensity={0.8} />
                <directionalLight
                  position={[5, 5, 5]}
                  intensity={1.2}
                  castShadow
                  shadow-mapSize={[1024, 1024]}
                />
                <directionalLight position={[-5, 3, -5]} intensity={0.5} />
                <spotLight
                  position={[0, 5, 0]}
                  angle={0.5}
                  penumbra={1}
                  intensity={0.8}
                  castShadow
                />

                {/* 设备模型 */}
                <DeviceRenderer deviceId={deviceId} isAnimating={isAnimating} />

                {/* 地面 */}
                <mesh
                  rotation={[-Math.PI / 2, 0, 0]}
                  position={[0, -0.5, 0]}
                  receiveShadow
                >
                  <planeGeometry args={[10, 10]} />
                  <meshStandardMaterial
                    color="#f5f5f5"
                    metalness={0.1}
                    roughness={0.8}
                  />
                </mesh>

                {/* 背景网格 */}
                <gridHelper
                  args={[10, 20, '#d1d5db', '#e5e7eb']}
                  position={[0, -0.49, 0]}
                />
              </Canvas>
            </Suspense>

            {/* 操作提示 */}
            <div className="absolute bottom-4 left-4 bg-white/90 dark:bg-neutral-800/90 backdrop-blur-sm px-4 py-2 rounded-lg shadow-lg">
              <p className="text-sm text-neutral-600 dark:text-neutral-300">
                🖱️ 拖动旋转 | 滚轮缩放 | 右键平移
              </p>
            </div>
          </div>

          {/* 右侧：设备信息 */}
          <div className="w-96 bg-gradient-to-br from-indigo-50 to-purple-50 dark:from-neutral-800 dark:to-neutral-900 p-8 overflow-y-auto custom-scrollbar">
            <div className="space-y-6">
              {/* 标题 */}
              <div>
                <h2 className="text-3xl font-bold text-neutral-900 dark:text-white mb-2">
                  {deviceInfo.name}
                </h2>
                <p className="text-lg text-neutral-600 dark:text-neutral-400">
                  {deviceInfo.nameEn}
                </p>
              </div>

              {/* 分隔线 */}
              <div className="h-px bg-gradient-to-r from-indigo-200 via-purple-200 to-transparent dark:from-indigo-800 dark:via-purple-800" />

              {/* 描述 */}
              <div>
                <h3 className="text-sm font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-2">
                  设备简介
                </h3>
                <p className="text-neutral-700 dark:text-neutral-300 leading-relaxed">
                  {deviceInfo.description}
                </p>
              </div>

              {/* 规格参数 */}
              {deviceInfo.specs && deviceInfo.specs.length > 0 && (
                <div>
                  <h3 className="text-sm font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-3">
                    技术规格
                  </h3>
                  <ul className="space-y-2">
                    {deviceInfo.specs.map((spec, index) => (
                      <li
                        key={index}
                        className="flex items-start gap-2 text-neutral-700 dark:text-neutral-300"
                      >
                        <span className="text-indigo-600 dark:text-indigo-400 mt-1">
                          ▹
                        </span>
                        <span>{spec}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {/* 使用场景 */}
              {deviceInfo.usage && (
                <div>
                  <h3 className="text-sm font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-2">
                    应用场景
                  </h3>
                  <p className="text-neutral-700 dark:text-neutral-300 leading-relaxed bg-white/50 dark:bg-neutral-800/50 p-4 rounded-lg">
                    {deviceInfo.usage}
                  </p>
                </div>
              )}

              {/* 动画控制按钮 */}
              {canAnimate && onToggleAnimation && (
                <div>
                  <button
                    onClick={onToggleAnimation}
                    className={`w-full px-4 py-3 rounded-lg font-medium transition-all duration-300 ${
                      isAnimating
                        ? 'bg-indigo-600 hover:bg-indigo-700 text-white'
                        : 'bg-indigo-100 hover:bg-indigo-200 dark:bg-indigo-900/30 dark:hover:bg-indigo-800/50 text-indigo-700 dark:text-indigo-300'
                    }`}
                  >
                    <div className="flex items-center justify-center gap-2">
                      {isAnimating ? (
                        <>
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
                              d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z"
                            />
                          </svg>
                          <span>停止动画演示</span>
                        </>
                      ) : (
                        <>
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
                              d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                            />
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                            />
                          </svg>
                          <span>播放动画演示</span>
                        </>
                      )}
                    </div>
                  </button>
                  {isAnimating && (
                    <p className="mt-2 text-xs text-center text-neutral-500 dark:text-neutral-400">
                      正在演示设备工作流程
                    </p>
                  )}
                </div>
              )}

              {/* 装饰性图标 */}
              <div className="pt-6 flex items-center justify-center">
                <div className="flex gap-3">
                  <div className="h-2 w-2 rounded-full bg-indigo-400 dark:bg-indigo-600 animate-pulse" />
                  <div className="h-2 w-2 rounded-full bg-purple-400 dark:bg-purple-600 animate-pulse delay-75" />
                  <div className="h-2 w-2 rounded-full bg-pink-400 dark:bg-pink-600 animate-pulse delay-150" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
