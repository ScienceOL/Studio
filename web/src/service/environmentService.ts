import { config } from '@/configs';
import apiClient from '@/service/http/client';
import type {
  CreateInviteRequest,
  CreateLabRequest,
  LabResource,
  UpdateLabRequest,
} from '@/types/environment';
import {
  normalizePaginationParams,
  type PaginationParams,
} from '@/utils/pagination';

// 环境（实验室）相关服务
export const environmentService = {
  // 创建实验室
  async createLab(data: CreateLabRequest) {
    const res = await apiClient.post(`${config.apiBaseUrl}/api/v1/lab`, data);
    return res.data;
  },

  // 更新实验室
  async updateLab(data: UpdateLabRequest) {
    const res = await apiClient.patch(`${config.apiBaseUrl}/api/v1/lab`, data);
    return res.data;
  },

  // 获取实验室列表
  async getLabList(params?: PaginationParams) {
    // 统一规范化分页参数
    const normalizedParams = normalizePaginationParams(params);
    const res = await apiClient.get(`${config.apiBaseUrl}/api/v1/lab/list`, {
      params: normalizedParams,
    });
    return res.data;
  },

  // 获取实验室信息
  async getLabInfo(uuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/info/${uuid}`
    );
    return res.data;
  },

  // 创建实验室资源（从 edge 侧）
  async createLabResource(data: Partial<LabResource> & { lab_uuid: string }) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/lab/resource`,
      data
    );
    return res.data;
  },

  // 获取实验室成员
  async getLabMembers(labUuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/member/${labUuid}`
    );
    return res.data;
  },

  // 删除实验室成员
  async deleteLabMember(labUuid: string, memberUuid: string) {
    const res = await apiClient.delete(
      `${config.apiBaseUrl}/api/v1/lab/member/${labUuid}/${memberUuid}`
    );
    return res.data;
  },

  // 创建邀请链接
  async createInvite(labUuid: string, data?: CreateInviteRequest) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/lab/invite/${labUuid}`,
      data
    );
    return res.data;
  },

  // 接受邀请
  async acceptInvite(inviteUuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/invite/${inviteUuid}`
    );
    return res.data;
  },

  // 获取用户信息
  async getUserInfo() {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/user/info`
    );
    return res.data;
  },
};
