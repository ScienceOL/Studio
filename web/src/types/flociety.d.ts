// Don't import this file directly, it is only for type checking

interface NodeTemplateProps {
  // id?: string; // 由前端添加的 uuid，用于标识节点
  name: string;
  description: string;
  version: string;
  created_at: string;
  updated_at: string;
  public: boolean;
  type: NodeTypes;
  data: NodeTemplateDataProps;
  creator: {
    username: string;
    avatar: string;
  };
}

// 节点数据
interface NodeTemplateDataProps {
  header: string;
  handles: WorkflowNodeDataHandlesProps[];
  params: Omit<WorkflowNodeDataParamProps, 'id'>[]; // param 属性决定 Node 上展示的与用户交互的表单，表单提交的行为回调到 Redux
  executors: Omit<WorkflowNodeExecutorProps, 'id'>[];
  footer?: string;
}

// 工作流帖子接口
interface WorkflowPostProps {
  title: string;
  cover: string;
  description: string;
  workflow: {
    id: number;
    uuid: string;
    name: string;
    created_at: string;
    creator: {
      username: string;
      avatar: string;
    };
    public?: boolean;
    as_template?: boolean;
  };
}
