import LogoLoading from '@/components/basic/loading';
import { OrbitControls, PerspectiveCamera, useGLTF } from '@react-three/drei';
import { Canvas } from '@react-three/fiber';
import { Suspense } from 'react';

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

// 移液站模型组件 - 放大尺寸，从OSS加载
function LiquidHandlerModel({ position = [0, 0.1, 0] }: PositionProps) {
  const OSS_BASE_URL =
    'https://storage.sciol.ac.cn/library/liquid_transform_xyz/meshes';

  const baseLink = useGLTF(`${OSS_BASE_URL}/base_link.glb`);
  const xLink = useGLTF(`${OSS_BASE_URL}/x_link.glb`);
  const yLink = useGLTF(`${OSS_BASE_URL}/y_link.glb`);
  const zLink = useGLTF(`${OSS_BASE_URL}/z_link.glb`);

  return (
    <group position={position} scale={1.5}>
      <primitive object={baseLink.scene.clone()} />
      <primitive object={xLink.scene.clone()} />
      <primitive object={yLink.scene.clone()} />
      <primitive object={zLink.scene.clone()} />
    </group>
  );
}

// 实验台组件 - 可复用，真实尺寸
function LabBench({
  position = [0, 0, 0],
  width = 5,
  depth = 2.5,
}: LabBenchProps) {
  return (
    <group position={position}>
      {/* 台面 */}
      <mesh position={[0, 0.9, 0]} receiveShadow castShadow>
        <boxGeometry args={[width, 0.05, depth]} />
        <meshStandardMaterial color="#d1d5db" metalness={0.4} roughness={0.3} />
      </mesh>
      {/* 台腿 */}
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
      {/* 下层储物板 */}
      <mesh position={[0, 0.25, 0]} receiveShadow>
        <boxGeometry args={[width - 0.2, 0.03, depth - 0.2]} />
        <meshStandardMaterial color="#9ca3af" metalness={0.3} roughness={0.5} />
      </mesh>
    </group>
  );
}

// 样品架和微孔板
function SampleRack({ position = [0, 0, 0] }: PositionProps) {
  return (
    <group position={position}>
      {/* 微孔板 */}
      <mesh position={[0, 0.02, 0]} castShadow>
        <boxGeometry args={[0.5, 0.04, 0.35]} />
        <meshStandardMaterial color="#1e293b" metalness={0.2} roughness={0.8} />
      </mesh>
      {/* 微孔板上的孔（装饰） */}
      {Array.from({ length: 48 }).map((_, i) => {
        const row = Math.floor(i / 8);
        const col = i % 8;
        return (
          <mesh
            key={i}
            position={[-0.17 + col * 0.05, 0.05, -0.12 + row * 0.05]}
          >
            <cylinderGeometry args={[0.015, 0.015, 0.02, 16]} />
            <meshStandardMaterial
              color={
                i % 3 === 0 ? '#3b82f6' : i % 3 === 1 ? '#8b5cf6' : '#ec4899'
              }
              metalness={0.5}
              roughness={0.3}
              emissive={
                i % 3 === 0 ? '#3b82f6' : i % 3 === 1 ? '#8b5cf6' : '#ec4899'
              }
              emissiveIntensity={0.2}
            />
          </mesh>
        );
      })}
    </group>
  );
}

