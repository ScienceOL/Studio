import { config } from '@/configs';
import apiClient from '@/service/http/client';
import type {
  BatchUpdateMaterialRequest,
  CreateMaterialEdgeRequest,
  CreateMaterialRequest,
  EdgeCreateMaterialRequest,
  EdgeUpsertMaterialRequest,
  QueryMaterialByUUIDRequest,
  QueryMaterialParams,
  SaveMaterialRequest,
} from '@/types/material';

// 物料相关服务
export const materialService = {
  // 创建物料
  async createMaterial(data: CreateMaterialRequest) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/lab/material`,
      data
    );
    return res.data;
  },

  // 查询物料（edge 侧查询物料资源）
  async queryMaterial(params: QueryMaterialParams) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/material`,
      {
        params,
      }
    );
    return res.data;
  },

  // 批量更新物料数据（edge 批量更新）
  async batchUpdateMaterial(data: BatchUpdateMaterialRequest) {
    const res = await apiClient.put(
      `${config.apiBaseUrl}/api/v1/lab/material`,
      data
    );
    return res.data;
  },

  // 保存物料
  async saveMaterial(data: SaveMaterialRequest) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/lab/material/save`,
      data
    );
    return res.data;
  },

  // 获取实验室所有设备列表（简化版）
  async getResourceList(params: { lab_uuid: string }) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/material/resource`,
      {
        params,
      }
    );
    return res.data;
  },

  // 获取资源模板详细信息（包含 actions、schema 等）
  async getResourceTemplates(params: { lab_uuid: string }) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/material/resource/templates`,
      {
        params,
      }
    );
    return res.data;
  },

  // 获取实验室所有动作
  async getActions(params?: { lab_uuid?: string; device_uuid?: string }) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/material/device/actions`,
      {
        params,
      }
    );
    return res.data;
  },

  // 创建物料连线
  async createMaterialEdge(data: CreateMaterialEdgeRequest) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/lab/material/edge`,
      data
    );
    return res.data;
  },

  // 下载物料 DAG
  async downloadMaterial(labUuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/material/download/${labUuid}`
    );
    return res.data;
  },

  // 获取物料模板
  async getMaterialTemplate(templateUuid: string) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/lab/material/template/${templateUuid}`
    );
    return res.data;
  },

  // WebSocket 连接（需要在组件中单独处理）
  getWebSocketUrl(labUuid: string): string {
    const wsProtocol = config.apiBaseUrl.startsWith('https') ? 'wss' : 'ws';
    const baseUrl = config.apiBaseUrl.replace(/^https?:/, '');
    return `${wsProtocol}:${baseUrl}/api/v1/ws/material/${labUuid}`;
  },
};

// Edge 侧物料相关服务（用于边缘设备上报）
export const edgeMaterialService = {
  // Edge 创建物料
  async createMaterial(data: EdgeCreateMaterialRequest) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/edge/material`,
      data
    );
    return res.data;
  },

  // Edge 更新或创建物料
  async upsertMaterial(data: EdgeUpsertMaterialRequest) {
    const res = await apiClient.put(
      `${config.apiBaseUrl}/api/v1/edge/material`,
      data
    );
    return res.data;
  },

  // Edge 创建连线
  async createEdge(data: CreateMaterialEdgeRequest) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/edge/material/edge`,
      data
    );
    return res.data;
  },

  // Edge 根据 UUID 查询物料
  async queryMaterialByUUID(data: QueryMaterialByUUIDRequest) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/v1/edge/material/query`,
      data
    );
    return res.data;
  },

  // Edge 下载物料
  async downloadMaterial(params: { lab_uuid: string }) {
    const res = await apiClient.get(
      `${config.apiBaseUrl}/api/v1/edge/material/download`,
      {
        params,
      }
    );
    return res.data;
  },
};
