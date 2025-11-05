/**
 * 🏪 Store Layer - Environment 客户端状态
 *
 * 职责：
 * 1. 管理客户端 UI 状态（当前选中的实验室、展开/折叠等）
 * 2. 管理会话状态（不需要持久化的临时状态）
 * 3. 管理实验室在线状态（实时更新）
 *
 * 注意：
 * - 不存储服务器数据（列表、详情等），那些由 React Query 管理
 * - 只存储 UI 交互状态和会话状态
 * - 实验室状态由 WebSocket 实时更新
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface LabStatusData {
  lab_uuid: string;
  is_online: boolean;
  last_connected_at?: string;
}

export interface EnvironmentState {
  // 当前选中的实验室 UUID
  selectedLabUuid: string | null;

  // 当前查看的实验室 UUID（进入详情页）
  currentLabUuid: string | null;

  // 是否显示创建对话框
  isCreateDialogOpen: boolean;

  // 是否显示 AK/SK 对话框
  isCredentialsDialogOpen: boolean;

  // 列表视图模式（grid/list）
  viewMode: 'grid' | 'list';

  // 实验室状态映射表 (lab_uuid -> status)
  labStatuses: Map<string, LabStatusData>;

  // WebSocket 连接状态
  isLabStatusConnected: boolean;
}

export interface EnvironmentActions {
  // 设置选中的实验室
  setSelectedLabUuid: (uuid: string | null) => void;

  // 设置当前查看的实验室
  setCurrentLabUuid: (uuid: string | null) => void;

  // 切换创建对话框
  setCreateDialogOpen: (open: boolean) => void;

  // 切换凭证对话框
  setCredentialsDialogOpen: (open: boolean) => void;

  // 切换视图模式
  setViewMode: (mode: 'grid' | 'list') => void;

  // 更新实验室状态
  updateLabStatus: (labUuid: string, status: LabStatusData) => void;

  // 获取实验室状态
  getLabStatus: (labUuid: string) => LabStatusData | undefined;

  // 设置 WebSocket 连接状态
  setLabStatusConnected: (connected: boolean) => void;

  // 重置状态
  reset: () => void;
}

const initialState: EnvironmentState = {
  selectedLabUuid: null,
  currentLabUuid: null,
  isCreateDialogOpen: false,
  isCredentialsDialogOpen: false,
  viewMode: 'grid',
  labStatuses: new Map(),
  isLabStatusConnected: false,
};

export const useEnvironmentStore = create<
  EnvironmentState & EnvironmentActions
>()(
  persist(
    (set, get) => ({
      ...initialState,

      setSelectedLabUuid: (uuid) => set({ selectedLabUuid: uuid }),

      setCurrentLabUuid: (uuid) => set({ currentLabUuid: uuid }),

      setCreateDialogOpen: (open) => set({ isCreateDialogOpen: open }),

      setCredentialsDialogOpen: (open) =>
        set({ isCredentialsDialogOpen: open }),

      setViewMode: (mode) => set({ viewMode: mode }),

      updateLabStatus: (labUuid, status) =>
        set((state) => {
          const newStatuses = new Map(state.labStatuses);
          newStatuses.set(labUuid, status);
          return { labStatuses: newStatuses };
        }),

      getLabStatus: (labUuid) => {
        return get().labStatuses.get(labUuid);
      },

      setLabStatusConnected: (connected) =>
        set({ isLabStatusConnected: connected }),

      reset: () => set(initialState),
    }),
    {
      name: 'environment-storage',
      // 只持久化视图模式，其他状态是会话级别的
      partialize: (state) => ({
        viewMode: state.viewMode,
      }),
    }
  )
);
