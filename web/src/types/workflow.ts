// 工作流相关类型定义

export interface Workflow {
  uuid: string;
  lab_uuid: string;
  name: string;
  description?: string;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  status?: string;
  owner_uuid: string;
  created_at: string;
  updated_at: string;
  [key: string]: unknown;
}

export interface WorkflowNode {
  id: string;
  type: string;
  position: {
    x: number;
    y: number;
  };
  data: {
    label: string;
    template_uuid?: string;
    parameters?: Record<string, unknown>;
    [key: string]: unknown;
  };
  [key: string]: unknown;
}

export interface WorkflowEdge {
  id: string;
  source: string;
  target: string;
  sourceHandle?: string;
  targetHandle?: string;
  [key: string]: unknown;
}

export interface CreateWorkflowRequest {
  lab_uuid: string;
  name: string;
  description?: string;
  nodes?: WorkflowNode[];
  edges?: WorkflowEdge[];
  [key: string]: unknown;
}

export interface UpdateWorkflowRequest {
  uuid: string;
  name?: string;
  description?: string;
  nodes?: WorkflowNode[];
  edges?: WorkflowEdge[];
  [key: string]: unknown;
}

export interface WorkflowTask {
  uuid: string;
  workflow_uuid: string;
  workflow_name?: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  progress?: number;
  result?: unknown;
  error?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  [key: string]: unknown;
}

export interface WorkflowTemplate {
  uuid: string;
  name: string;
  description?: string;
  tags?: string[];
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  is_public: boolean;
  author_uuid?: string;
  author_name?: string;
  fork_count?: number;
  use_count?: number;
  created_at: string;
  updated_at: string;
  [key: string]: unknown;
}

export interface ForkTemplateRequest {
  template_uuid: string;
  name?: string;
  description?: string;
  [key: string]: unknown;
}

export interface NodeTemplate {
  uuid: string;
  name: string;
  description?: string;
  tags?: string[];
  type: string;
  icon?: string;
  input_schema?: Record<string, unknown>;
  output_schema?: Record<string, unknown>;
  parameters?: Record<string, unknown>;
  is_public: boolean;
  created_at: string;
  [key: string]: unknown;
}

export interface RunWorkflowRequest {
  workflow_uuid: string;
  lab_uuid: string;
  parameters?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface DuplicateWorkflowRequest {
  uuid: string;
  name?: string;
  [key: string]: unknown;
}

export interface ImportWorkflowRequest {
  lab_uuid: string;
  workflow_data: unknown;
  [key: string]: unknown;
}

export interface ExportWorkflowParams {
  uuid: string;
  format?: 'json' | 'yaml';
  [key: string]: unknown;
}
