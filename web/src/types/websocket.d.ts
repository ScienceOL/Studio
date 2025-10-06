// 声明全局类型（无需导出）
declare global {
  // 基础动作类型
  type BasicActionType =
    | 'connect'
    | 'disconnect'
    | 'ping'
    | 'pong'
    | 'status'
    | 'error';

  type WorkflowActionType =
    | BasicActionType
    | 'start'
    | 'stop'
    | 'pause'
    | 'resume'
    | 'validate'
    | 'debug'
    | 'workflowUpdate';

  // 泛型基础消息接口
  interface BaseWebSocketMsgProps<TData = any, TAction = BasicActionType> {
    id: string;
    time: string;
    action: TAction;
    version?: string;
    data?: TData;
  }

  interface WebSocketPostMsgProps<TData = any, TAction = BasicActionType>
    extends BaseWebSocketMsgProps<TData, TAction> {
    // 请求专用字段
    // 请求类型通常不需要 code/message
  }

  interface WebSocketResponseMsgProps<TData = any, TAction = BasicActionType>
    extends BaseWebSocketMsgProps<TData, TAction> {
    // 响应专用字段
    code: number; // 响应必须包含状态码, 0 for success, 4 digits for error, e.g. 1000
    message?: string; // 错误信息可选
    type?: string; // 后端调用的返回函数
  }

  interface WorkflowMsgDataProps {
    workflow?: any;
    params?: any;
    type?: 'info' | 'warning' | 'error'; // console 的消息类型，用来渲染颜色
    stack_trace?: string; // console 中点击后的 stack trace
    node_uuid?: string; // 返回的结果中用来定位到 node
    status?: 'draft' | 'skipped' | 'pending' | 'running' | 'success' | 'failed';
    executors?: WorkflowNodeExecutorProps[];
  }

  type WorkflowWebSocketResponseMsgProps = WebSocketResponseMsgProps<
    WorkflowMsgDataProps,
    WorkflowActionType
  >;
}

// 确保文件作为模块（避免全局污染）
export {};
