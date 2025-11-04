/**
 * 📄 ActionRunnerDialog 组件
 *
 * 职责：执行设备动作的对话框
 *
 * 功能：
 * 1. 显示选中的动作信息
 * 2. 配置动作参数（Monaco Editor）
 * 3. 执行动作并返回结果
 * 4. 传递正确的 device_id（Material.name）
 */

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { config } from '@/configs';
import apiClient from '@/service/http/client';
import { getAuthenticatedWsUrl } from '@/service/ws/client';
import { useActionLogStore } from '@/store/actionLogStore';
import type { DeviceActionInfo, Material } from '@/types/material';
import Editor, { loader } from '@monaco-editor/react';
import {
  AlertCircle,
  CheckCircle2,
  Copy,
  Loader2,
  Play,
  RefreshCw,
  XCircle,
} from 'lucide-react';
import * as monaco from 'monaco-editor';
import { useCallback, useEffect, useRef, useState } from 'react';
import useWebSocket, { ReadyState } from 'react-use-websocket';

// 配置 Monaco Editor
loader.config({ monaco });

interface ActionResult {
  job_id: string;
  task_id: string;
  device_id: string;
  action_name: string;
  status: string;
  feedback_data?: Record<string, unknown>;
  return_info?: Record<string, unknown>;
}

interface ActionRunnerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  material: Material;
  action: DeviceActionInfo;
  labUuid: string;
  onExecutionComplete?: (result: {
    task_uuid: string;
    status: string;
    result?: ActionResult;
  }) => void;
}