// 试剂瓶
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
          {/* 瓶身 */}
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
          {/* 瓶盖 */}
          <mesh position={[0, 0.14, 0]} castShadow>
            <cylinderGeometry args={[0.06, 0.06, 0.03, 16]} />
            <meshStandardMaterial
              color="#1f2937"
              metalness={0.8}
              roughness={0.3}
            />
          </mesh>
          {/* 液体 */}
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
      {/* 底座 */}
      <mesh castShadow>
        <cylinderGeometry args={[0.15, 0.15, 0.03, 32]} />
        <meshStandardMaterial color="#374151" metalness={0.6} roughness={0.4} />
      </mesh>
      {/* 支架 */}
      <mesh position={[0, 0.15, 0]} castShadow>
        <cylinderGeometry args={[0.02, 0.02, 0.3, 16]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>
      {/* 移液枪 */}
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

// 显微镜（简化版）
function Microscope({ position = [0, 0, 0] }: PositionProps) {
  return (
    <group position={position} scale={0.8}>
      {/* 底座 */}
      <mesh castShadow>
        <cylinderGeometry args={[0.2, 0.25, 0.05, 32]} />
        <meshStandardMaterial color="#1f2937" metalness={0.8} roughness={0.2} />
      </mesh>
      {/* 支柱 */}
      <mesh position={[0, 0.3, 0]} castShadow>
        <cylinderGeometry args={[0.03, 0.03, 0.5, 16]} />
        <meshStandardMaterial color="#374151" metalness={0.9} roughness={0.1} />
      </mesh>
      {/* 镜头组 */}
      <mesh position={[0, 0.5, 0.1]} rotation={[Math.PI / 6, 0, 0]} castShadow>
        <cylinderGeometry args={[0.08, 0.06, 0.15, 16]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>
      {/* 物镜 */}
      <mesh position={[0, 0.15, 0]} castShadow>
        <cylinderGeometry args={[0.05, 0.04, 0.08, 16]} />
        <meshStandardMaterial color="#4b5563" metalness={0.9} roughness={0.1} />
      </mesh>
    </group>
  );
}

// 电脑显示器 - 真实尺寸（24寸显示器）
function Monitor({
  position = [0, 0, 0],
  rotation = [0, 0, 0],
}: PositionRotationProps) {
  return (
    <group position={position} rotation={rotation}>
      {/* 底座 */}
      <mesh castShadow>
        <cylinderGeometry args={[0.15, 0.18, 0.03, 32]} />
        <meshStandardMaterial color="#1f2937" metalness={0.7} roughness={0.3} />
      </mesh>
      {/* 支架 */}
      <mesh position={[0, 0.2, 0]} castShadow>
        <cylinderGeometry args={[0.02, 0.02, 0.4, 16]} />
        <meshStandardMaterial color="#374151" metalness={0.8} roughness={0.2} />
      </mesh>
      {/* 显示器外框 */}
      <mesh position={[0, 0.5, 0]} castShadow>
        <boxGeometry args={[0.55, 0.35, 0.03]} />
        <meshStandardMaterial color="#111827" metalness={0.3} roughness={0.7} />
      </mesh>
      {/* 屏幕 */}
      <mesh position={[0, 0.5, 0.016]}>
        <boxGeometry args={[0.52, 0.32, 0.001]} />
        <meshStandardMaterial
          color="#3b82f6"
          emissive="#3b82f6"
          emissiveIntensity={0.4}
          metalness={0.1}
          roughness={0.9}
        />
      </mesh>
      {/* 键盘 */}
      <mesh position={[0, 0, 0.25]} castShadow>
        <boxGeometry args={[0.45, 0.02, 0.15]} />
        <meshStandardMaterial color="#1f2937" metalness={0.4} roughness={0.6} />
      </mesh>
      {/* 鼠标 */}
      <mesh position={[0.3, 0.01, 0.25]} castShadow>
        <boxGeometry args={[0.06, 0.015, 0.1]} />
        <meshStandardMaterial color="#374151" metalness={0.5} roughness={0.5} />
      </mesh>
    </group>
  );
}

// AGV 小车 - 大型尺寸（约 120cm × 90cm）带6轴复杂机械臂
function AGVRobot({
  position = [0, 0, 0],
  rotation = [0, 0, 0],
}: PositionRotationProps) {
  return (
    <group position={position} rotation={rotation}>
      {/* 底盘 - 放大1.5倍 */}
      <mesh position={[0, 0.2, 0]} castShadow>
        <boxGeometry args={[1.2, 0.3, 0.9]} />
        <meshStandardMaterial color="#fbbf24" metalness={0.6} roughness={0.3} />
      </mesh>
      {/* 轮子 - 放大 */}
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
      {/* 托盘 - 放大 */}
      <mesh position={[0, 0.38, 0]} castShadow>
        <boxGeometry args={[1.05, 0.04, 0.75]} />
        <meshStandardMaterial color="#d1d5db" metalness={0.5} roughness={0.3} />
      </mesh>

      {/* 机械臂底座 - 大型圆柱 */}
      <mesh position={[0, 0.55, 0]} castShadow>
        <cylinderGeometry args={[0.15, 0.18, 0.25, 16]} />
        <meshStandardMaterial color="#374151" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 底座旋转盘 - 关节1 */}
      <mesh position={[0, 0.68, 0]} castShadow>
        <cylinderGeometry args={[0.13, 0.13, 0.05, 16]} />
        <meshStandardMaterial color="#1f2937" metalness={0.9} roughness={0.1} />
      </mesh>

      {/* 主立柱 - 加高到0.9m */}
      <mesh position={[0, 1.15, 0]} castShadow>
        <cylinderGeometry args={[0.07, 0.07, 0.9, 16]} />
        <meshStandardMaterial color="#4b5563" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 立柱装饰环 */}
      <mesh position={[0, 0.85, 0]} castShadow>
        <cylinderGeometry args={[0.08, 0.08, 0.06, 16]} />
        <meshStandardMaterial color="#1f2937" metalness={0.9} roughness={0.1} />
      </mesh>
      <mesh position={[0, 1.35, 0]} castShadow>
        <cylinderGeometry args={[0.08, 0.08, 0.06, 16]} />
        <meshStandardMaterial color="#1f2937" metalness={0.9} roughness={0.1} />
      </mesh>

      {/* 肩部关节 - 关节2（球形） */}
      <mesh position={[0, 1.6, 0]} castShadow>
        <sphereGeometry args={[0.11, 16, 16]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 肩部连接件 */}
      <mesh position={[0.08, 1.6, 0]} rotation={[0, 0, 0]} castShadow>
        <boxGeometry args={[0.16, 0.12, 0.12]} />
        <meshStandardMaterial color="#374151" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 机械臂第一节（上臂）- 加长 */}
      <mesh position={[0.38, 1.6, 0]} rotation={[0, 0, 0]} castShadow>
        <boxGeometry args={[0.6, 0.09, 0.09]} />
        <meshStandardMaterial color="#3b82f6" metalness={0.7} roughness={0.3} />
      </mesh>

      {/* 上臂装饰条 */}
      <mesh position={[0.38, 1.65, 0]} rotation={[0, 0, 0]} castShadow>
        <boxGeometry args={[0.58, 0.02, 0.095]} />
        <meshStandardMaterial color="#2563eb" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 肘部关节 - 关节3（复杂结构） */}
      <mesh position={[0.68, 1.6, 0]} castShadow>
        <sphereGeometry args={[0.08, 16, 16]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>
      <mesh position={[0.68, 1.6, 0]} rotation={[0, 0, 0]} castShadow>
        <cylinderGeometry args={[0.09, 0.09, 0.08, 16]} />
        <meshStandardMaterial color="#374151" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 机械臂第二节（前臂）- 加长并带角度 */}
      <mesh
        position={[0.95, 1.45, 0]}
        rotation={[0, 0, -Math.PI / 8]}
        castShadow
      >
        <boxGeometry args={[0.5, 0.07, 0.07]} />
        <meshStandardMaterial color="#3b82f6" metalness={0.7} roughness={0.3} />
      </mesh>

      {/* 前臂装饰条 */}
      <mesh
        position={[0.95, 1.48, 0]}
        rotation={[0, 0, -Math.PI / 8]}
        castShadow
      >
        <boxGeometry args={[0.48, 0.02, 0.075]} />
        <meshStandardMaterial color="#2563eb" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 腕部关节 - 关节4（球形旋转） */}
      <mesh position={[1.19, 1.33, 0]} castShadow>
        <sphereGeometry args={[0.06, 16, 16]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 腕部旋转轴 - 关节5 */}
      <mesh
        position={[1.27, 1.33, 0]}
        rotation={[0, 0, Math.PI / 2]}
        castShadow
      >
        <cylinderGeometry args={[0.04, 0.04, 0.15, 16]} />
        <meshStandardMaterial color="#4b5563" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 腕部末端关节 - 关节6 */}
      <mesh position={[1.35, 1.33, 0]} castShadow>
        <sphereGeometry args={[0.05, 16, 16]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 末端执行器底座 */}
      <mesh position={[1.42, 1.33, 0]} castShadow>
        <cylinderGeometry args={[0.05, 0.04, 0.08, 16]} />
        <meshStandardMaterial color="#374151" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 机械臂末端执行器（夹爪）- 更精细 */}
      <group position={[1.48, 1.33, 0]}>
        {/* 左爪 */}
        <mesh position={[-0.03, 0, 0]} castShadow>
          <boxGeometry args={[0.025, 0.15, 0.02]} />
          <meshStandardMaterial
            color="#1f2937"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
        <mesh position={[-0.035, -0.08, 0]} castShadow>
          <boxGeometry args={[0.015, 0.04, 0.025]} />
          <meshStandardMaterial
            color="#3b82f6"
            metalness={0.7}
            roughness={0.3}
          />
        </mesh>

        {/* 右爪 */}
        <mesh position={[0.03, 0, 0]} castShadow>
          <boxGeometry args={[0.025, 0.15, 0.02]} />
          <meshStandardMaterial
            color="#1f2937"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
        <mesh position={[0.035, -0.08, 0]} castShadow>
          <boxGeometry args={[0.015, 0.04, 0.025]} />
          <meshStandardMaterial
            color="#3b82f6"
            metalness={0.7}
            roughness={0.3}
          />
        </mesh>

        {/* 夹爪传感器（红色LED） */}
        <mesh position={[0, 0.05, 0.015]}>
          <sphereGeometry args={[0.01, 8, 8]} />
          <meshStandardMaterial
            color="#ef4444"
            emissive="#ef4444"
            emissiveIntensity={2}
          />
        </mesh>
      </group>

      {/* 立柱顶部装饰 */}
      <mesh position={[0, 1.7, 0]} castShadow>
        <cylinderGeometry args={[0.09, 0.06, 0.2, 16]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>

      {/* 指示灯阵列 */}
      <mesh position={[0, 0.35, -0.45]}>
        <sphereGeometry args={[0.035, 16, 16]} />
        <meshStandardMaterial
          color="#10b981"
          emissive="#10b981"
          emissiveIntensity={1.5}
        />
      </mesh>
      <mesh position={[0.08, 0.25, -0.3]}>
        <sphereGeometry args={[0.025, 16, 16]} />
        <meshStandardMaterial
          color="#3b82f6"
          emissive="#3b82f6"
          emissiveIntensity={1.5}
        />
      </mesh>
    </group>
  );
}

