'use client';

import LogoLoading from '@/components/basic/loading';
import {
  Html,
  OrbitControls,
  PerspectiveCamera,
  useGLTF,
} from '@react-three/drei';
import { Canvas, useFrame } from '@react-three/fiber';
import { Suspense, useRef, useState } from 'react';
import type { Group, Mesh } from 'three';
// 使用原生动画，无需额外依赖
import { getDeviceInfo } from './deviceInfo';
// 导入所有设备组件
import {
  Beaker,
  Monitor,
  PetriDishStack,
  PipetteRack,
  ReagentBottle,
  ReagentRack,
  SampleRack,
  StorageCabinet,
} from './deviceComponents';

type Position3D = [number, number, number];
type Rotation3D = [number, number, number];

interface PositionProps {
  position?: Position3D;
}

interface PositionRotationProps {
  position?: Position3D;
  rotation?: Rotation3D;
}

interface LabBenchProps {
  position?: Position3D;
  width?: number;
  depth?: number;
}

interface ClickableDeviceProps extends PositionProps {
  deviceId: string;
  children: React.ReactNode;
  onDeviceClick: (deviceId: string) => void;
  isHighlighted?: boolean;
  isAnimating?: boolean;
  disabled?: boolean;
}

// 可点击设备包装器
function ClickableDevice({
  deviceId,
  position = [0, 0, 0],
  children,
  onDeviceClick,
  isHighlighted = false,
  isAnimating = false,
  disabled = false,
}: ClickableDeviceProps) {
  const [hovered, setHovered] = useState(false);
  const groupRef = useRef<Group>(null);
  const deviceInfo = getDeviceInfo(deviceId);

  // 高亮动画 - 使用 useFrame 实现平滑过渡
  const targetScale = isHighlighted ? 1.05 : hovered ? 1.02 : 1;
  const currentScale = useRef(1);

  useFrame(() => {
    if (groupRef.current) {
      currentScale.current += (targetScale - currentScale.current) * 0.1;
      groupRef.current.scale.setScalar(currentScale.current);
    }
  });

  // 点击动画
  useFrame((state) => {
    if (groupRef.current && isAnimating) {
      // 简单的脉冲动画
      const pulse = Math.sin(state.clock.elapsedTime * 2) * 0.05;
      groupRef.current.scale.setScalar(1 + pulse);
    }
  });

  return (
    <group
      ref={groupRef}
      position={position}
      onClick={(e) => {
        if (disabled) return;
        e.stopPropagation();
        onDeviceClick(deviceId);
      }}
      onPointerOver={(e) => {
        if (disabled) return;
        e.stopPropagation();
        setHovered(true);
        document.body.style.cursor = 'pointer';
      }}
      onPointerOut={() => {
        if (disabled) return;
        setHovered(false);
        document.body.style.cursor = 'default';
      }}
    >
      {children}

      {/* 高亮光晕效果 */}
      {(hovered || isHighlighted) && (
        <mesh position={[0, 0.5, 0]}>
          <sphereGeometry args={[0.5, 16, 16]} />
          <meshBasicMaterial
            color="#818cf8"
            transparent
            opacity={0.2}
            depthWrite={false}
          />
        </mesh>
      )}

      {/* 悬浮提示 - 只在hover时显示，不在highlighted时显示，避免遮挡场景 */}
      {hovered && !isHighlighted && deviceInfo && (
        <Html position={[0, 1, 0]} center>
          <div className="pointer-events-none px-3 py-2 bg-indigo-600/95 backdrop-blur-sm text-white rounded-lg shadow-xl text-sm whitespace-nowrap animate-fade-in">
            <div className="font-semibold">{deviceInfo.name}</div>
            <div className="text-xs opacity-90">点击查看详情</div>
          </div>
        </Html>
      )}
    </group>
  );
}