export default function ActionRunnerDialog({
  open,
  onOpenChange,
  material,
  action,
  labUuid,
  onExecutionComplete,
}: ActionRunnerDialogProps) {
  const [paramJson, setParamJson] = useState<string>('{}');
  const [result, setResult] = useState<ActionResult | null>(null);
  const [taskUuid, setTaskUuid] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [isQuerying, setIsQuerying] = useState(false);
  const [error, setError] = useState<string>('');
  const [wsUrl, setWsUrl] = useState<string | null>(null);
  const [executionLogs, setExecutionLogs] = useState<
    Array<{
      timestamp: string;
      message: string;
      type: 'info' | 'success' | 'error' | 'warning';
    }>
  >([]);

  // 使用 ref 追踪上一次的状态，避免触发重渲染
  const lastStatusRef = useRef<string>('');

  // 日志管理
  const { addLog, updateLog } = useActionLogStore();

  // 添加执行日志 - 使用 useCallback 稳定引用
  const addExecutionLog = useCallback(
    (
      message: string,
      type: 'info' | 'success' | 'error' | 'warning' = 'info'
    ) => {
      setExecutionLogs((prev) => [
        ...prev,
        {
          timestamp: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
          message,
          type,
        },
      ]);
    },
    []
  );

  // 使用 react-use-websocket hook
  const { lastMessage, readyState } = useWebSocket(
    wsUrl,
    {
      shouldReconnect: () => false, // 不自动重连，任务完成后手动关闭
      reconnectAttempts: 0,
    },
    !!wsUrl // 只有当 wsUrl 不为 null 时才连接
  );

  // 当 action 变化时，自动填充默认参数
  useEffect(() => {
    if (action?.goal_default) {
      setParamJson(JSON.stringify(action.goal_default, null, 2));
    } else if (action?.schema) {
      // 从 schema 生成示例参数
      const example = generateExampleFromSchema(action.schema);
      setParamJson(JSON.stringify(example, null, 2));
    } else {
      setParamJson('{}');
    }
    // 重置状态
    setResult(null);
    setTaskUuid('');
    setError('');
    setExecutionLogs([]);
    lastStatusRef.current = '';
  }, [action]);

  // 从 schema 生成示例参数
  const generateExampleFromSchema = (
    schema: unknown
  ): Record<string, unknown> => {
    if (!schema || typeof schema !== 'object') return {};

    const schemaObj = schema as Record<string, unknown>;
    const properties = schemaObj.properties as
      | Record<string, unknown>
      | undefined;

    if (!properties) return {};

    const example: Record<string, unknown> = {};

    Object.entries(properties).forEach(([key, prop]) => {
      if (!prop || typeof prop !== 'object') return;

      const propObj = prop as Record<string, unknown>;
      const type = propObj.type as string | undefined;
      const defaultValue = propObj.default;

      if (defaultValue !== undefined) {
        example[key] = defaultValue;
      } else {
        switch (type) {
          case 'string':
            example[key] = '';
            break;
          case 'number':
          case 'integer':
            example[key] = 0;
            break;
          case 'boolean':
            example[key] = false;
            break;
          case 'array':
            example[key] = [];
            break;
          case 'object':
            example[key] = {};
            break;
          default:
            example[key] = null;
        }
      }
    });

    return example;
  };

  // 格式化 JSON
  const formatJson = () => {
    try {
      const parsed = JSON.parse(paramJson);
      setParamJson(JSON.stringify(parsed, null, 2));
      setError('');
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setError(`格式化失败: ${message}`);
    }
  };

  // 执行动作
  const handleRunAction = async () => {
    setError('');
    setResult(null);
    setTaskUuid('');

    // 验证和解析参数JSON
    let param: Record<string, unknown>;
    try {
      param = JSON.parse(paramJson);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setError(`参数JSON格式错误: ${message}`);
      return;
    }

    // 构建请求数据 - 使用 Material.id 作为正确的 device_id
    const requestData = {
      lab_uuid: labUuid,
      device_id: material.id, // 使用 Material.id 作为设备 ID
      action: action.name,
      action_type: action.type,
      param,
    };

    console.log('执行动作请求:', requestData);
    console.log('Material 完整数据:', {
      id: material.id,
      name: material.name,
      uuid: material.uuid,
      type: material.type,
      class: material.class,
    });

    // 添加日志
    addExecutionLog(`准备执行动作: ${action.name}`, 'info');
    addExecutionLog(`设备 ID: ${material.id}`, 'info');

    // 发送请求
    setIsLoading(true);
    try {
      const response = await apiClient.post(
        `${config.apiBaseUrl}/api/v1/lab/action/run`,
        requestData
      );

      if (response.data.code === 0) {
        const uuid = response.data.data?.task_uuid;
        setTaskUuid(uuid);
        console.log('任务已创建', `任务 UUID: ${uuid}`);
        addExecutionLog(`✓ 任务创建成功: ${uuid}`, 'success');

        // 记录日志（异步）
        await addLog({
          taskUuid: uuid,
          labUuid: labUuid,
          deviceId: material.id,
          deviceName: material.name,
          actionName: action.name,
          status: 'pending',
          startTime: new Date().toISOString(),
        });

        // 通知父组件执行已开始
        if (onExecutionComplete) {
          onExecutionComplete({
            task_uuid: uuid,
            status: 'pending',
          });
        }

        // 连接 WebSocket 接收实时状态更新（带认证 token）
        const authenticatedWsUrl = getAuthenticatedWsUrl(
          `/api/v1/ws/action/${uuid}`
        );
        console.log('连接 WebSocket:', authenticatedWsUrl);
        addExecutionLog('正在连接 WebSocket 获取实时状态...', 'info');
        setWsUrl(authenticatedWsUrl);
      } else {
        const errMsg =
          response.data?.msg ||
          response.data?.error?.msg ||
          response.data?.message ||
          response.data?.error?.message ||
          '未知错误';
        setError(`请求失败: ${errMsg}`);
        addExecutionLog(`✗ 请求失败: ${errMsg}`, 'error');
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '未知错误';
      setError(`网络错误: ${message}`);
      addExecutionLog(`✗ 网络错误: ${message}`, 'error');
    } finally {
      setIsLoading(false);
    }
  };

  // 查询结果
  const queryResult = async (uuid?: string) => {
    const queryUuid = uuid || taskUuid;
    if (!queryUuid) {
      setError('请先执行动作以获取任务 UUID');
      return;
    }

    setIsQuerying(true);
    setError('');
    try {
      const response = await apiClient.get(
        `${config.apiBaseUrl}/api/v1/lab/action/result/${queryUuid}`
      );

      if (response.data.code === 0) {
        const resultData = response.data.data;
        setResult(resultData);
        console.log('查询成功', `状态: ${resultData?.status || 'unknown'}`);

        // 通知父组件执行完成
        if (onExecutionComplete && resultData) {
          onExecutionComplete({
            task_uuid: queryUuid,
            status: resultData.status,
            result: resultData,
          });
        }
      } else {
        const errMsg =
          response.data?.msg ||
          response.data?.error?.msg ||
          response.data?.message ||
          response.data?.error?.message ||
          '未知错误';
        setError(`查询失败: ${errMsg}`);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '未知错误';
      setError(`查询错误: ${message}`);
    } finally {
      setIsQuerying(false);
    }
  };

  // 处理 WebSocket 消息
  useEffect(() => {
    if (!lastMessage) return;

    try {
      const data = JSON.parse(lastMessage.data);
      console.log(
        '收到 WebSocket 消息 - 完整结构:',
        JSON.stringify(data, null, 2)
      );
      console.log('data.code:', data.code);
      console.log('data.data:', data.data);
      console.log('data.data?.data:', data.data?.data);
      console.log('data.data?.data?.action:', data.data?.data?.action);

      // 尝试多种可能的消息结构
      let jobData = null;

      // 结构1: data.data.data.data (嵌套4层)
      if (
        data.code === 0 &&
        data.data?.data?.action === 'action_status_update'
      ) {
        jobData = data.data.data.data;
        console.log('匹配结构1 (4层嵌套)');
      }
      // 结构2: data.data.data (嵌套3层)
      else if (
        data.code === 0 &&
        data.data?.action === 'action_status_update'
      ) {
        jobData = data.data.data;
        console.log('匹配结构2 (3层嵌套)');
      }
      // 结构3: data.data (嵌套2层)
      else if (data.code === 0 && data.action === 'action_status_update') {
        jobData = data.data;
        console.log('匹配结构3 (2层嵌套)');
      }
      // 结构4: 直接在 data 中
      else if (data.action === 'action_status_update') {
        jobData = data;
        console.log('匹配结构4 (1层)');
      }

      if (jobData) {
        console.log('解析到的 jobData:', jobData);

        // 检查状态是否变化
        const statusChanged = jobData.status !== lastStatusRef.current;

        // 只在状态变化时添加日志，避免重复日志导致性能问题
        if (statusChanged) {
          const statusEmojiMap: Record<string, string> = {
            pending: '⏳',
            running: '▶️',
            success: '✓',
            failed: '✗',
            fail: '✗',
          };
          const statusEmoji = statusEmojiMap[jobData.status] || '●';

          addExecutionLog(
            `${statusEmoji} 状态: ${jobData.status}`,
            jobData.status === 'success'
              ? 'success'
              : jobData.status === 'failed' || jobData.status === 'fail'
              ? 'error'
              : jobData.status === 'running'
              ? 'warning'
              : 'info'
          );
        }

        // 更新结果
        const newResult: ActionResult = {
          job_id: jobData.job_id,
          task_id: jobData.task_id,
          device_id: jobData.device_id,
          action_name: jobData.action_name,
          status: jobData.status,
          feedback_data: jobData.feedback_data,
          return_info: jobData.return_info,
        };

        setResult(newResult);

        // 只在状态变化时更新日志存储，避免频繁更新（异步）
        if (statusChanged) {
          const now = new Date().toISOString();
          const isCompleted =
            jobData.status === 'success' ||
            jobData.status === 'failed' ||
            jobData.status === 'fail';

          // 异步更新，不阻塞 UI
          updateLog(taskUuid, {
            status: jobData.status,
            statusUpdate: {
              status: jobData.status,
              timestamp: now,
              feedbackData: jobData.feedback_data,
              returnInfo: jobData.return_info,
            },
            ...(isCompleted && {
              endTime: now,
              finalResult: {
                jobId: jobData.job_id,
                feedbackData: jobData.feedback_data,
                returnInfo: jobData.return_info,
              },
            }),
          }).catch((err) => {
            console.error('更新日志失败:', err);
          });

          // 更新引用，避免重复处理
          lastStatusRef.current = jobData.status;
        }

        // 通知父组件
        if (onExecutionComplete) {
          onExecutionComplete({
            task_uuid: taskUuid,
            status: jobData.status,
            result: newResult,
          });
        }

        // 如果任务完成，断开 WebSocket
        const isCompleted =
          jobData.status === 'success' ||
          jobData.status === 'failed' ||
          jobData.status === 'fail';

        if (isCompleted) {
          console.log('任务完成，断开 WebSocket');
          addExecutionLog(
            '任务执行完成',
            jobData.status === 'success' ? 'success' : 'error'
          );
          setWsUrl(null);
        }
      } else {
        console.warn('未匹配到任何消息结构，原始消息:', data);
        addExecutionLog('⚠ 收到未识别的消息格式', 'warning');
      }
    } catch (err) {
      console.error('解析 WebSocket 消息失败:', err);
      addExecutionLog(
        `✗ 消息解析失败: ${err instanceof Error ? err.message : String(err)}`,
        'error'
      );
    }
  }, [lastMessage, taskUuid, onExecutionComplete, updateLog, addExecutionLog]);

  // 监听 WebSocket 连接状态
  useEffect(() => {
    const statusText = {
      [ReadyState.CONNECTING]: '连接中...',
      [ReadyState.OPEN]: '已连接',
      [ReadyState.CLOSING]: '断开中...',
      [ReadyState.CLOSED]: '已断开',
      [ReadyState.UNINSTANTIATED]: '未实例化',
    }[readyState];

    console.log('WebSocket 状态:', statusText);

    if (readyState === ReadyState.OPEN && wsUrl) {
      addExecutionLog('✓ WebSocket 连接成功', 'success');
    } else if (readyState === ReadyState.CLOSED && wsUrl) {
      setError('WebSocket 连接已关闭');
      addExecutionLog('WebSocket 连接已关闭', 'warning');
    }
  }, [readyState, wsUrl, addExecutionLog]);

  // 对话框关闭时断开 WebSocket
  useEffect(() => {
    if (!open) {
      setWsUrl(null);
    }
  }, [open]);

  // 复制到剪贴板
  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    console.log('已复制到剪贴板');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange} size="5xl">
      <DialogContent className="max-w-6xl max-h-[90vh] p-0 gap-0">
        <DialogHeader className="px-6 pt-6 pb-4 border-b dark:border-neutral-800">
          <DialogTitle className="flex items-center gap-2">
            <Play className="h-5 w-5" />
            执行动作：{action.name}
          </DialogTitle>
          <DialogDescription className="space-y-1">
            <div className="text-neutral-600 dark:text-neutral-400">
              设备名称：
              <span className="text-neutral-900 dark:text-neutral-100">
                {material.name}
              </span>
            </div>
            <div className="text-xs text-neutral-500 dark:text-neutral-500">
              设备 ID:{' '}
              <code className="bg-neutral-100 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 px-1.5 py-0.5 rounded border border-neutral-200 dark:border-neutral-700">
                {material.id}
              </code>
              {' • '}
              类型:{' '}
              <span className="text-neutral-700 dark:text-neutral-300">
                {material.class || material.type}
              </span>
            </div>
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-2 gap-0 h-[calc(90vh-100px)]">
          {/* 左侧：Schema 参考 */}
          <div className="flex flex-col border-r border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-950">
            <div className="px-4 py-3 border-b border-neutral-200 dark:border-neutral-800 bg-neutral-50 dark:bg-neutral-900/50">
              <Label className="text-sm font-medium text-neutral-900 dark:text-neutral-100">
                Schema 参考
              </Label>
              <p className="text-xs text-neutral-500 dark:text-neutral-400 mt-1">
                参数结构说明
              </p>
            </div>
            <div className="flex-1 overflow-hidden">
              <Editor
                height="100%"
                defaultLanguage="json"
                value={JSON.stringify(action.schema || {}, null, 2)}
                options={{
                  readOnly: true,
                  minimap: { enabled: false },
                  fontSize: 13,
                  lineNumbers: 'on',
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  wordWrap: 'on',
                  tabSize: 2,
                  foldingStrategy: 'indentation',
                  showFoldingControls: 'mouseover',
                  glyphMargin: true,
                }}
                theme={'vs-dark'}
              />
            </div>
          </div>

          {/* 右侧：参数输入和执行结果 */}
          <div className="flex flex-col overflow-hidden bg-white dark:bg-neutral-950">
            <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
              {/* 参数编辑器 */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label className="text-sm font-medium text-neutral-900 dark:text-neutral-100">
                    动作参数
                  </Label>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={formatJson}
                    className="text-xs h-7 hover:bg-neutral-100 dark:hover:bg-neutral-800"
                  >
                    格式化
                  </Button>
                </div>
                <div className="border rounded-lg overflow-hidden border-neutral-300 dark:border-neutral-700">
                  <Editor
                    height="300px"
                    defaultLanguage="json"
                    value={paramJson}
                    onChange={(value) => setParamJson(value || '{}')}
                    theme={'vs-dark'}
                    options={{
                      minimap: { enabled: false },
                      fontSize: 13,
                      lineNumbers: 'on',
                      scrollBeyondLastLine: false,
                      automaticLayout: true,
                      tabSize: 2,
                      wordWrap: 'on',
                      foldingStrategy: 'indentation',
                      showFoldingControls: 'mouseover',
                      glyphMargin: true,
                    }}
                  />
                </div>
              </div>

              {/* 错误提示 */}
              {error && (
                <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg flex items-start gap-2">
                  <AlertCircle className="h-4 w-4 text-red-600 dark:text-red-400 mt-0.5 flex-shrink-0" />
                  <p className="text-sm text-red-600 dark:text-red-400">
                    {error}
                  </p>
                </div>
              )}

              {/* 执行按钮 */}
              <Button
                onClick={handleRunAction}
                disabled={isLoading}
                className="w-full"
                size="lg"
              >
                {isLoading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    执行中...
                  </>
                ) : (
                  <>
                    <Play className="mr-2 h-4 w-4" />
                    执行动作
                  </>
                )}
              </Button>

              {/* Task UUID */}
              {taskUuid && (
                <div className="space-y-3 p-4 bg-gradient-to-br from-indigo-50 to-purple-50 dark:from-indigo-950/40 dark:to-purple-950/40 rounded-lg border border-indigo-200 dark:border-indigo-800/60">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className="h-2 w-2 rounded-full bg-indigo-500 dark:bg-indigo-400 animate-pulse" />
                      <Label className="text-sm font-semibold text-indigo-900 dark:text-indigo-200">
                        任务 UUID
                      </Label>
                    </div>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyToClipboard(taskUuid)}
                        className="h-8 text-indigo-700 dark:text-indigo-300 hover:bg-indigo-100 dark:hover:bg-indigo-900/50 hover:text-indigo-900 dark:hover:text-indigo-100"
                        title="复制 UUID"
                      >
                        <Copy className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => queryResult()}
                        disabled={isQuerying}
                        className="h-8 text-indigo-700 dark:text-indigo-300 hover:bg-indigo-100 dark:hover:bg-indigo-900/50 hover:text-indigo-900 dark:hover:text-indigo-100 disabled:opacity-50"
                        title="刷新结果"
                      >
                        {isQuerying ? (
                          <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        ) : (
                          <RefreshCw className="h-3.5 w-3.5" />
                        )}
                      </Button>
                    </div>
                  </div>
                  <div className="relative group">
                    <div className="font-mono text-xs text-indigo-900 dark:text-indigo-100 bg-white/80 dark:bg-neutral-900/60 px-3 py-2.5 rounded border border-indigo-200/60 dark:border-indigo-700/60 break-all select-all backdrop-blur-sm">
                      {taskUuid}
                    </div>
                    <div className="absolute inset-0 rounded bg-gradient-to-r from-indigo-500/0 via-indigo-500/10 dark:via-indigo-400/10 to-indigo-500/0 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
                  </div>
                </div>
              )}

              {/* 执行日志 */}
              {executionLogs.length > 0 && (
                <div className="space-y-2">
                  <Label className="text-sm font-medium text-neutral-900 dark:text-neutral-100">
                    执行日志
                  </Label>
                  <div className="max-h-40 overflow-y-auto bg-neutral-900 dark:bg-neutral-950 rounded-lg border border-neutral-700 dark:border-neutral-800 p-3 space-y-1.5 font-mono text-xs">
                    {executionLogs.map((log, index) => (
                      <div
                        key={index}
                        className={`flex items-start gap-2 ${
                          log.type === 'error'
                            ? 'text-red-400'
                            : log.type === 'success'
                            ? 'text-green-400'
                            : log.type === 'warning'
                            ? 'text-yellow-400'
                            : 'text-neutral-300'
                        }`}
                      >
                        <span className="text-neutral-500 shrink-0">
                          [{log.timestamp}]
                        </span>
                        <span className="break-all">{log.message}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* 执行结果 */}
              {result && (
                <div className="space-y-3 p-4 bg-neutral-50 dark:bg-neutral-900/50 rounded-lg border border-neutral-200 dark:border-neutral-800">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {result.status === 'success' ? (
                        <CheckCircle2 className="h-5 w-5 text-green-500 dark:text-green-400" />
                      ) : result.status === 'fail' ||
                        result.status === 'failed' ? (
                        <XCircle className="h-5 w-5 text-red-500 dark:text-red-400" />
                      ) : (
                        <Loader2 className="h-5 w-5 text-yellow-500 dark:text-yellow-400 animate-spin" />
                      )}
                      <span className="font-semibold capitalize text-neutral-900 dark:text-neutral-100">
                        执行结果: {result.status}
                      </span>
                    </div>
                    {(result.feedback_data || result.return_info) && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          copyToClipboard(
                            JSON.stringify(
                              result.return_info || result.feedback_data,
                              null,
                              2
                            )
                          )
                        }
                        className="h-7 border-neutral-300 dark:border-neutral-700 hover:bg-neutral-100 dark:hover:bg-neutral-800"
                      >
                        <Copy className="mr-1 h-3 w-3" />
                        复制
                      </Button>
                    )}
                  </div>

                  {/* 基本信息 - 使用 Monaco */}
                  <div className="space-y-2">
                    <Label className="text-xs font-medium text-neutral-700 dark:text-neutral-300">
                      基本信息
                    </Label>
                    <div className="border rounded overflow-hidden border-neutral-300 dark:border-neutral-700">
                      <Editor
                        height="100px"
                        defaultLanguage="json"
                        value={JSON.stringify(
                          {
                            job_id: result.job_id,
                            device_id: result.device_id,
                            action: result.action_name,
                          },
                          null,
                          2
                        )}
                        options={{
                          readOnly: true,
                          minimap: { enabled: false },
                          fontSize: 12,
                          lineNumbers: 'off',
                          wordWrap: 'on',
                          scrollBeyondLastLine: false,
                          automaticLayout: true,
                          folding: false,
                        }}
                        theme={'vs-dark'}
                      />
                    </div>
                  </div>

                  {/* 返回数据 - 使用 Monaco */}
                  {(result.feedback_data || result.return_info) && (
                    <div className="space-y-2">
                      <Label className="text-xs font-medium text-neutral-700 dark:text-neutral-300">
                        返回数据
                      </Label>
                      <div className="border rounded overflow-hidden border-neutral-300 dark:border-neutral-700">
                        <Editor
                          height="250px"
                          defaultLanguage="json"
                          value={JSON.stringify(
                            result.return_info || result.feedback_data,
                            null,
                            2
                          )}
                          options={{
                            readOnly: true,
                            minimap: { enabled: false },
                            fontSize: 12,
                            lineNumbers: 'on',
                            scrollBeyondLastLine: false,
                            automaticLayout: true,
                            folding: true,
                            wordWrap: 'on',
                            foldingStrategy: 'indentation',
                            showFoldingControls: 'mouseover',
                            glyphMargin: true,
                          }}
                          theme={'vs-dark'}
                        />
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
