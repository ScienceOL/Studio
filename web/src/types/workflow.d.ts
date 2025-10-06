import { Dispatch } from '@reduxjs/toolkit';
import { Edge, NodeProps } from 'reactflow';

declare global {
  type HandleDataSourceProps = 'executor' | 'param' | 'handle';

  interface WorkflowNodeDataHandlesProps {
    key: string;
    type: 'source' | 'target';
    label?: string;
    hasConnected?: boolean;
    data_source?: HandleDataSourceProps;
    data_key?: string;
    rope?: string;
  }

  interface WorkflowNodeDataParamProps {
    id: string;
    key: string;
    type: 'DEFAULT' | 'FILE' | 'TEXTAREA' | 'SELECT';
    select_choices?: { name: string; value: string }[];
    source: string | object;
    executor?: string[]; // 记录这个 Param 运行的 Result 的 key
    title?: string;
    attachment?: string;
    schema?: {
      panel_type?: 'default';
      schema: object;
      uiSchema?: object;
    };
  }

  interface WorkflowNodeExecutorProps {
    id: string;
    key: string; // 记录 Result 的实际的 key
    source: string; // 记录 Result 的结果
    script?: string; // 运行用户输入的脚本
    mirage?: string; // Result 运行测试用 key
    type: 'Parameter' | 'Artifact';
    containerd?: boolean; // 是否在容器中运行
    image?: string; // 容器镜像
    resource?: object; // 资源配置
    readable?: boolean; // 是否在节点中展示
    title?: string; // 前端展示的标题
  }

  interface WorkflowNodeDataProps {
    header: string;
    // 不同节点中有相同的 key，这个 key 用来判断 handle 之间是否可以连接
    handles: WorkflowNodeDataHandlesProps[];
    params: WorkflowNodeDataParamProps[]; // param 属性决定 Node 上展示的与用户交互的表单，表单提交的行为回调到 Redux
    executors: WorkflowNodeExecutorProps[];
    footer?: string;
  }

  interface WorkflowNodeProps extends Node<WorkflowNodeDataProps, NodeTypes> {
    id: string;
    type?: NodeTypes;
    template: string; // 这个 template 无法传入到 ReactFlow 画布中的 Node
    version: string;
    data: WorkflowNodeDataProps;
    dragHandle?: string | '.drag-handle';
    position: { x: number; y: number };
    positionAbsolute?: { x: number; y: number };
    status?: 'draft' | 'skipped' | 'pending' | 'running' | 'success' | 'failed';
    minimized?: boolean;
    disabled?: boolean;
  }

  type addNodeProps = Omit<WorkflowNodeProps, 'id' | 'position'> & {
    id?: string;
  };

  interface WorkflowBasicProps {
    id: string;
    uuid: string;
    name: string;
    created_at: Date | string;
    updated_at: Date | string;
    creator: { username: string; avatar: string };
    public?: boolean;
    as_template?: boolean;
  }

  interface WorkflowProps extends WorkflowBasicProps {
    cover?: string;
    description?: string;

    status:
      | 'draft'
      | 'pending'
      | 'running'
      | 'finished'
      | 'canceled'
      | 'failed';
    nodes: WorkflowNodeProps[];
    edges: Edge[];
  }

  interface WorkflowStateProps {
    sideMenuVisible: boolean;
    contextMenuVisible: boolean;
    contextMenuX: number;
    contextMenuY: number;
    activeMenuItems: string[];
    sliderOverlayVisible: boolean;
    sliderOverlay?: {
      nodeId: string;
      paramId?: string;
      executorId?: string;
    };
    nodes: WorkflowNodeProps[];
    edges: Edge[];
    workflow: WorkflowProps;
    workflowList: WorkflowBasicProps[];
    tasks: {
      id: string;
      snapshot: any;
      status: string;
      created_at: string;
      finished_at: string;
      workflow: string;
      uuid: string;
    }[];
    consoleInfo: WorkflowWebSocketResponseMsgProps[];
    status: 'idle' | 'loading' | 'error';
  }
  interface ContextMenuItemProps {
    action: string;
    label: string;
    icon: React.FC<React.SVGProps<SVGSVGElement>>;
    arrow?: boolean;
    onClick?: () => (dispatch: Dispatch) => void;
    subContextMenuItems?: ContextMenuItemProps[];
  }

  interface BasicNodeProps extends NodeProps {
    data: WorkflowNodeDataProps;
    children?: React.ReactNode;
    // minimized?: boolean;
  }

  interface ExecutedNodeMessageProps {
    id: string;
    header: string;
    status: 'success' | 'failed' | 'running' | 'draft' | 'skipped' | 'pending';
    executor?: [];
    messages?: { type: 'info' | 'warning' | 'error'; message: string }[];
  }

  interface ExecuteStatusProps {
    uuid: string;
    header: string;
    status: 'success' | 'failed' | 'running' | 'draft' | 'skipped' | 'pending';
    executor: { [key: string]: string | object }[];
    messages: { type: 'info' | 'warning' | 'error'; message: string }[];
  }
}