// 移液站模型组件 - 支持动画
function LiquidHandlerModel({
  position = [0, 0.1, 0],
  isAnimating = false,
}: PositionProps & { isAnimating?: boolean }) {
  const OSS_BASE_URL =
    'https://storage.sciol.ac.cn/library/liquid_transform_xyz/meshes';

  const baseLink = useGLTF(`${OSS_BASE_URL}/base_link.glb`);
  const xLink = useGLTF(`${OSS_BASE_URL}/x_link.glb`);
  const yLink = useGLTF(`${OSS_BASE_URL}/y_link.glb`);
  const zLink = useGLTF(`${OSS_BASE_URL}/z_link.glb`);

  const xLinkRef = useRef<Group>(null);
  const yLinkRef = useRef<Group>(null);
  const zLinkRef = useRef<Group>(null);

  // 移液站工作动画：X、Y、Z 轴移动
  useFrame((state) => {
    if (isAnimating) {
      const time = state.clock.elapsedTime;
      if (xLinkRef.current) {
        xLinkRef.current.position.x = Math.sin(time * 0.5) * 0.3;
      }
      if (yLinkRef.current) {
        yLinkRef.current.position.y = Math.sin(time * 0.7 + 1) * 0.2;
      }
      if (zLinkRef.current) {
        zLinkRef.current.position.z = Math.sin(time * 0.6 + 2) * 0.25;
      }
    }
  });

  return (
    <group position={position} scale={1.5}>
      <primitive object={baseLink.scene.clone()} />
      <group ref={xLinkRef}>
        <primitive object={xLink.scene.clone()} />
      </group>
      <group ref={yLinkRef}>
        <primitive object={yLink.scene.clone()} />
      </group>
      <group ref={zLinkRef}>
        <primitive object={zLink.scene.clone()} />
      </group>
    </group>
  );
}

// 试剂瓶组
function ReagentBottles({ position = [0, 0, 0] }: PositionProps) {
  const bottles = [
    { pos: [0, 0, 0], color: '#ef4444' },
    { pos: [0.2, 0, 0], color: '#3b82f6' },
    { pos: [0.4, 0, 0], color: '#10b981' },
    { pos: [0, 0, 0.2], color: '#f59e0b' },
    { pos: [0.2, 0, 0.2], color: '#8b5cf6' },
    { pos: [0.4, 0, 0.2], color: '#06b6d4' },
  ];

  return (
    <group position={position}>
      {bottles.map((bottle, i) => (
        <group key={i} position={bottle.pos as [number, number, number]}>
          <mesh castShadow>
            <cylinderGeometry args={[0.08, 0.08, 0.25, 16]} />
            <meshStandardMaterial
              color={bottle.color}
              metalness={0.1}
              roughness={0.2}
              transparent
              opacity={0.7}
            />
          </mesh>
          <mesh position={[0, 0.14, 0]} castShadow>
            <cylinderGeometry args={[0.06, 0.06, 0.03, 16]} />
            <meshStandardMaterial
              color="#1f2937"
              metalness={0.8}
              roughness={0.3}
            />
          </mesh>
          <mesh position={[0, -0.03, 0]}>
            <cylinderGeometry args={[0.075, 0.075, 0.18, 16]} />
            <meshStandardMaterial
              color={bottle.color}
              metalness={0.3}
              roughness={0.1}
              emissive={bottle.color}
              emissiveIntensity={0.3}
            />
          </mesh>
        </group>
      ))}
    </group>
  );
}

