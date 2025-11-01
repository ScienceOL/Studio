import { config } from '@/configs';
import apiClient from '@/service/http/client';

// 工作流相关服务
export const workflowService = {
  // ========== 工作流任务相关 ==========

  // 获取工作流任务列表
  async getTaskList(
    uuid: string,
    params?: {
      page?: number;
      page_size?: number;
      status?: string;
      [key: string]: unknown;
    }
  ) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/task/${uuid}`,
      {
        params,
      }
    );
    return res.data;
  },

  // 下载工作流任务
  async downloadTask(uuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/task/download/${uuid}`
    );
    return res.data;
  },

  // ========== 工作流模板相关 ==========
  // 获取工作流模板详情
  async getWorkflowDetail(uuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/template/detail/${uuid}`
    );
    return res.data;
  },

  // Fork 工作流模板
  async forkTemplate(data: {
    template_uuid: string;
    name?: string;
    description?: string;
    [key: string]: unknown;
  }) {
    const res = await apiClient.put(
      `${config.apiBaseUrl}/api/v1/lab/workflow/template/fork`,
      data
    );
    return res.data;
  },

  // 获取工作流模板标签
  async getWorkflowTemplateTags() {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/template/tags`
    );
    return res.data;
  },

  // 按实验室获取工作流模板标签
  async getWorkflowTemplateTagsByLab(labUuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/template/tags/${labUuid}`
    );
    return res.data;
  },

  // 获取工作流模板列表
  async getWorkflowTemplateList(params?: {
    page?: number;
    page_size?: number;
    tags?: string[];
    keyword?: string;
    [key: string]: unknown;
  }) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/template/list`,
      {
        params,
      }
    );
    return res.data;
  },

  // ========== 节点模板相关 ==========

  // 获取节点模板标签
  async getNodeTemplateTags(labUuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/node/template/tags/${labUuid}`
    );
    return res.data;
  },

  // 获取节点模板列表
  async getNodeTemplateList(params?: {
    lab_uuid?: string;
    page?: number;
    page_size?: number;
    tags?: string[];
    [key: string]: unknown;
  }) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/node/template/list`,
      {
        params,
      }
    );
    return res.data;
  },

  // 获取节点模板详情
  async getNodeTemplateDetail(uuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/node/template/detail/${uuid}`
    );
    return res.data;
  },

  // ========== 我的工作流相关 ==========

  // 创建工作流
  async createWorkflow(data: {
    lab_uuid: string;
    name: string;
    description?: string;
    nodes?: unknown[];
    edges?: unknown[];
    [key: string]: unknown;
  }) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/lab/workflow/owner`,
      data
    );
    return res.data;
  },

  // 更新工作流
  async updateWorkflow(data: {
    uuid: string;
    name?: string;
    description?: string;
    nodes?: unknown[];
    edges?: unknown[];
    [key: string]: unknown;
  }) {
    const res = await apiClient.patch(
      `${config.apiBaseUrl}/api/v1/lab/workflow/owner`,
      data
    );
    return res.data;
  },

  // 删除工作流
  async deleteWorkflow(uuid: string) {
    const res = await apiClient.delete(
      `${config.apiBaseUrl}/api/v1/lab/workflow/owner/${uuid}`
    );
    return res.data;
  },

  // 获取工作流列表
  async getWorkflowList(params?: {
    lab_uuid?: string;
    page?: number;
    page_size?: number;
    [key: string]: unknown;
  }) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/owner/list`,
      {
        params,
      }
    );
    return res.data;
  },

  // 导出工作流
  async exportWorkflow(params: { uuid: string; [key: string]: unknown }) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/workflow/owner/export`,
      {
        params,
      }
    );
    return res.data;
  },

  // 导入工作流
  async importWorkflow(data: {
    lab_uuid: string;
    workflow_data: unknown;
    [key: string]: unknown;
  }) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/lab/workflow/owner/import`,
      data
    );
    return res.data;
  },

  // 复制工作流
  async duplicateWorkflow(data: {
    uuid: string;
    name?: string;
    [key: string]: unknown;
  }) {
    const res = await apiClient.put(
      `${config.apiBaseUrl}/api/v1/lab/workflow/owner/duplicate`,
      data
    );
    return res.data;
  },

  // ========== 运行工作流 ==========

  // 运行工作流
  async runWorkflow(data: {
    workflow_uuid: string;
    lab_uuid: string;
    parameters?: Record<string, unknown>;
    [key: string]: unknown;
  }) {
    const res = await apiClient.put(
      `${config.apiBaseUrl}/api/v1/lab/run/workflow`,
      data
    );
    return res.data;
  },

  // WebSocket 连接（需要在组件中单独处理）
  getWebSocketUrl(uuid: string): string {
    const wsProtocol = config.apiBaseUrl.startsWith('https') ? 'wss' : 'ws';
    const baseUrl = config.apiBaseUrl.replace(/^https?:/, '');
    return `${wsProtocol}:${baseUrl}/api/v1/lab/workflow/ws/workflow/${uuid}`;
  },
};
