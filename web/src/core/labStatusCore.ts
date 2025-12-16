/**
 * 🎯 Core Layer - Lab Status 核心业务逻辑
 *
 * 职责：
 * 1. 管理全局唯一的 WebSocket 连接
 * 2. 处理实验室状态更新逻辑
 * 3. 更新 Store 状态
 * 4. 提供统一的 API 接口
 *
 * 注意：
 * - 使用单例模式，确保全局只有一个 WebSocket 连接
 * - 所有状态存储在 environmentStore 中
 * - 组件通过 useLabStatus hook 订阅状态变化
 */

import { getAuthenticatedWsUrl } from '@/service/ws/client';
import { useEnvironmentStore } from '@/store/environmentStore';
import { v4 as uuidv4 } from 'uuid';

export interface LabStatusData {
  lab_uuid: string;
  is_online: boolean;
  last_connected_at?: string;
}

interface WebSocketMessage {
  code: number;
  data: {
    action: string;
    msg_uuid: string;
    data?: LabStatusData[] | LabStatusData;
  };
  timestamp: number;
}

type StatusUpdateCallback = (statuses: LabStatusData[]) => void;

class LabStatusManager {
  private ws: WebSocket | null = null;
  private reconnectTimer: number | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectInterval = 3000;
  private pendingRequests = new Map<string, (data: unknown) => void>();
  private callbacks = new Set<StatusUpdateCallback>();
  private isConnecting = false;

  constructor() {
    console.log('🚀 [LabStatusCore] Manager initialized');
  }