// 移液枪架
function PipetteStand({ position = [0, 0, 0] }: PositionProps) {
  return (
    <group position={position}>
      <mesh castShadow>
        <cylinderGeometry args={[0.15, 0.15, 0.03, 32]} />
        <meshStandardMaterial color="#374151" metalness={0.6} roughness={0.4} />
      </mesh>
      <mesh position={[0, 0.15, 0]} castShadow>
        <cylinderGeometry args={[0.02, 0.02, 0.3, 16]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>
      {[0, 0.1, 0.2].map((y, i) => (
        <group key={i} position={[0.08, y, 0]}>
          <mesh rotation={[0, 0, Math.PI / 6]} castShadow>
            <cylinderGeometry args={[0.015, 0.01, 0.2, 16]} />
            <meshStandardMaterial
              color={['#ef4444', '#3b82f6', '#10b981'][i]}
              metalness={0.7}
              roughness={0.3}
            />
          </mesh>
        </group>
      ))}
    </group>
  );
}

// 实验室货架
function LabShelf({
  position = [0, 0, 0],
  rotation = [0, 0, 0],
}: PositionRotationProps) {
  return (
    <group position={position} rotation={rotation}>
      {[
        [-0.4, 0.75, -0.2],
        [0.4, 0.75, -0.2],
        [-0.4, 0.75, 0.2],
        [0.4, 0.75, 0.2],
      ].map((pos, i) => (
        <mesh key={i} position={pos as [number, number, number]} castShadow>
          <boxGeometry args={[0.04, 1.5, 0.04]} />
          <meshStandardMaterial
            color="#6b7280"
            metalness={0.7}
            roughness={0.3}
          />
        </mesh>
      ))}
      {[0.3, 0.7, 1.1, 1.5].map((y, i) => (
        <mesh key={i} position={[0, y, 0]} castShadow receiveShadow>
          <boxGeometry args={[0.9, 0.02, 0.45]} />
          <meshStandardMaterial
            color="#d1d5db"
            metalness={0.3}
            roughness={0.6}
          />
        </mesh>
      ))}
    </group>
  );
}

// 实验台组件
function LabBench({
  position = [0, 0, 0],
  width = 5,
  depth = 2.5,
}: LabBenchProps) {
  return (
    <group position={position}>
      <mesh position={[0, 0.9, 0]} receiveShadow castShadow>
        <boxGeometry args={[width, 0.05, depth]} />
        <meshStandardMaterial color="#d1d5db" metalness={0.4} roughness={0.3} />
      </mesh>
      {[
        [-width / 2 + 0.15, 0.45, -depth / 2 + 0.15],
        [width / 2 - 0.15, 0.45, -depth / 2 + 0.15],
        [-width / 2 + 0.15, 0.45, depth / 2 - 0.15],
        [width / 2 - 0.15, 0.45, depth / 2 - 0.15],
      ].map((pos, i) => (
        <mesh key={i} position={pos as [number, number, number]} castShadow>
          <boxGeometry args={[0.05, 0.9, 0.05]} />
          <meshStandardMaterial
            color="#6b7280"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
      ))}
      <mesh position={[0, 0.25, 0]} receiveShadow>
        <boxGeometry args={[width - 0.2, 0.03, depth - 0.2]} />
        <meshStandardMaterial color="#9ca3af" metalness={0.3} roughness={0.5} />
      </mesh>
    </group>
  );
}

// 显微镜 - 支持动画
function Microscope({
  position = [0, 0, 0],
  isAnimating = false,
}: PositionProps & { isAnimating?: boolean }) {
  const lensRef = useRef<Mesh>(null);

  // 显微镜观察动画：镜头上下移动
  useFrame((state) => {
    if (lensRef.current && isAnimating) {
      const time = state.clock.elapsedTime;
      lensRef.current.position.y = 0.5 + Math.sin(time * 1.5) * 0.05;
    }
  });

  return (
    <group position={position} scale={0.8}>
      <mesh castShadow>
        <cylinderGeometry args={[0.2, 0.25, 0.05, 32]} />
        <meshStandardMaterial color="#1f2937" metalness={0.8} roughness={0.2} />
      </mesh>
      <mesh position={[0, 0.3, 0]} castShadow>
        <cylinderGeometry args={[0.03, 0.03, 0.5, 16]} />
        <meshStandardMaterial color="#374151" metalness={0.9} roughness={0.1} />
      </mesh>
      <mesh
        ref={lensRef}
        position={[0, 0.5, 0.1]}
        rotation={[Math.PI / 6, 0, 0]}
        castShadow
      >
        <cylinderGeometry args={[0.08, 0.06, 0.15, 16]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>
      <mesh position={[0, 0.15, 0]} castShadow>
        <cylinderGeometry args={[0.05, 0.04, 0.08, 16]} />
        <meshStandardMaterial color="#4b5563" metalness={0.9} roughness={0.1} />
      </mesh>
    </group>
  );
}

// AGV机器人 - 支持动画
function AGVRobot({
  position = [0, 0, 0],
  rotation = [0, 0, 0],
  isAnimating = false,
}: PositionRotationProps & { isAnimating?: boolean }) {
  const robotRef = useRef<Group>(null);
  const armRef = useRef<Group>(null);

  // AGV移动和机械臂动画
  useFrame((state) => {
    if (isAnimating) {
      const time = state.clock.elapsedTime;
      if (robotRef.current) {
        // 前后移动
        robotRef.current.position.z = Math.sin(time * 0.3) * 0.5;
      }
      if (armRef.current) {
        // 机械臂摆动
        armRef.current.rotation.y = Math.sin(time * 0.5) * 0.3;
        armRef.current.rotation.z = Math.sin(time * 0.7) * 0.2;
      }
    }
  });

  return (
    <group ref={robotRef} position={position} rotation={rotation}>
      <mesh position={[0, 0.2, 0]} castShadow>
        <boxGeometry args={[1.2, 0.3, 0.9]} />
        <meshStandardMaterial color="#fbbf24" metalness={0.6} roughness={0.3} />
      </mesh>
      {[
        [-0.52, 0.1, 0.38],
        [0.52, 0.1, 0.38],
        [-0.52, 0.1, -0.38],
        [0.52, 0.1, -0.38],
      ].map((pos, i) => (
        <mesh
          key={i}
          position={pos as [number, number, number]}
          rotation={[0, 0, Math.PI / 2]}
          castShadow
        >
          <cylinderGeometry args={[0.12, 0.12, 0.15, 16]} />
          <meshStandardMaterial
            color="#1f2937"
            metalness={0.7}
            roughness={0.4}
          />
        </mesh>
      ))}
      <mesh position={[0, 0.38, 0]} castShadow>
        <boxGeometry args={[1.05, 0.04, 0.75]} />
        <meshStandardMaterial color="#d1d5db" metalness={0.5} roughness={0.3} />
      </mesh>
      <mesh position={[0, 0.55, 0]} castShadow>
        <cylinderGeometry args={[0.15, 0.18, 0.25, 16]} />
        <meshStandardMaterial color="#374151" metalness={0.8} roughness={0.2} />
      </mesh>
      <group ref={armRef} position={[0, 1.15, 0]}>
        <mesh castShadow>
          <cylinderGeometry args={[0.07, 0.07, 0.9, 16]} />
          <meshStandardMaterial
            color="#4b5563"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
      </group>
    </group>
  );
}

// 离心机 - 支持动画
function Centrifuge({
  position = [0, 0, 0],
  isAnimating = false,
}: PositionProps & { isAnimating?: boolean }) {
  const rotorRef = useRef<Mesh>(null);

  // 离心机旋转动画

  return (
    <group position={position}>
      <mesh castShadow>
        <cylinderGeometry args={[0.15, 0.18, 0.25, 32]} />
        <meshStandardMaterial color="#e5e7eb" metalness={0.6} roughness={0.3} />
      </mesh>
      <mesh ref={rotorRef} position={[0, 0.14, 0]} castShadow>
        <cylinderGeometry args={[0.13, 0.13, 0.03, 32]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>
      <mesh position={[0.16, 0.05, 0]} castShadow>
        <boxGeometry args={[0.05, 0.08, 0.06]} />
        <meshStandardMaterial color="#1f2937" metalness={0.5} roughness={0.5} />
      </mesh>
      <mesh position={[0.19, 0.08, 0]}>
        <sphereGeometry args={[0.01, 16, 16]} />
        <meshStandardMaterial
          color={isAnimating ? '#ef4444' : '#10b981'}
          emissive={isAnimating ? '#ef4444' : '#10b981'}
          emissiveIntensity={isAnimating ? 2 : 1}
        />
      </mesh>
    </group>
  );
}

// 其他设备组件（简化版，从原文件导入）
// 为了简化，这里只实现关键的可点击设备
// 其他设备可以从 deviceComponents.tsx 导入

interface InteractiveLabSceneProps {
  onDeviceClick: (deviceId: string) => void;
  highlightedDevice?: string | null;
  animatingDevice?: string | null;
  disabled?: boolean; // 禁用交互（当模态框打开时）
}

// 交互式实验室场景
function InteractiveLabScene({
  onDeviceClick,
  highlightedDevice,
  animatingDevice,
  disabled = false,
}: InteractiveLabSceneProps) {
  return (
    <>
      <PerspectiveCamera makeDefault position={[8, 5, 8]} fov={50} />
      <OrbitControls
        enableZoom={true}
        autoRotate
        autoRotateSpeed={0.5}
        maxPolarAngle={Math.PI / 2.2}
        minPolarAngle={Math.PI / 8}
        maxDistance={15}
        minDistance={5}
        target={[0, 1, 0]}
        maxAzimuthAngle={Math.PI / 2}
        minAzimuthAngle={-Math.PI / 2}
      />

      {/* 环境光照 */}
      <ambientLight intensity={0.6} />
      <directionalLight
        position={[8, 10, 6]}
        intensity={1.5}
        castShadow
        shadow-mapSize={[2048, 2048]}
        shadow-camera-left={-12}
        shadow-camera-right={12}
        shadow-camera-top={12}
        shadow-camera-bottom={-12}
      />
      <directionalLight position={[-5, 8, -5]} intensity={0.4} />

      {/* 天花板灯光 */}
      {[
        [-4, 4.5, -2],
        [4, 4.5, -2],
        [-4, 4.5, 3],
        [4, 4.5, 3],
        [0, 4.5, 0],
      ].map((pos, i) => (
        <spotLight
          key={i}
          position={pos as [number, number, number]}
          angle={0.5}
          penumbra={1}
          intensity={0.8}
          castShadow
          color="#ffffff"
        />
      ))}

      {/* 中央工作台 */}
      <LabBench position={[0, 0, 0]} width={4.5} depth={2} />
      <ClickableDevice
        deviceId="liquid-handler"
        position={[0, 0.95, 0]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'liquid-handler'}
        isAnimating={animatingDevice === 'liquid-handler'}
        disabled={disabled}
      >
        <LiquidHandlerModel
          isAnimating={animatingDevice === 'liquid-handler'}
        />
      </ClickableDevice>
      <SampleRack position={[-1.3, 0.95, 0.4]} />
      <ReagentBottles position={[1.2, 0.95, -0.5]} />
      <PipetteStand position={[-1.5, 0.95, -0.5]} />
      <ClickableDevice
        deviceId="centrifuge"
        position={[1.5, 0.95, 0.6]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'centrifuge'}
        isAnimating={animatingDevice === 'centrifuge'}
        disabled={disabled}
      >
        <Centrifuge isAnimating={animatingDevice === 'centrifuge'} />
      </ClickableDevice>
      <PipetteRack position={[-1.2, 0.95, -0.7]} />
      <Beaker position={[1.3, 0.95, -0.7]} color="#3b82f6" />
      <Beaker position={[1.5, 0.95, -0.7]} color="#10b981" />
      <PetriDishStack position={[0.8, 0.95, 0.6]} />

      {/* 左侧工作台 */}
      <LabBench position={[-4.5, 0, -0.5]} width={1.5} depth={5} />
      <ClickableDevice
        deviceId="monitor"
        position={[-4.5, 0.95, -2.7]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'monitor'}
        disabled={disabled}
      >
        <group rotation={[0, Math.PI / 2, 0]}>
          <mesh castShadow>
            <boxGeometry args={[0.55, 0.35, 0.03]} />
            <meshStandardMaterial
              color="#111827"
              metalness={0.3}
              roughness={0.7}
            />
          </mesh>
        </group>
      </ClickableDevice>
      <ClickableDevice
        deviceId="microscope"
        position={[-4.5, 0.95, 0.5]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'microscope'}
        isAnimating={animatingDevice === 'microscope'}
        disabled={disabled}
      >
        <Microscope isAnimating={animatingDevice === 'microscope'} />
      </ClickableDevice>
      <ReagentBottles position={[-4.5, 0.95, 1.5]} />
      <ReagentRack position={[-4.5, 0.95, -2.8]} />
      <ClickableDevice
        deviceId="centrifuge"
        position={[-4.5, 0.95, -1.2]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'centrifuge'}
        isAnimating={animatingDevice === 'centrifuge'}
        disabled={disabled}
      >
        <Centrifuge isAnimating={animatingDevice === 'centrifuge'} />
      </ClickableDevice>
      <PetriDishStack position={[-4.2, 0.95, 0.2]} />
      <Beaker position={[-4.7, 0.95, 1.0]} color="#8b5cf6" />

      {/* 右侧工作台 */}
      <LabBench position={[4.5, 0, -0.5]} width={1.5} depth={5} />
      <ClickableDevice
        deviceId="monitor"
        position={[4.5, 0.95, -2.7]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'monitor'}
        disabled={disabled}
      >
        <group rotation={[0, -Math.PI / 2, 0]}>
          <mesh castShadow>
            <boxGeometry args={[0.55, 0.35, 0.03]} />
            <meshStandardMaterial
              color="#111827"
              metalness={0.3}
              roughness={0.7}
            />
          </mesh>
        </group>
      </ClickableDevice>
      <SampleRack position={[4.5, 0.95, 0.5]} />
      <SampleRack position={[4.5, 0.95, 1.3]} />
      <ReagentRack position={[4.5, 0.95, -2.8]} />
      <PipetteRack position={[4.5, 0.95, -1.2]} />
      <Beaker position={[4.2, 0.95, 0.2]} color="#ef4444" />
      <Beaker position={[4.7, 0.95, 1.5]} color="#f59e0b" />

      {/* 后方工作台 */}
      <LabBench position={[0, 0, -4.2]} width={5} depth={1.5} />
      <Monitor position={[-1.8, 0.95, -4.2]} rotation={[0, 0, 0]} />
      <Monitor position={[1.8, 0.95, -4.2]} rotation={[0, 0, 0]} />
      <ClickableDevice
        deviceId="microscope"
        position={[0, 0.95, -4.2]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'microscope'}
        isAnimating={animatingDevice === 'microscope'}
        disabled={disabled}
      >
        <Microscope isAnimating={animatingDevice === 'microscope'} />
      </ClickableDevice>
      <ClickableDevice
        deviceId="centrifuge"
        position={[-1.2, 0.95, -4.2]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'centrifuge'}
        isAnimating={animatingDevice === 'centrifuge'}
        disabled={disabled}
      >
        <Centrifuge isAnimating={animatingDevice === 'centrifuge'} />
      </ClickableDevice>
      <ClickableDevice
        deviceId="centrifuge"
        position={[1.2, 0.95, -4.2]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'centrifuge'}
        isAnimating={animatingDevice === 'centrifuge'}
        disabled={disabled}
      >
        <Centrifuge isAnimating={animatingDevice === 'centrifuge'} />
      </ClickableDevice>
      <PipetteRack position={[0, 0.95, -4.0]} />
      <ReagentRack position={[-2.2, 0.95, -4.2]} />
      <ReagentRack position={[2.2, 0.95, -4.2]} />

      {/* 前方工作台 */}
      <LabBench position={[-1.5, 0, 3.5]} width={2.5} depth={1.5} />
      <Monitor position={[-1.5, 0.95, 3.5]} rotation={[0, Math.PI, 0]} />
      <ReagentBottles position={[-2.3, 0.95, 3.3]} />
      <PetriDishStack position={[-1.8, 0.95, 3.3]} />
      <PetriDishStack position={[-1.3, 0.95, 3.3]} />
      <Beaker position={[-2.2, 0.95, 3.6]} color="#ec4899" />

      {/* AGV 小车 */}
      <ClickableDevice
        deviceId="agv-robot"
        position={[-2.5, 0, 1.5]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'agv-robot'}
        isAnimating={animatingDevice === 'agv-robot'}
        disabled={disabled}
      >
        <AGVRobot
          rotation={[0, Math.PI / 6, 0]}
          isAnimating={animatingDevice === 'agv-robot'}
        />
      </ClickableDevice>
      <ClickableDevice
        deviceId="agv-robot"
        position={[2.8, 0, -2]}
        onDeviceClick={onDeviceClick}
        isHighlighted={highlightedDevice === 'agv-robot'}
        isAnimating={animatingDevice === 'agv-robot'}
        disabled={disabled}
      >
        <AGVRobot
          rotation={[0, -Math.PI / 4, 0]}
          isAnimating={animatingDevice === 'agv-robot'}
        />
      </ClickableDevice>

      {/* 储物柜 - 靠墙排列 */}
      <StorageCabinet position={[-6.5, 0, -3.5]} />
      <StorageCabinet position={[-6.5, 0, -1.5]} />
      <StorageCabinet position={[-6.5, 0, 0.5]} />
      <StorageCabinet position={[6.5, 0, -3.5]} />
      <StorageCabinet position={[6.5, 0, -1.5]} />
      <StorageCabinet position={[6.5, 0, 0.5]} />

      {/* 开放式货架 - 摆满试剂 */}
      <LabShelf position={[-6.5, 0, 2.5]} rotation={[0, Math.PI / 2, 0]} />
      <LabShelf position={[6.5, 0, 2.5]} rotation={[0, -Math.PI / 2, 0]} />

      {/* 架子上的物品 - 左侧 */}
      <ReagentRack position={[-6.5, 0.3, 2.5]} />
      <ReagentRack position={[-6.5, 0.7, 2.5]} />
      <ReagentBottle position={[-6.5, 1.1, 2.3]} color="#ef4444" size="large" />
      <ReagentBottle position={[-6.5, 1.1, 2.5]} color="#3b82f6" size="large" />
      <ReagentBottle position={[-6.5, 1.1, 2.7]} color="#10b981" size="large" />
      <PetriDishStack position={[-6.5, 1.5, 2.4]} />
      <PetriDishStack position={[-6.5, 1.5, 2.6]} />

      {/* 架子上的物品 - 右侧 */}
      <ReagentRack position={[6.5, 0.3, 2.5]} />
      <ReagentRack position={[6.5, 0.7, 2.5]} />
      <ReagentBottle position={[6.5, 1.1, 2.3]} color="#8b5cf6" size="large" />
      <ReagentBottle position={[6.5, 1.1, 2.5]} color="#f59e0b" size="large" />
      <ReagentBottle position={[6.5, 1.1, 2.7]} color="#ec4899" size="large" />
      <PetriDishStack position={[6.5, 1.5, 2.4]} />
      <PetriDishStack position={[6.5, 1.5, 2.6]} />

      {/* 地板 */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0, 0]} receiveShadow>
        <planeGeometry args={[16, 12]} />
        <meshStandardMaterial color="#e8e8e8" metalness={0.1} roughness={0.7} />
      </mesh>
      <gridHelper
        args={[16, 32, '#d1d5db', '#e5e7eb']}
        position={[0, 0.01, 0]}
      />
    </>
  );
}

// 加载占位组件
function LoadingFallback() {
  return (
    <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-indigo-50 to-purple-50 dark:from-neutral-900 dark:to-neutral-800">
      <LogoLoading variant="large" animationType="galaxy" />
    </div>
  );
}

// 主组件
export default function InteractiveLabSceneComponent({
  onDeviceClick,
  highlightedDevice,
  animatingDevice,
}: InteractiveLabSceneProps) {
  return (
    <div className="h-full w-full">
      <Suspense fallback={<LoadingFallback />}>
        <Canvas
          shadows
          dpr={[1, 2]}
          className="h-full w-full"
          gl={{ antialias: true }}
        >
          <InteractiveLabScene
            onDeviceClick={onDeviceClick}
            highlightedDevice={highlightedDevice}
            animatingDevice={animatingDevice}
          />
        </Canvas>
      </Suspense>
    </div>
  );
}
