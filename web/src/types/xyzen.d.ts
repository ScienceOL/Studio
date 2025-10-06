// 消息类型
export interface Message {
  id: string;
  sender: string;
  content: string;
  timestamp: string;
  avatar?: string;
  isCurrentUser?: boolean;
}

// AI助手类型
export interface Assistant {
  id: string;
  key?: string; // 助手的唯一标识符
  title: string;
  description: string;
  iconType: string;
  iconColor: string;
  category: string;
  chats?: ChatData[]; // 与该助手的历史对话列表
}

// 聊天通道类型 (对应后端 Chat 模型)
export interface Channel {
  id: string; // 聊天会话的唯一标识符
  name: string; // 会话标题
  assistantId?: string; // 关联助手的 ID
  messages: Message[];
  connected: boolean;
  error: string | null;
  websocket: WebSocket | null;
  isPinned?: boolean; // 是否置顶
  createdAt?: string; // 创建时间
  updatedAt?: string; // 更新时间
}

// 历史会话摘要，用于显示在历史列表中
export interface ChatHistoryItem {
  id: string; // 聊天会话ID
  title: string; // 会话标题
  assistantId: string; // 关联助手ID
  assistantTitle: string; // 助手名称
  lastMessage?: string; // 最后一条消息的内容摘要
  createdAt: string; // 创建时间
  updatedAt: string; // 更新时间
  isPinned: boolean; // 是否置顶
  messageCount: number; // 消息数量
}

// Xyzen Redux状态类型
export interface XyzenStateProps {
  isXyzenOpen: boolean;
  panelWidth: number;
  activeChannelUUID: string | null;
  assistants: Assistant[];
  assistantsLoading: boolean;
  assistantsError: string | null;
  channels: Record<string, Channel>;
  searchQuery: string;
  user: {
    username: string;
    avatar: string;
  };
  view: 'nodes' | 'chat' | 'workflows' | 'history'; // 当前视图：节点列表、聊天或历史记录
  chatHistory: ChatHistoryItem[]; // 聊天历史列表
  chatHistoryLoading: boolean;
  chatHistoryError: string | null;
}

// 从后端API返回的聊天会话数据类型
export interface ChatData {
  id: string;
  title: string;
  assistant?: string; // 助手ID
  assistant_name?: string; // 助手名称
  messages_count: number;
  last_message?: {
    content: string;
    timestamp: string;
  };
  created_at: string;
  updated_at: string;
  is_pinned: boolean;
}

// WebSocket连接类型
export interface WebSocketConnection {
  socket: WebSocket;
  channelUUID: string;
  reconnectAttempts: number;
  reconnectTimeout: NodeJS.Timeout | null;
  reconnecting: boolean;
}

// 工作流项目类型 (用于XyzenWorkflows组件)
export interface WorkflowItem {
  id: string;
  title: string;
  description: string;
  author: string;
  created: Date;
  lastUpdated: Date;
  nodeCount: number;
}

// Tab选项接口 (用于Xyzen组件)
export interface TabItem {
  id: string;
  title: string;
}

// 创建聊天会话参数接口
export interface CreateChatParams {
  uuid: string;
  assistantId?: string;
  title?: string;
}

// Redux Action Payload 类型
export interface SetActiveChannelPayload {
  channelUUID: string;
}

export interface SetChannelConnectedPayload {
  channelUUID: string;
  connected: boolean;
}

export interface SetChannelErrorPayload {
  channelUUID: string;
  error: string | null;
}

export interface AddMessagePayload {
  channelUUID: string;
  message: {
    id?: string;
    sender: string;
    content: string;
    timestamp: string;
    avatar: string;
    isCurrentUser: boolean;
  };
}

// 聊天状态更新接口
export interface ChatStatusUpdateParams {
  chatId: string;
  isTyping?: boolean;
  isOnline?: boolean;
}

// 聊天消息参数接口
export interface SendMessageParams {
  channelUUID: string;
  message: string;
  context?: any;
}

// 来自服务器的消息响应接口
export interface ServerMessageResponse {
  new_message?: {
    id: string;
    sender: string;
    content: string;
    timestamp: string;
    avatar?: string;
    is_ai: boolean;
  };
  error?: string;
  status?: string;
  message?: string;
  result?: any;
}
