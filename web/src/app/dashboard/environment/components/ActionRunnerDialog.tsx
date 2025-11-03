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
import { useEffect, useState } from 'react';

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

  // 检测系统主题
  const [isDarkMode, setIsDarkMode] = useState(() => {
    if (typeof window === 'undefined') return false;
    return (
      document.documentElement.classList.contains('dark') ||
      window.matchMedia('(prefers-color-scheme: dark)').matches
    );
  });

  // 监听主题变化
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const updateTheme = () => {
      setIsDarkMode(
        document.documentElement.classList.contains('dark') ||
          window.matchMedia('(prefers-color-scheme: dark)').matches
      );
    };

    const observer = new MutationObserver(updateTheme);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    });

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', updateTheme);

    return () => {
      observer.disconnect();
      mediaQuery.removeEventListener('change', updateTheme);
    };
  }, []);

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

        // 通知父组件执行已开始
        if (onExecutionComplete) {
          onExecutionComplete({
            task_uuid: uuid,
            status: 'pending',
          });
        }

        // 自动查询结果
        setTimeout(() => queryResult(uuid), 2000);
      } else {
        setError(`请求失败: ${response.data.msg || '未知错误'}`);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '未知错误';
      setError(`网络错误: ${message}`);
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
        setError(`查询失败: ${response.data.msg || '未知错误'}`);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '未知错误';
      setError(`查询错误: ${message}`);
    } finally {
      setIsQuerying(false);
    }
  };

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
                theme={isDarkMode ? 'vs-dark' : 'vs'}
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
                    theme={isDarkMode ? 'vs-dark' : 'vs'}
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
                        theme={isDarkMode ? 'vs-dark' : 'vs'}
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
                          theme={isDarkMode ? 'vs-dark' : 'vs'}
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
