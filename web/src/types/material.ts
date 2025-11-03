// 物料相关类型定义

export interface Material {
  uuid: string;
  lab_uuid?: string;
  id: string; // 唯一标识，主键（对应后端的 id 字段）
  name: string; // 设备名称（device_id，用于 action 执行）
  type: string; // 类型：device/resource
  class?: string; // 资源类名，对应 ResourceTemplate.name（例如 gripper.mock）
  status?: string; // 状态：active/inactive 等
  parent_uuid?: string; // 父节点 UUID
  parent?: string | null; // 父节点 ID（对应后端的 parent 字段）
  children?: string[]; // 子节点 ID 列表
  position?: {
    x: number;
    y: number;
    z: number;
  };
  config?: Record<string, unknown>; // 配置信息
  data?: Record<string, unknown>; // 数据信息
  properties?: Record<string, unknown>;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
}

export interface CreateMaterialRequest {
  lab_uuid: string;
  name: string;
  type?: string;
  properties?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface QueryMaterialParams {
  lab_uuid?: string;
  uuid?: string;
  type?: string;
  [key: string]: unknown;
}

export interface BatchUpdateMaterialRequest {
  materials: Array<{
    uuid: string;
    name?: string;
    status?: string;
    properties?: Record<string, unknown>;
    [key: string]: unknown;
  }>;
}

export interface SaveMaterialRequest {
  lab_uuid: string;
  materials?: Material[];
  edges?: MaterialEdge[];
  [key: string]: unknown;
}

export interface MaterialResource {
  uuid: string;
  lab_uuid: string;
  device_uuid: string;
  device_name: string;
  device_type: string;
  status: string;
  capabilities?: string[];
  [key: string]: unknown;
}

export interface DeviceAction {
  uuid: string;
  device_uuid: string;
  name: string;
  description?: string;
  action_type: string;
  parameters?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface MaterialEdge {
  uuid?: string;
  lab_uuid: string;
  source: string;
  target: string;
  source_handle?: string;
  target_handle?: string;
  properties?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface CreateMaterialEdgeRequest {
  lab_uuid: string;
  source: string;
  target: string;
  source_handle?: string;
  target_handle?: string;
  properties?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface MaterialTemplate {
  uuid: string;
  name: string;
  description?: string;
  type: string;
  content?: unknown;
  [key: string]: unknown;
}

export interface MaterialDAG {
  lab_uuid: string;
  nodes: Material[];
  edges: MaterialEdge[];
  version?: string;
  created_at: string;
  updated_at: string;
}

// Edge 侧物料相关类型
export interface EdgeCreateMaterialRequest {
  lab_uuid: string;
  device_uuid: string;
  materials: Array<{
    name: string;
    type: string;
    properties?: Record<string, unknown>;
    [key: string]: unknown;
  }>;
  [key: string]: unknown;
}

export interface EdgeUpsertMaterialRequest {
  lab_uuid: string;
  materials: Array<{
    uuid?: string;
    name?: string;
    type?: string;
    properties?: Record<string, unknown>;
    [key: string]: unknown;
  }>;
}

export interface QueryMaterialByUUIDRequest {
  uuids: string[];
  lab_uuid?: string;
}

// Resource 资源信息类型（对应后端 ResourceInfo）
export interface ResourceInfo {
  uuid: string;
  name: string;
  parent_uuid?: string;
  [key: string]: unknown;
}

// Resource 资源模板类型（对应后端 ResourceNodeTemplate）
export interface ResourceTemplate {
  uuid: string;
  name: string;
  resource_type: string;
  language?: string;
  icon?: string;
  description?: string | null;
  tags?: string[];
  version?: string;
  module?: string;
  model?: Record<string, unknown>;
  data_schema?: Record<string, unknown>;
  config_schema?: Record<string, unknown>;
  actions?: DeviceActionInfo[];
  material_count?: number;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
}

// 设备动作信息
export interface DeviceActionInfo {
  name: string;
  type: string;
  schema?: Record<string, unknown>;
  goal?: Record<string, unknown>;
  goal_default?: Record<string, unknown>;
  feedback?: Record<string, unknown>;
  result?: Record<string, unknown>;
}