// 储物柜 - 真实尺寸（标准实验室储物柜 90cm宽 × 180cm高 × 60cm深）
function StorageCabinet({ position = [0, 0, 0] }: PositionProps) {
  return (
    <group position={position}>
      {/* 柜体 */}
      <mesh position={[0, 0.9, 0]} castShadow receiveShadow>
        <boxGeometry args={[0.9, 1.8, 0.6]} />
        <meshStandardMaterial color="#9ca3af" metalness={0.2} roughness={0.8} />
      </mesh>
      {/* 柜门 - 上下分层 */}
      {[
        [0.2, 1.35, 0.301], // 右上
        [-0.2, 1.35, 0.301], // 左上
        [0.2, 0.45, 0.301], // 右下
        [-0.2, 0.45, 0.301], // 左下
      ].map((pos, i) => (
        <mesh key={i} position={pos as [number, number, number]} castShadow>
          <boxGeometry args={[0.42, 0.85, 0.02]} />
          <meshStandardMaterial
            color="#d1d5db"
            metalness={0.3}
            roughness={0.6}
          />
        </mesh>
      ))}
      {/* 把手 */}
      {[
        [0.08, 1.35, 0.32],
        [-0.32, 1.35, 0.32],
        [0.08, 0.45, 0.32],
        [-0.32, 0.45, 0.32],
      ].map((pos, i) => (
        <mesh key={i} position={pos as [number, number, number]} castShadow>
          <cylinderGeometry args={[0.012, 0.012, 0.15, 8]} />
          <meshStandardMaterial
            color="#4b5563"
            metalness={0.9}
            roughness={0.1}
          />
        </mesh>
      ))}
    </group>
  );
}

