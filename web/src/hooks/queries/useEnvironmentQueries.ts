/**
 * 🎣 Query Hook Layer - Environment 数据查询
 *
 * 职责：
 * 1. 封装 Service 层的 HTTP 请求
 * 2. 提供 TanStack Query 缓存策略
 * 3. 管理服务器状态（列表、详情等）
 *
 * 注意：Query 层不写 Store，服务器状态由 React Query 管理
 */

import { environmentService } from '@/service';
import { materialService } from '@/service/materialService';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

// ============= Query Keys =============
export const environmentKeys = {
  all: ['environment'] as const,
  lists: () => [...environmentKeys.all, 'list'] as const,
  list: (params?: Record<string, unknown>) =>
    [...environmentKeys.lists(), params] as const,
  details: () => [...environmentKeys.all, 'detail'] as const,
  detail: (uuid: string) => [...environmentKeys.details(), uuid] as const,
  members: (labUuid: string) =>
    [...environmentKeys.all, 'members', labUuid] as const,
  userInfo: () => [...environmentKeys.all, 'userInfo'] as const,
};

// ============= Query Hooks =============

/**
 * 获取实验室列表
 * 注意：分页参数会在 Service 层自动规范化，传 undefined 会使用默认值
 */
export function useLabList(params?: { page?: number; page_size?: number }) {
  return useQuery({
    queryKey: environmentKeys.list(params),
    queryFn: () => environmentService.getLabList(params),
    staleTime: 30000, // 30秒内认为数据是新鲜的
    gcTime: 5 * 60 * 1000, // 5分钟后垃圾回收
  });
}

/**
 * 获取实验室详情
 */
export function useLabDetail(uuid: string, enabled = true) {
  return useQuery({
    queryKey: environmentKeys.detail(uuid),
    queryFn: () => environmentService.getLabInfo(uuid),
    enabled: !!uuid && enabled,
    staleTime: 60000, // 1分钟
    select: (data) => data?.data,
  });
}

/**
 * 获取实验室成员列表
 */
export function useLabMembers(labUuid: string, enabled = true) {
  return useQuery({
    queryKey: environmentKeys.members(labUuid),
    queryFn: () => environmentService.getLabMembers(labUuid),
    enabled: !!labUuid && enabled,
    staleTime: 30000,
    select: (data) => data?.data || [],
  });
}

/**
 * 获取当前用户信息
 */
export function useUserInfo() {
  return useQuery({
    queryKey: environmentKeys.userInfo(),
    queryFn: () => environmentService.getUserInfo(),
    staleTime: 5 * 60 * 1000, // 5分钟
  });
}

/**
 * 获取实验室的资源模板列表
 */
export function useResourceTemplates(labUuid: string, enabled = true) {
  return useQuery({
    queryKey: [...environmentKeys.detail(labUuid), 'resource-templates'],
    queryFn: () => materialService.getResourceTemplates({ lab_uuid: labUuid }),
    enabled: !!labUuid && enabled,
    staleTime: 60000, // 1分钟
    select: (data) => data?.data?.templates || [],
  });
}

/**
 * 获取实验室的物料列表
 */
export function useMaterials(labUuid: string, enabled = true) {
  return useQuery({
    queryKey: [...environmentKeys.detail(labUuid), 'materials'],
    queryFn: () => materialService.downloadMaterial(labUuid),
    enabled: !!labUuid && enabled,
    staleTime: 30000, // 30秒
    select: (data) => data?.data?.nodes || [],
  });
}

// ============= Mutation Hooks =============

/**
 * 创建实验室
 */
export function useCreateLab() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; description?: string }) =>
      environmentService.createLab(data),
    onSuccess: () => {
      // 创建成功后，使列表缓存失效
      queryClient.invalidateQueries({ queryKey: environmentKeys.lists() });
    },
  });
}

/**
 * 更新实验室
 */
export function useUpdateLab() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { uuid: string; name?: string; description?: string }) =>
      environmentService.updateLab(data),
    onSuccess: (_, variables) => {
      // 更新成功后，使相关缓存失效
      queryClient.invalidateQueries({
        queryKey: environmentKeys.detail(variables.uuid),
      });
      queryClient.invalidateQueries({ queryKey: environmentKeys.lists() });
    },
  });
}

/**
 * 删除实验室成员
 */
export function useDeleteLabMember() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      labUuid,
      memberUuid,
    }: {
      labUuid: string;
      memberUuid: string;
    }) => environmentService.deleteLabMember(labUuid, memberUuid),
    onSuccess: (_, variables) => {
      // 删除成功后，使成员列表缓存失效
      queryClient.invalidateQueries({
        queryKey: environmentKeys.members(variables.labUuid),
      });
    },
  });
}

/**
 * 创建邀请链接
 */
export function useCreateInvite() {
  return useMutation({
    mutationFn: ({
      labUuid,
      data,
    }: {
      labUuid: string;
      data?: { expires_at?: string; role?: string };
    }) => environmentService.createInvite(labUuid, data),
  });
}

/**
 * 接受邀请
 */
export function useAcceptInvite() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (inviteUuid: string) =>
      environmentService.acceptInvite(inviteUuid),
    onSuccess: () => {
      // 接受邀请后，刷新实验室列表
      queryClient.invalidateQueries({ queryKey: environmentKeys.lists() });
    },
  });
}
