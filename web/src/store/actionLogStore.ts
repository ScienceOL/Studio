/**
 * 🎯 Action 执行日志 Store
 *
 * 职责：
 * 1. 存储动作执行的完整历史记录（使用 IndexedDB）
 * 2. 记录每个状态变化的时间戳
 * 3. 提供日志查询和过滤功能
 * 4. 自动清理超过 500 条的旧日志
 */

import { create } from 'zustand';

export interface ActionLogEntry {
  id: string; // 唯一标识
  taskUuid: string; // 任务 UUID
  labUuid: string; // 实验室 UUID
  deviceId: string; // 设备 ID
  deviceName?: string; // 设备名称
  actionName: string; // 动作名称
  status: 'pending' | 'running' | 'success' | 'failed' | 'fail';
  startTime: string; // ISO 8601 格式
  endTime?: string; // ISO 8601 格式
  duration?: number; // 持续时间（毫秒）

  // 状态变化历史
  statusHistory: {
    status: string;
    timestamp: string;
    feedbackData?: Record<string, unknown>;
    returnInfo?: Record<string, unknown>;
  }[];

  // 最终结果
  finalResult?: {
    jobId: string;
    feedbackData?: Record<string, unknown>;
    returnInfo?: Record<string, unknown>;
  };

  // 错误信息
  error?: string;
}

interface ActionLogState {
  logs: ActionLogEntry[];
  maxLogs: number; // 最大保存日志数量

  // Actions (异步操作返回 Promise)
  addLog: (log: Omit<ActionLogEntry, 'id' | 'statusHistory'>) => Promise<void>;
  updateLog: (
    taskUuid: string,
    update: {
      status?: ActionLogEntry['status'];
      endTime?: string;
      duration?: number;
      finalResult?: ActionLogEntry['finalResult'];
      error?: string;
      statusUpdate?: {
        status: string;
        timestamp: string;
        feedbackData?: Record<string, unknown>;
        returnInfo?: Record<string, unknown>;
      };
    }
  ) => Promise<void>;
  getLog: (taskUuid: string) => ActionLogEntry | undefined;
  getLogs: (filters?: {
    deviceId?: string;
    status?: string;
    startDate?: string;
    endDate?: string;
  }) => ActionLogEntry[];
  clearLogs: () => Promise<void>;
  deleteLog: (taskUuid: string) => Promise<void>;
}

// IndexedDB 配置
const DB_NAME = 'ActionLogDB';
const DB_VERSION = 1;
const STORE_NAME = 'logs';
const MAX_LOGS = 500; // 最大保存 500 条日志

// IndexedDB 工具类
class ActionLogDB {
  private db: IDBDatabase | null = null;