// 实验室架子（开放式货架）
function LabShelf({
  position = [0, 0, 0],
  rotation = [0, 0, 0],
}: PositionRotationProps) {
  return (
    <group position={position} rotation={rotation}>
      {/* 立柱 */}
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
      {/* 层板 */}
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

// 单个试剂瓶（可复用）
function ReagentBottle({
  position = [0, 0, 0],
  color = '#3b82f6',
  size = 'medium',
}: PositionProps & { color?: string; size?: 'small' | 'medium' | 'large' }) {
  const sizes = {
    small: { radius: 0.04, height: 0.12 },
    medium: { radius: 0.06, height: 0.2 },
    large: { radius: 0.08, height: 0.28 },
  };
  const { radius, height } = sizes[size];

  return (
    <group position={position}>
      {/* 瓶身 */}
      <mesh castShadow>
        <cylinderGeometry args={[radius, radius, height, 16]} />
        <meshPhysicalMaterial
          color={color}
          metalness={0.1}
          roughness={0.2}
          transparent
          opacity={0.7}
          transmission={0.5}
        />
      </mesh>
      {/* 瓶盖 */}
      <mesh position={[0, height / 2 + 0.015, 0]} castShadow>
        <cylinderGeometry args={[radius * 0.75, radius * 0.75, 0.03, 16]} />
        <meshStandardMaterial color="#1f2937" metalness={0.8} roughness={0.3} />
      </mesh>
      {/* 液体 */}
      <mesh position={[0, -height * 0.15, 0]}>
        <cylinderGeometry
          args={[radius * 0.9, radius * 0.9, height * 0.7, 16]}
        />
        <meshStandardMaterial
          color={color}
          metalness={0.3}
          roughness={0.1}
          emissive={color}
          emissiveIntensity={0.3}
        />
      </mesh>
    </group>
  );
}

// 试剂瓶架（多瓶组合）
function ReagentRack({ position = [0, 0, 0] }: PositionProps) {
  const colors = [
    '#ef4444',
    '#f59e0b',
    '#10b981',
    '#3b82f6',
    '#8b5cf6',
    '#ec4899',
  ];
  const bottles = [];

  // 3x3 排列
  for (let row = 0; row < 3; row++) {
    for (let col = 0; col < 3; col++) {
      bottles.push({
        pos: [col * 0.15 - 0.15, 0, row * 0.15 - 0.15] as Position3D,
        color: colors[(row * 3 + col) % colors.length],
      });
    }
  }

  return (
    <group position={position}>
      {/* 底托 */}
      <mesh position={[0, -0.05, 0]} castShadow>
        <boxGeometry args={[0.5, 0.02, 0.5]} />
        <meshStandardMaterial color="#9ca3af" metalness={0.5} roughness={0.4} />
      </mesh>
      {/* 试剂瓶 */}
      {bottles.map((bottle, i) => (
        <ReagentBottle
          key={i}
          position={bottle.pos}
          color={bottle.color}
          size="small"
        />
      ))}
    </group>
  );
}

// 离心机
function Centrifuge({ position = [0, 0, 0] }: PositionProps) {
  return (
    <group position={position}>
      {/* 主体 */}
      <mesh castShadow>
        <cylinderGeometry args={[0.15, 0.18, 0.25, 32]} />
        <meshStandardMaterial color="#e5e7eb" metalness={0.6} roughness={0.3} />
      </mesh>
      {/* 盖子 */}
      <mesh position={[0, 0.14, 0]} castShadow>
        <cylinderGeometry args={[0.13, 0.13, 0.03, 32]} />
        <meshStandardMaterial color="#6b7280" metalness={0.8} roughness={0.2} />
      </mesh>
      {/* 控制面板 */}
      <mesh position={[0.16, 0.05, 0]} castShadow>
        <boxGeometry args={[0.05, 0.08, 0.06]} />
        <meshStandardMaterial color="#1f2937" metalness={0.5} roughness={0.5} />
      </mesh>
      {/* 指示灯 */}
      <mesh position={[0.19, 0.08, 0]}>
        <sphereGeometry args={[0.01, 16, 16]} />
        <meshStandardMaterial
          color="#10b981"
          emissive="#10b981"
          emissiveIntensity={2}
        />
      </mesh>
    </group>
  );
}

// 移液枪架（多支）
function PipetteRack({ position = [0, 0, 0] }: PositionProps) {
  const colors = ['#ef4444', '#3b82f6', '#10b981', '#f59e0b'];

  return (
    <group position={position}>
      {/* 底座 */}
      <mesh castShadow>
        <boxGeometry args={[0.3, 0.03, 0.12]} />
        <meshStandardMaterial color="#374151" metalness={0.6} roughness={0.4} />
      </mesh>
      {/* 移液枪 */}
      {colors.map((color, i) => (
        <group key={i} position={[-0.1 + i * 0.07, 0.08, 0]}>
          <mesh rotation={[0, 0, Math.PI / 8]} castShadow>
            <cylinderGeometry args={[0.012, 0.008, 0.15, 12]} />
            <meshStandardMaterial
              color={color}
              metalness={0.7}
              roughness={0.3}
            />
          </mesh>
        </group>
      ))}
    </group>
  );
}

// 培养皿堆叠
function PetriDishStack({ position = [0, 0, 0] }: PositionProps) {
  return (
    <group position={position}>
      {[0, 0.015, 0.03, 0.045, 0.06].map((y, i) => (
        <mesh key={i} position={[0, y, 0]} castShadow>
          <cylinderGeometry args={[0.045, 0.045, 0.012, 32]} />
          <meshPhysicalMaterial
            color="#e0e0e0"
            metalness={0.1}
            roughness={0.1}
            transparent
            opacity={0.8}
            transmission={0.3}
          />
        </mesh>
      ))}
    </group>
  );
}

// 烧杯
function Beaker({
  position = [0, 0, 0],
  color = '#3b82f6',
}: PositionProps & { color?: string }) {
  return (
    <group position={position}>
      {/* 杯体 */}
      <mesh castShadow>
        <cylinderGeometry args={[0.055, 0.05, 0.12, 32, 1, true]} />
        <meshPhysicalMaterial
          color="#ffffff"
          metalness={0.1}
          roughness={0.1}
          transparent
          opacity={0.4}
          transmission={0.7}
          thickness={0.5}
        />
      </mesh>
      {/* 液体 */}
      <mesh position={[0, -0.02, 0]}>
        <cylinderGeometry args={[0.052, 0.048, 0.08, 32]} />
        <meshStandardMaterial
          color={color}
          metalness={0.2}
          roughness={0.3}
          transparent
          opacity={0.7}
          emissive={color}
          emissiveIntensity={0.2}
        />
      </mesh>
    </group>
  );
}

// 完整实验室场景 - 真实比例和布局
function LabScene() {
  return (
    <>
      <PerspectiveCamera makeDefault position={[6, 4, 6]} fov={60} />
      <OrbitControls
        enableZoom={true}
        autoRotate
        autoRotateSpeed={0.5}
        maxPolarAngle={Math.PI / 2.3}
        minPolarAngle={Math.PI / 6}
        maxDistance={12}
        minDistance={3}
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

      {/* 中央工作台 - 沿 X 轴，放置移液站，稍微大一些 */}
      <LabBench position={[0, 0, 0]} width={4.5} depth={2} />
      <LiquidHandlerModel position={[0, 0.95, 0]} />
      <SampleRack position={[-1.3, 0.95, 0.4]} />
      <ReagentBottles position={[1.2, 0.95, -0.5]} />
      <PipetteStand position={[-1.5, 0.95, -0.5]} />

      {/* 左侧工作台 - 沿 Y 轴，加长到5m */}
      <LabBench position={[-4.5, 0, -0.5]} width={1.5} depth={5} />
      <Monitor position={[-4.5, 0.95, -2.7]} rotation={[0, Math.PI / 2, 0]} />
      <Microscope position={[-4.5, 0.95, 0.5]} />
      <ReagentBottles position={[-4.5, 0.95, 1.5]} />

      {/* 右侧工作台 - 沿 Y 轴，加长到5m */}
      <LabBench position={[4.5, 0, -0.5]} width={1.5} depth={5} />
      <Monitor position={[4.5, 0.95, -2.7]} rotation={[0, -Math.PI / 2, 0]} />
      <SampleRack position={[4.5, 0.95, 0.5]} />
      <SampleRack position={[4.5, 0.95, 1.3]} />

      {/* 后方工作台 - 沿 X 轴 */}
      <LabBench position={[0, 0, -4.2]} width={5} depth={1.5} />
      <Monitor position={[-1.8, 0.95, -4.2]} rotation={[0, 0, 0]} />
      <Monitor position={[1.8, 0.95, -4.2]} rotation={[0, 0, 0]} />
      <Microscope position={[0, 0.95, -4.2]} />

      {/* 前方工作台 - 沿 X 轴 */}
      <LabBench position={[-1.5, 0, 3.5]} width={2.5} depth={1.5} />
      <Monitor position={[-1.5, 0.95, 3.5]} rotation={[0, Math.PI, 0]} />
      <ReagentBottles position={[-2.3, 0.95, 3.3]} />

      {/* AGV 小车 - 在通道中 */}
      <AGVRobot position={[-2.5, 0, 1.5]} rotation={[0, Math.PI / 6, 0]} />
      <AGVRobot position={[2.8, 0, -2]} rotation={[0, -Math.PI / 4, 0]} />

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

      {/* 架子上的物品 - 左侧（在架子范围内） */}
      <ReagentRack position={[-6.5, 0.3, 2.5]} />
      <ReagentRack position={[-6.5, 0.7, 2.5]} />
      <ReagentBottle position={[-6.5, 1.1, 2.3]} color="#ef4444" size="large" />
      <ReagentBottle position={[-6.5, 1.1, 2.5]} color="#3b82f6" size="large" />
      <ReagentBottle position={[-6.5, 1.1, 2.7]} color="#10b981" size="large" />
      <PetriDishStack position={[-6.5, 1.5, 2.4]} />
      <PetriDishStack position={[-6.5, 1.5, 2.6]} />

      {/* 架子上的物品 - 右侧（在架子范围内） */}
      <ReagentRack position={[6.5, 0.3, 2.5]} />
      <ReagentRack position={[6.5, 0.7, 2.5]} />
      <ReagentBottle position={[6.5, 1.1, 2.3]} color="#8b5cf6" size="large" />
      <ReagentBottle position={[6.5, 1.1, 2.5]} color="#f59e0b" size="large" />
      <ReagentBottle position={[6.5, 1.1, 2.7]} color="#ec4899" size="large" />
      <PetriDishStack position={[6.5, 1.5, 2.4]} />
      <PetriDishStack position={[6.5, 1.5, 2.6]} />

      {/* 桌面上的额外设备 */}
      {/* 中央桌 */}
      <Centrifuge position={[1.5, 0.95, 0.6]} />
      <PipetteRack position={[-1.2, 0.95, -0.7]} />
      <Beaker position={[1.3, 0.95, -0.7]} color="#3b82f6" />
      <Beaker position={[1.5, 0.95, -0.7]} color="#10b981" />
      <PetriDishStack position={[0.8, 0.95, 0.6]} />

      {/* 左侧桌 - 调整到新的5m桌面范围（Z: -3到2） */}
      <ReagentRack position={[-4.5, 0.95, -2.8]} />
      <Centrifuge position={[-4.5, 0.95, -1.2]} />
      <PetriDishStack position={[-4.2, 0.95, 0.2]} />
      <Beaker position={[-4.7, 0.95, 1.0]} color="#8b5cf6" />

      {/* 右侧桌 - 调整到新的5m桌面范围（Z: -3到2） */}
      <ReagentRack position={[4.5, 0.95, -2.8]} />
      <PipetteRack position={[4.5, 0.95, -1.2]} />
      <Beaker position={[4.2, 0.95, 0.2]} color="#ef4444" />
      <Beaker position={[4.7, 0.95, 1.5]} color="#f59e0b" />

      {/* 后方桌 */}
      <Centrifuge position={[-1.2, 0.95, -4.2]} />
      <Centrifuge position={[1.2, 0.95, -4.2]} />
      <PipetteRack position={[0, 0.95, -4.0]} />
      <ReagentRack position={[-2.2, 0.95, -4.2]} />
      <ReagentRack position={[2.2, 0.95, -4.2]} />

      {/* 前方桌 */}
      <PetriDishStack position={[-1.8, 0.95, 3.3]} />
      <PetriDishStack position={[-1.3, 0.95, 3.3]} />
      <Beaker position={[-2.2, 0.95, 3.6]} color="#ec4899" />

      {/* 地板 - 16m × 12m */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0, 0]} receiveShadow>
        <planeGeometry args={[16, 12]} />
        <meshStandardMaterial color="#e8e8e8" metalness={0.1} roughness={0.7} />
      </mesh>

      {/* 地板网格线 */}
      <gridHelper
        args={[16, 32, '#d1d5db', '#e5e7eb']}
        position={[0, 0.01, 0]}
      />

      {/* 玻璃墙壁 - 后墙（透明蓝色玻璃） */}
      <mesh position={[0, 2.5, -6]} receiveShadow>
        <planeGeometry args={[16, 5]} />
        <meshPhysicalMaterial
          color="#3b82f6"
          metalness={0.1}
          roughness={0.05}
          transparent
          opacity={0.15}
          transmission={0.92}
          thickness={0.8}
          clearcoat={1}
          clearcoatRoughness={0.1}
        />
      </mesh>
      {/* 玻璃框架 - 后墙 */}
      {[-7.5, -2.5, 2.5, 7.5].map((x, i) => (
        <mesh key={`back-frame-v-${i}`} position={[x, 2.5, -6.02]} castShadow>
          <boxGeometry args={[0.08, 5, 0.08]} />
          <meshStandardMaterial
            color="#374151"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
      ))}
      {[0.2, 2.5, 4.8].map((y, i) => (
        <mesh key={`back-frame-h-${i}`} position={[0, y, -6.02]} castShadow>
          <boxGeometry args={[16, 0.06, 0.08]} />
          <meshStandardMaterial
            color="#374151"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
      ))}

      {/* 玻璃墙壁 - 左墙（透明绿色玻璃） */}
      <mesh
        position={[-8, 2.5, 0]}
        rotation={[0, Math.PI / 2, 0]}
        receiveShadow
      >
        <planeGeometry args={[12, 5]} />
        <meshPhysicalMaterial
          color="#10b981"
          metalness={0.1}
          roughness={0.05}
          transparent
          opacity={0.12}
          transmission={0.94}
          thickness={0.8}
          clearcoat={1}
          clearcoatRoughness={0.1}
        />
      </mesh>
      {/* 玻璃框架 - 左墙 */}
      {[-5.5, -1, 3.5, 5.5].map((z, i) => (
        <mesh key={`left-frame-v-${i}`} position={[-8.02, 2.5, z]} castShadow>
          <boxGeometry args={[0.08, 5, 0.08]} />
          <meshStandardMaterial
            color="#374151"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
      ))}

      {/* 玻璃墙壁 - 右墙（透明紫色玻璃） */}
      <mesh
        position={[8, 2.5, 0]}
        rotation={[0, -Math.PI / 2, 0]}
        receiveShadow
      >
        <planeGeometry args={[12, 5]} />
        <meshPhysicalMaterial
          color="#8b5cf6"
          metalness={0.1}
          roughness={0.05}
          transparent
          opacity={0.12}
          transmission={0.94}
          thickness={0.8}
          clearcoat={1}
          clearcoatRoughness={0.1}
        />
      </mesh>
      {/* 玻璃框架 - 右墙 */}
      {[-5.5, -1, 3.5, 5.5].map((z, i) => (
        <mesh key={`right-frame-v-${i}`} position={[8.02, 2.5, z]} castShadow>
          <boxGeometry args={[0.08, 5, 0.08]} />
          <meshStandardMaterial
            color="#374151"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
      ))}

      {/* 玻璃墙壁 - 前墙左侧（带门，橙色玻璃） */}
      <mesh position={[-4, 2.5, 6]} receiveShadow>
        <planeGeometry args={[8, 5]} />
        <meshPhysicalMaterial
          color="#f59e0b"
          metalness={0.1}
          roughness={0.05}
          transparent
          opacity={0.15}
          transmission={0.92}
          thickness={0.8}
          clearcoat={1}
          clearcoatRoughness={0.1}
        />
      </mesh>
      {/* 玻璃墙壁 - 前墙右侧（粉色玻璃） */}
      <mesh position={[4, 2.5, 6]} receiveShadow>
        <planeGeometry args={[8, 5]} />
        <meshPhysicalMaterial
          color="#ec4899"
          metalness={0.1}
          roughness={0.05}
          transparent
          opacity={0.15}
          transmission={0.92}
          thickness={0.8}
          clearcoat={1}
          clearcoatRoughness={0.1}
        />
      </mesh>
      {/* 玻璃框架 - 前墙 */}
      {[-7.5, -4, 0, 4, 7.5].map((x, i) => (
        <mesh key={`front-frame-v-${i}`} position={[x, 2.5, 6.02]} castShadow>
          <boxGeometry args={[0.08, 5, 0.08]} />
          <meshStandardMaterial
            color="#374151"
            metalness={0.8}
            roughness={0.2}
          />
        </mesh>
      ))}

      {/* 天花板 - 半透明玻璃 */}
      <mesh rotation={[Math.PI / 2, 0, 0]} position={[0, 5, 0]} receiveShadow>
        <planeGeometry args={[16, 12]} />
        <meshPhysicalMaterial
          color="#ffffff"
          metalness={0.3}
          roughness={0.2}
          transparent
          opacity={0.5}
          transmission={0.6}
          thickness={1}
        />
      </mesh>

      {/* 天花板框架 */}
      {[-7.5, 0, 7.5].map((x, i) => (
        <mesh key={`ceiling-x-${i}`} position={[x, 5, 0]} castShadow>
          <boxGeometry args={[0.1, 0.1, 12]} />
          <meshStandardMaterial
            color="#6b7280"
            metalness={0.7}
            roughness={0.3}
          />
        </mesh>
      ))}
      {[-5.5, 0, 5.5].map((z, i) => (
        <mesh key={`ceiling-z-${i}`} position={[0, 5, z]} castShadow>
          <boxGeometry args={[16, 0.1, 0.1]} />
          <meshStandardMaterial
            color="#6b7280"
            metalness={0.7}
            roughness={0.3}
          />
        </mesh>
      ))}
    </>
  );
}

// 预加载模型 - 从OSS加载
const OSS_MODEL_BASE =
  'https://storage.sciol.ac.cn/library/liquid_transform_xyz/meshes';
useGLTF.preload(`${OSS_MODEL_BASE}/base_link.glb`);
useGLTF.preload(`${OSS_MODEL_BASE}/x_link.glb`);
useGLTF.preload(`${OSS_MODEL_BASE}/y_link.glb`);
useGLTF.preload(`${OSS_MODEL_BASE}/z_link.glb`);

// 加载占位组件
function LoadingFallback() {
  return (
    <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-indigo-50 to-purple-50 dark:from-neutral-900 dark:to-neutral-800">
      <LogoLoading variant="large" animationType="galaxy" />
    </div>
  );
}

// 主组件
export default function LabScene3D() {
  return (
    <div className="h-full w-full">
      <Suspense fallback={<LoadingFallback />}>
        <Canvas
          shadows
          dpr={[1, 2]}
          className="h-full w-full"
          gl={{ antialias: true }}
        >
          <LabScene />
        </Canvas>
      </Suspense>
    </div>
  );
}