  /**
   * 连接 WebSocket
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN || this.isConnecting) {
      console.log('⚠️ [LabStatusCore] Already connected or connecting');
      return;
    }

    this.isConnecting = true;
    const wsUrl = getAuthenticatedWsUrl('/api/v1/ws/lab/status');
    console.log('🔌 [LabStatusCore] Connecting to:', wsUrl);

    try {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('🟢 [LabStatusCore] Connected');
        this.isConnecting = false;
        this.reconnectAttempts = 0;

        // 更新 store 连接状态
        useEnvironmentStore.getState().setLabStatusConnected(true);
      };

      this.ws.onmessage = (event) => {
        this.handleMessage(event.data);
      };

      this.ws.onerror = (error) => {
        console.error('❌ [LabStatusCore] WebSocket error:', error);
        this.isConnecting = false;
      };

      this.ws.onclose = (event) => {
        console.log(
          '🔴 [LabStatusCore] Disconnected:',
          event.code,
          event.reason
        );
        this.isConnecting = false;
        this.ws = null;

        // 更新 store 连接状态
        useEnvironmentStore.getState().setLabStatusConnected(false);

        // 尝试重连
        this.scheduleReconnect();
      };
    } catch (error) {
      console.error('❌ [LabStatusCore] Failed to create WebSocket:', error);
      this.isConnecting = false;
      this.scheduleReconnect();
    }
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    console.log('🔌 [LabStatusCore] Disconnecting...');

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    useEnvironmentStore.getState().setLabStatusConnected(false);
  }

  /**
   * 安排重连
   */
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error(
        '❌ [LabStatusCore] Max reconnect attempts reached, giving up'
      );
      return;
    }

    if (this.reconnectTimer) {
      return;
    }

    this.reconnectAttempts++;
    console.log(
      `⏳ [LabStatusCore] Reconnecting in ${this.reconnectInterval}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
    );

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, this.reconnectInterval);
  }

  /**
   * 处理接收到的消息
   */
  private handleMessage(data: string): void {
    console.log('📨 [LabStatusCore] Raw message:', data);

    try {
      const message: WebSocketMessage = JSON.parse(data);
      console.log('📦 [LabStatusCore] Parsed message:', message);

      if (message.code !== 0) {
        console.error('❌ [LabStatusCore] Error response:', message);
        return;
      }

      const { action, msg_uuid, data: responseData } = message.data;
      console.log(`🎯 [LabStatusCore] Action: ${action}, MsgUUID: ${msg_uuid}`);

      // 处理请求响应
      const resolver = this.pendingRequests.get(msg_uuid);
      if (resolver && responseData) {
        console.log(`✅ [LabStatusCore] Resolved request: ${msg_uuid}`);
        resolver(responseData);
        this.pendingRequests.delete(msg_uuid);
      }

      // 处理状态更新通知
      if (action === 'status_update') {
        console.log('🔔 [LabStatusCore] Received status update action');
        if (Array.isArray(responseData)) {
          console.log(
            '🔔 [LabStatusCore] Processing status update array:',
            responseData
          );
          this.handleStatusUpdate(responseData);
        } else {
          console.warn(
            '⚠️ [LabStatusCore] Status update data is not an array:',
            responseData
          );
        }
      }
    } catch (error) {
      console.error('❌ [LabStatusCore] Failed to parse message:', error);
    }
  }

  /**
   * 处理状态更新
   */
  private handleStatusUpdate(statuses: LabStatusData[]): void {
    console.log('🔔 [LabStatusCore] Status update:', statuses);

    // 更新 store
    const store = useEnvironmentStore.getState();
    statuses.forEach((status) => {
      store.updateLabStatus(status.lab_uuid, status);
    });

    // 触发回调
    this.callbacks.forEach((callback) => {
      try {
        callback(statuses);
      } catch (error) {
        console.error('❌ [LabStatusCore] Callback error:', error);
      }
    });

    console.log('✨ [LabStatusCore] Status update completed');
  }

  /**
   * 发送请求
   */
  private sendRequest<T>(action: string, data?: unknown): Promise<T> {
    return new Promise((resolve, reject) => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        reject(new Error('WebSocket is not connected'));
        return;
      }

      const msgUuid = uuidv4();
      const message: Record<string, unknown> = {
        action,
        msg_uuid: msgUuid,
      };
      if (data) {
        message.data = data;
      }

      // 注册响应处理器
      this.pendingRequests.set(msgUuid, resolve as (data: unknown) => void);

      // 设置超时
      setTimeout(() => {
        if (this.pendingRequests.has(msgUuid)) {
          this.pendingRequests.delete(msgUuid);
          reject(new Error('Request timeout'));
        }
      }, 10000);

      const messageStr = JSON.stringify(message);
      console.log(`📤 [LabStatusCore] Sending ${action}:`, messageStr);
      this.ws.send(messageStr);
    });
  }

  /**
   * 查询所有实验室状态
   */
  async queryList(): Promise<LabStatusData[]> {
    console.log('🔍 [LabStatusCore] Querying lab list...');
    try {
      const data = await this.sendRequest<LabStatusData[]>('query_list');
      console.log(`✅ [LabStatusCore] Received ${data.length} lab(s)`);

      // 更新 store
      const store = useEnvironmentStore.getState();
      data.forEach((status) => {
        store.updateLabStatus(status.lab_uuid, status);
      });

      return data;
    } catch (error) {
      console.error('❌ [LabStatusCore] Failed to query list:', error);
      throw error;
    }
  }

  /**
   * 查询单个实验室状态
   */
  async queryDetail(labUuid: string): Promise<LabStatusData> {
    console.log(`🔍 [LabStatusCore] Querying lab detail: ${labUuid}`);
    try {
      const data = await this.sendRequest<LabStatusData>('query_detail', {
        lab_uuid: labUuid,
      });
      console.log(`✅ [LabStatusCore] Received lab detail:`, data);

      // 更新 store
      useEnvironmentStore.getState().updateLabStatus(labUuid, data);

      return data;
    } catch (error) {
      console.error('❌ [LabStatusCore] Failed to query detail:', error);
      throw error;
    }
  }

  /**
   * 订阅状态更新
   */
  subscribe(callback: StatusUpdateCallback): () => void {
    console.log('📡 [LabStatusCore] Adding subscriber');
    this.callbacks.add(callback);

    // 返回取消订阅函数
    return () => {
      console.log('📡 [LabStatusCore] Removing subscriber');
      this.callbacks.delete(callback);
    };
  }

  /**
   * 获取连接状态
   */
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// 导出单例实例
export const LabStatusCore = new LabStatusManager();
