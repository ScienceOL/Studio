// 从 LabScene3D 导出所有设备组件，供 DeviceDetailModal 使用
import { useGLTF } from '@react-three/drei';
import { useFrame } from '@react-three/fiber';
import type { JSX } from 'react';
import { useRef } from 'react';
import type { Group, Mesh } from 'three';

type Position3D = [number, number, number];
type Rotation3D = [number, number, number];

interface PositionProps {
  position?: Position3D;
}

interface PositionRotationProps {
  position?: Position3D;
  rotation?: Rotation3D;
}

// 移液站模型组件
export function LiquidHandlerModel({
  position = [0, 0.1, 0],
  isAnimating = false,
}: PositionProps & { isAnimating?: boolean }): JSX.Element {
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

// 显微镜
export function Microscope({
  position = [0, 0, 0],
  isAnimating = false,
}: PositionProps & { isAnimating?: boolean }): JSX.Element {
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

// 显示器
export function Monitor({
  position = [0, 0, 0],
  rotation = [0, 0, 0],
}: PositionRotationProps): JSX.Element {
  return (
    <group position={position} rotation={rotation}>
      <mesh castShadow>
        <cylinderGeometry args={[0.15, 0.18, 0.03, 32]} />
        <meshStandardMaterial color="#1f2937" metalness={0.7} roughness={0.3} />
      </mesh>
      <mesh position={[0, 0.2, 0]} castShadow>
        <cylinderGeometry args={[0.02, 0.02, 0.4, 16]} />
        <meshStandardMaterial color="#374151" metalness={0.8} roughness={0.2} />
      </mesh>
      <mesh position={[0, 0.5, 0]} castShadow>
        <boxGeometry args={[0.55, 0.35, 0.03]} />
        <meshStandardMaterial color="#111827" metalness={0.3} roughness={0.7} />
      </mesh>
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
      <mesh position={[0, 0, 0.25]} castShadow>
        <boxGeometry args={[0.45, 0.02, 0.15]} />
        <meshStandardMaterial color="#1f2937" metalness={0.4} roughness={0.6} />
      </mesh>
      <mesh position={[0.3, 0.01, 0.25]} castShadow>
        <boxGeometry args={[0.06, 0.015, 0.1]} />
        <meshStandardMaterial color="#374151" metalness={0.5} roughness={0.5} />
      </mesh>
    </group>
  );
}

// AGV机器人（简化版，只显示核心部分）
export function AGVRobot({
  position = [0, 0, 0],
  rotation = [0, 0, 0],
  isAnimating = false,
}: PositionRotationProps & { isAnimating?: boolean }): JSX.Element {
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

// 离心机
export function Centrifuge({
  position = [0, 0, 0],
  isAnimating = false,
}: PositionProps & { isAnimating?: boolean }): JSX.Element {
  const rotorRef = useRef<Mesh>(null);

  // 离心机旋转动画
  useFrame(() => {
    if (rotorRef.current && isAnimating) {
      rotorRef.current.rotation.y += 0.1;
    }
  });

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

// 移液枪架
export function PipetteRack({
  position = [0, 0, 0],
}: PositionProps): JSX.Element {
  const colors = ['#ef4444', '#3b82f6', '#10b981', '#f59e0b'];

  return (
    <group position={position}>
      <mesh castShadow>
        <boxGeometry args={[0.3, 0.03, 0.12]} />
        <meshStandardMaterial color="#374151" metalness={0.6} roughness={0.4} />
      </mesh>
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

// 烧杯
export function Beaker({
  position = [0, 0, 0],
  color = '#3b82f6',
}: PositionProps & { color?: string }): JSX.Element {
  return (
    <group position={position}>
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

// 储物柜
export function StorageCabinet({
  position = [0, 0, 0],
}: PositionProps): JSX.Element {
  return (
    <group position={position}>
      <mesh position={[0, 0.9, 0]} castShadow receiveShadow>
        <boxGeometry args={[0.9, 1.8, 0.6]} />
        <meshStandardMaterial color="#9ca3af" metalness={0.2} roughness={0.8} />
      </mesh>
      {[
        [0.2, 1.35, 0.301],
        [-0.2, 1.35, 0.301],
        [0.2, 0.45, 0.301],
        [-0.2, 0.45, 0.301],
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
    </group>
  );
}

// 试剂瓶架
export function ReagentRack({
  position = [0, 0, 0],
}: PositionProps): JSX.Element {
  const colors = [
    '#ef4444',
    '#f59e0b',
    '#10b981',
    '#3b82f6',
    '#8b5cf6',
    '#ec4899',
  ];
  const bottles = [];

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
      <mesh position={[0, -0.05, 0]} castShadow>
        <boxGeometry args={[0.5, 0.02, 0.5]} />
        <meshStandardMaterial color="#9ca3af" metalness={0.5} roughness={0.4} />
      </mesh>
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

// 试剂瓶
export function ReagentBottle({
  position = [0, 0, 0],
  color = '#3b82f6',
  size = 'medium',
}: PositionProps & {
  color?: string;
  size?: 'small' | 'medium' | 'large';
}): JSX.Element {
  const sizes = {
    small: { radius: 0.04, height: 0.12 },
    medium: { radius: 0.06, height: 0.2 },
    large: { radius: 0.08, height: 0.28 },
  };
  const { radius, height } = sizes[size];

  return (
    <group position={position}>
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
      <mesh position={[0, height / 2 + 0.015, 0]} castShadow>
        <cylinderGeometry args={[radius * 0.75, radius * 0.75, 0.03, 16]} />
        <meshStandardMaterial color="#1f2937" metalness={0.8} roughness={0.3} />
      </mesh>
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

// 培养皿
export function PetriDishStack({
  position = [0, 0, 0],
}: PositionProps): JSX.Element {
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

// 样品架
export function SampleRack({
  position = [0, 0, 0],
}: PositionProps): JSX.Element {
  return (
    <group position={position}>
      <mesh position={[0, 0.02, 0]} castShadow>
        <boxGeometry args={[0.5, 0.04, 0.35]} />
        <meshStandardMaterial color="#1e293b" metalness={0.2} roughness={0.8} />
      </mesh>
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
