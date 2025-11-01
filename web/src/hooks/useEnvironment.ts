/**
 * 🎣 Hook Layer - Environment 能力暴露
 *
 * 职责：
 * 1. 组合 Query Hook 和 Core 方法
 * 2. 订阅 Store 状态
 * 3. 提供统一的 API 给组件
 *
 * 数据流：
 * - 简单查询：Component → useEnvironment → Query Hook → Service
 * - 复杂流程：Component → useEnvironment → Core → Service + Store
 */

import { EnvironmentCore } from '@/core/environmentCore';
import {
  useAcceptInvite,
  useCreateInvite,
  useCreateLab,
  useDeleteLabMember,
  useLabDetail,
  useLabList,
  useLabMembers,
  useUpdateLab,
} from '@/hooks/queries/useEnvironmentQueries';
import { useEnvironmentStore } from '@/store/environmentStore';
import { useCallback } from 'react';

/**
 * Environment 主 Hook
 * 提供所有环境相关的能力
 */
export function useEnvironment() {
  // ========== 订阅 Store 状态 ==========
  const selectedLabUuid = useEnvironmentStore((state) => state.selectedLabUuid);
  const currentLabUuid = useEnvironmentStore((state) => state.currentLabUuid);
  const isCreateDialogOpen = useEnvironmentStore(
    (state) => state.isCreateDialogOpen
  );
  const isCredentialsDialogOpen = useEnvironmentStore(
    (state) => state.isCredentialsDialogOpen
  );
  const viewMode = useEnvironmentStore((state) => state.viewMode);

  // ========== Store Actions ==========
  const setSelectedLabUuid = useEnvironmentStore(
    (state) => state.setSelectedLabUuid
  );
  const setCreateDialogOpen = useEnvironmentStore(
    (state) => state.setCreateDialogOpen
  );
  const setCredentialsDialogOpen = useEnvironmentStore(
    (state) => state.setCredentialsDialogOpen
  );
  const setViewMode = useEnvironmentStore((state) => state.setViewMode);

  // ========== Query Hooks（简单数据获取） ==========
  const labListQuery = useLabList();
  const currentLabQuery = useLabDetail(currentLabUuid || '', !!currentLabUuid);
  const membersQuery = useLabMembers(currentLabUuid || '', !!currentLabUuid);

  // ========== Mutations ==========
  const createLabMutation = useCreateLab();
  const updateLabMutation = useUpdateLab();
  const deleteMemberMutation = useDeleteLabMember();
  const createInviteMutation = useCreateInvite();
  const acceptInviteMutation = useAcceptInvite();

  // ========== Core 方法（复杂流程编排） ==========

  // 进入实验室
  const enterLab = useCallback(async (labUuid: string) => {
    await EnvironmentCore.enterLab(labUuid);
  }, []);

  // 退出实验室
  const exitLab = useCallback(() => {
    EnvironmentCore.exitLab();
  }, []);

  // 创建并进入实验室
  const createAndEnterLab = useCallback(
    async (data: { name: string; description?: string }) => {
      const labUuid = await EnvironmentCore.createAndEnterLab(data);
      // 触发列表刷新
      await labListQuery.refetch();
      return labUuid;
    },
    [labListQuery]
  );

  // 生成凭证
  const getLabCredentials = useCallback(async (labUuid: string) => {
    return await EnvironmentCore.getLabCredentials(labUuid);
  }, []);

  // 复制到剪贴板
  const copyToClipboard = useCallback(async (text: string, label?: string) => {
    await EnvironmentCore.copyToClipboard(text, label);
  }, []);

  // 移除成员
  const removeMember = useCallback(
    async (labUuid: string, memberUuid: string, memberName?: string) => {
      await EnvironmentCore.removeMember(labUuid, memberUuid, memberName);
    },
    []
  );

  // ========== 派生状态 ==========
  // 注意：后端返回的是 { code: 0, data: { data: [...], page: 1, page_size: 20 } }
  const labs = labListQuery.data?.data?.data || [];
  const currentLab = currentLabQuery.data?.data;
  const members = membersQuery.data?.data?.data || [];

  const isLoadingLabs = labListQuery.isLoading;
  const isLoadingCurrentLab = currentLabQuery.isLoading;
  const isLoadingMembers = membersQuery.isLoading;

  return {
    // ===== 状态 =====
    selectedLabUuid,
    currentLabUuid,
    isCreateDialogOpen,
    isCredentialsDialogOpen,
    viewMode,

    // ===== 数据 =====
    labs,
    currentLab,
    members,

    // ===== 加载状态 =====
    isLoadingLabs,
    isLoadingCurrentLab,
    isLoadingMembers,

    // ===== Query 对象（用于访问更多元信息） =====
    labListQuery,
    currentLabQuery,
    membersQuery,

    // ===== UI Actions =====
    setSelectedLabUuid,
    setCreateDialogOpen,
    setCredentialsDialogOpen,
    setViewMode,

    // ===== 业务 Actions（简单） =====
    createLab: createLabMutation.mutateAsync,
    updateLab: updateLabMutation.mutateAsync,
    deleteMember: deleteMemberMutation.mutateAsync,
    createInvite: createInviteMutation.mutateAsync,
    acceptInvite: acceptInviteMutation.mutateAsync,

    // ===== 业务 Actions（复杂流程） =====
    enterLab,
    exitLab,
    createAndEnterLab,
    getLabCredentials,
    copyToClipboard,
    removeMember,

    // ===== Mutation 状态 =====
    isCreating: createLabMutation.isPending,
    isUpdating: updateLabMutation.isPending,
    isDeletingMember: deleteMemberMutation.isPending,
  };
}

/**
 * 获取特定实验室详情的 Hook
 */
export function useLabInfo(labUuid: string, enabled = true) {
  const query = useLabDetail(labUuid, enabled);

  return {
    lab: query.data?.data,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}

/**
 * 获取实验室成员的 Hook
 */
export function useLabMembersList(labUuid: string, enabled = true) {
  const query = useLabMembers(labUuid, enabled);

  return {
    members: query.data?.data?.members || [],
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}