  async init(): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        this.db = request.result;
        resolve();
      };

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;

        // 创建对象存储
        if (!db.objectStoreNames.contains(STORE_NAME)) {
          const objectStore = db.createObjectStore(STORE_NAME, {
            keyPath: 'id',
          });
          // 创建索引
          objectStore.createIndex('taskUuid', 'taskUuid', { unique: false });
          objectStore.createIndex('labUuid', 'labUuid', { unique: false });
          objectStore.createIndex('startTime', 'startTime', { unique: false });
          objectStore.createIndex('status', 'status', { unique: false });
        }
      };
    });
  }

  async getAllLogs(): Promise<ActionLogEntry[]> {
    if (!this.db) await this.init();

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([STORE_NAME], 'readonly');
      const objectStore = transaction.objectStore(STORE_NAME);
      const request = objectStore.getAll();

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        const logs = request.result as ActionLogEntry[];
        // 按开始时间倒序排序
        logs.sort(
          (a, b) =>
            new Date(b.startTime).getTime() - new Date(a.startTime).getTime()
        );
        resolve(logs);
      };
    });
  }

  async addLog(log: ActionLogEntry): Promise<void> {
    if (!this.db) await this.init();

    // 先检查并清理旧日志
    await this.cleanupOldLogs();

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
      const objectStore = transaction.objectStore(STORE_NAME);
      const request = objectStore.add(log);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  async updateLog(id: string, updates: Partial<ActionLogEntry>): Promise<void> {
    if (!this.db) await this.init();

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
      const objectStore = transaction.objectStore(STORE_NAME);
      const getRequest = objectStore.get(id);

      getRequest.onerror = () => reject(getRequest.error);
      getRequest.onsuccess = () => {
        const log = getRequest.result;
        if (log) {
          Object.assign(log, updates);
          const putRequest = objectStore.put(log);
          putRequest.onerror = () => reject(putRequest.error);
          putRequest.onsuccess = () => resolve();
        } else {
          resolve();
        }
      };
    });
  }

  async getLogByTaskUuid(
    taskUuid: string
  ): Promise<ActionLogEntry | undefined> {
    if (!this.db) await this.init();

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([STORE_NAME], 'readonly');
      const objectStore = transaction.objectStore(STORE_NAME);
      const index = objectStore.index('taskUuid');
      const request = index.get(taskUuid);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve(request.result);
    });
  }

  async deleteLog(taskUuid: string): Promise<void> {
    if (!this.db) await this.init();

    const logs = await this.getAllLogs();
    const log = logs.find((l) => l.taskUuid === taskUuid);
    if (!log) return;

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
      const objectStore = transaction.objectStore(STORE_NAME);
      const request = objectStore.delete(log.id);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  async clearAll(): Promise<void> {
    if (!this.db) await this.init();

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
      const objectStore = transaction.objectStore(STORE_NAME);
      const request = objectStore.clear();

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  // 清理超过限制的旧日志
  async cleanupOldLogs(): Promise<void> {
    const logs = await this.getAllLogs();
    if (logs.length >= MAX_LOGS) {
      // 删除最旧的日志（保留最新的 MAX_LOGS - 1 条）
      const logsToDelete = logs.slice(MAX_LOGS - 1);

      return new Promise((resolve, reject) => {
        const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
        const objectStore = transaction.objectStore(STORE_NAME);

        let completed = 0;
        logsToDelete.forEach((log) => {
          const request = objectStore.delete(log.id);
          request.onsuccess = () => {
            completed++;
            if (completed === logsToDelete.length) resolve();
          };
          request.onerror = () => reject(request.error);
        });

        if (logsToDelete.length === 0) resolve();
      });
    }
  }
}

const dbInstance = new ActionLogDB();

// Zustand Store
export const useActionLogStore = create<ActionLogState>((set, get) => ({
  logs: [],
  maxLogs: MAX_LOGS,

  addLog: async (log) => {
    const newLog: ActionLogEntry = {
      ...log,
      id: `log_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      statusHistory: [
        {
          status: log.status,
          timestamp: log.startTime,
        },
      ],
    };

    await dbInstance.addLog(newLog);
    const logs = await dbInstance.getAllLogs();
    set({ logs });
  },

  updateLog: async (taskUuid, update) => {
    const logs = get().logs;
    const log = logs.find((l) => l.taskUuid === taskUuid);
    if (!log) return;

    const updatedFields: Partial<ActionLogEntry> = {};

    if (update.status) {
      updatedFields.status = update.status;
    }
    if (update.endTime) {
      updatedFields.endTime = update.endTime;
    }
    if (update.duration !== undefined) {
      updatedFields.duration = update.duration;
    }
    if (update.finalResult) {
      updatedFields.finalResult = update.finalResult;
    }
    if (update.error) {
      updatedFields.error = update.error;
    }
    if (update.statusUpdate) {
      updatedFields.statusHistory = [...log.statusHistory, update.statusUpdate];
    }

    await dbInstance.updateLog(log.id, updatedFields);
    const updatedLogs = await dbInstance.getAllLogs();
    set({ logs: updatedLogs });
  },

  getLog: (taskUuid) => {
    return get().logs.find((log) => log.taskUuid === taskUuid);
  },

  getLogs: (filters) => {
    let logs = get().logs;

    if (!filters) return logs;

    if (filters.deviceId) {
      logs = logs.filter((log) => log.deviceId === filters.deviceId);
    }

    if (filters.status) {
      logs = logs.filter((log) => log.status === filters.status);
    }

    if (filters.startDate) {
      logs = logs.filter((log) => log.startTime >= filters.startDate!);
    }

    if (filters.endDate) {
      logs = logs.filter(
        (log) => log.endTime && log.endTime <= filters.endDate!
      );
    }

    return logs;
  },

  clearLogs: async () => {
    await dbInstance.clearAll();
    set({ logs: [] });
  },

  deleteLog: async (taskUuid) => {
    await dbInstance.deleteLog(taskUuid);
    const logs = await dbInstance.getAllLogs();
    set({ logs });
  },
}));

// 清理旧的 localStorage 数据
function cleanupOldLocalStorage() {
  try {
    const oldDataKey = 'action-log-storage';
    if (localStorage.getItem(oldDataKey)) {
      localStorage.removeItem(oldDataKey);
      console.log('🗑️ 已清除旧的 localStorage 日志数据');
    }
  } catch (error) {
    console.error('清理旧数据时出错:', error);
  }
}

// 初始化：加载 IndexedDB 中的数据
dbInstance.init().then(async () => {
  // 清除旧的 localStorage 数据
  cleanupOldLocalStorage();

  // 加载所有日志
  const logs = await dbInstance.getAllLogs();
  useActionLogStore.setState({ logs });

  console.log(`📊 已加载 ${logs.length} 条日志记录（IndexedDB）`);
});
