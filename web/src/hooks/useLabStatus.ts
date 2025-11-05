/**
 * 🎣 Hook Layer - Lab Status Hook
 *
 * 职责：
 * 1. 提供 React 组件使用的 hook 接口
 * 2. 订阅 Core 层的状态更新
 * 3. 从 Store 读取状态
 *
 * 注意：
 * - 不直接管理 WebSocket 连接
 * - 所有业务逻辑在 Core 层
 * - 所有状态在 Store 层
 */

import { LabStatusCore } from '@/core/labStatusCore';
import type { LabStatusData } from '@/store/environmentStore';
import { useEnvironmentStore } from '@/store/environmentStore';
import { useEffect, useState } from 'react';

interface UseLabStatusOptions {
  onStatusUpdate?: (statuses: LabStatusData[]) => void;
  autoConnect?: boolean;
  autoQueryList?: boolean; // 是否自动查询所有实验室
  labUuid?: string; // 指定实验室 UUID（用于详情页）
  autoQueryDetail?: boolean; // 是否自动查询单个实验室详情
}

/**
 * 实验室状态 Hook
 *
 * @example
 * // 列表页使用 - 自动查询所有实验室
 * const { isConnected, labStatuses } = useLabStatus({
 *   autoQueryList: true,
 *   onStatusUpdate: (statuses) => {
 *     console.log('状态更新:', statuses);
 *   }
 * });
 *
 * @example
 * // 详情页使用 - 自动查询单个实验室
 * const { isConnected, getStatus } = useLabStatus({
 *   labUuid: 'xxx-xxx-xxx',
 *   autoQueryDetail: true,
 *   onStatusUpdate: (statuses) => {
 *     console.log('状态更新:', statuses);
 *   }
 * });
 *
 * @example
 * // 手动控制查询
 * const { queryList, queryDetail } = useLabStatus();
 * // 在某个时机手动调用 queryList() 或 queryDetail(uuid)
 */
export function useLabStatus(options: UseLabStatusOptions = {}) {
  const {
    onStatusUpdate,
    autoConnect = true,
    autoQueryList = false,
    labUuid,
    autoQueryDetail = false,
  } = options;

  // 从 store 读取连接状态
  const isConnected = useEnvironmentStore(
    (state) => state.isLabStatusConnected
  );

  // 从 store 读取所有状态
  const labStatuses = useEnvironmentStore((state) => state.labStatuses);

  // 本地状态：是否已初始化
  const [isInitialized, setIsInitialized] = useState(false);

  // 自动连接
  useEffect(() => {
    if (autoConnect && !isInitialized) {
      console.log('🔌 [useLabStatus] Auto-connecting...');
      LabStatusCore.connect();
      setIsInitialized(true);
    }
  }, [autoConnect, isInitialized]);

  // 自动查询列表（连接成功后）
  useEffect(() => {
    if (autoQueryList && isConnected) {
      console.log('🔍 [useLabStatus] Auto-querying lab list...');
      LabStatusCore.queryList().catch((error) => {
        console.error('❌ [useLabStatus] Auto-query list failed:', error);
      });
    }
  }, [autoQueryList, isConnected]);

  // 自动查询详情（连接成功后）
  useEffect(() => {
    if (autoQueryDetail && isConnected && labUuid) {
      console.log(`🔍 [useLabStatus] Auto-querying lab detail: ${labUuid}`);
      LabStatusCore.queryDetail(labUuid).catch((error) => {
        console.error('❌ [useLabStatus] Auto-query detail failed:', error);
      });
    }
  }, [autoQueryDetail, isConnected, labUuid]);

  // 订阅状态更新
  useEffect(() => {
    if (!onStatusUpdate) return;

    console.log('📡 [useLabStatus] Subscribing to status updates');
    const unsubscribe = LabStatusCore.subscribe(onStatusUpdate);

    return () => {
      console.log('📡 [useLabStatus] Unsubscribing from status updates');
      unsubscribe();
    };
  }, [onStatusUpdate]);

  // 查询所有实验室状态
  const queryList = async (): Promise<LabStatusData[]> => {
    return LabStatusCore.queryList();
  };

  // 查询单个实验室状态
  const queryDetail = async (labUuid: string): Promise<LabStatusData> => {
    return LabStatusCore.queryDetail(labUuid);
  };

  // 获取特定实验室的状态
  const getStatus = (labUuid: string): LabStatusData | undefined => {
    return useEnvironmentStore.getState().getLabStatus(labUuid);
  };

  // 手动连接
  const connect = () => {
    LabStatusCore.connect();
  };

  // 手动断开
  const disconnect = () => {
    LabStatusCore.disconnect();
  };

  return {
    // 状态
    isConnected,
    labStatuses,

    // 方法
    queryList,
    queryDetail,
    getStatus,
    connect,
    disconnect,
  };
}
