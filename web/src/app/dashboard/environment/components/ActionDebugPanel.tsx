/**
 * 📄 ActionDebug 组件
 *
 * 职责：手动执行设备动作的调试页面
 *
 * 功能：
 * 1. 输入JSON格式的动作参数
 * 2. 发送动作请求
 * 3. 查看执行结果
 */

import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { config } from '@/configs';
import apiClient from '@/service/http/client';
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

interface ActionRequest {
  lab_uuid: string;
  device_id: string;
  action: string;
  action_type: string;
  param?: Record<string, unknown>;
}

interface ActionResult {
  job_id: string;
  task_id: string;
  device_id: string;
  action_name: string;
  status: string;
  feedback_data?: Record<string, unknown>;
  return_info?: Record<string, unknown>;
}

interface ActionDebugProps {
  labUuid?: string;
}

export default function ActionDebugPanel({ labUuid }: ActionDebugProps) {
  const showToast = (
    title: string,
    description: string,
    variant: 'default' | 'destructive' = 'default'
  ) => {
    // 简单的通知实现
    console.log(`[${variant}] ${title}: ${description}`);
  };

  // JSON输入和结果
  const [jsonInput, setJsonInput] = useState<string>(
    '{\n  "device_id": "",\n  "action": "",\n  "action_type": "",\n  "param": {}\n}'
  );
  const [result, setResult] = useState<ActionResult | null>(null);
  const [taskUuid, setTaskUuid] = useState<string>('');

  // 状态
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

  const handleRunAction = async () => {
    setError('');
    setResult(null);
    setTaskUuid('');

    // 验证实验室
    if (!labUuid) {
      setError('实验室 UUID 未提供');
      return;
    }

    // 验证和解析JSON
    let requestData: ActionRequest;
    try {
      const parsed = JSON.parse(jsonInput);
      requestData = {
        lab_uuid: labUuid,
        device_id: parsed.device_id,
        action: parsed.action,
        action_type: parsed.action_type,
        param: parsed.param,
      };

      // 验证必填字段
      if (
        !requestData.device_id ||
        !requestData.action ||
        !requestData.action_type
      ) {
        setError('device_id, action 和 action_type 为必填字段');
        return;
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setError(`JSON 格式错误: ${message}`);
      return;
    }

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
        showToast('任务已创建', `任务 UUID: ${uuid}`);

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
        setResult(response.data.data);
        showToast(
          '查询成功',
          `状态: ${response.data.data?.status || 'unknown'}`
        );
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

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    showToast('已复制', '内容已复制到剪贴板');
  };

  const formatJson = () => {
    try {
      const parsed = JSON.parse(jsonInput);
      setJsonInput(JSON.stringify(parsed, null, 2));
      setError('');
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setError(`格式化失败: ${message}`);
    }
  };

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 左侧：输入区域 */}
        <Card>
          <CardHeader>
            <CardTitle>动作参数</CardTitle>
            <CardDescription>输入 JSON 格式的动作参数</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* JSON 输入 */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>JSON 参数</Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={formatJson}
                  className="text-xs h-7"
                >
                  格式化
                </Button>
              </div>
              <div className="border rounded-lg overflow-hidden dark:border-neutral-700">
                <Editor
                  height="400px"
                  defaultLanguage="json"
                  value={jsonInput}
                  onChange={(value) => setJsonInput(value || '{}')}
                  theme={isDarkMode ? 'vs-dark' : 'vs'}
                  options={{
                    minimap: { enabled: false },
                    fontSize: 13,
                    lineNumbers: 'on',
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                    tabSize: 2,
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
                <AlertCircle className="h-4 w-4 text-red-600 dark:text-red-400 mt-0.5" />
                <p className="text-sm text-red-600 dark:text-red-400">
                  {error}
                </p>
              </div>
            )}

            {/* 操作按钮 */}
            <div className="flex gap-3">
              <Button
                onClick={handleRunAction}
                disabled={isLoading || !labUuid}
                className="flex-1"
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
            </div>

            {/* Task UUID */}
            {taskUuid && (
              <div className="space-y-2 p-3 bg-neutral-100 dark:bg-neutral-800 rounded-lg">
                <div className="flex items-center justify-between">
                  <Label className="text-xs">任务 UUID</Label>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(taskUuid)}
                  >
                    <Copy className="h-3 w-3" />
                  </Button>
                </div>
                <p className="font-mono text-xs break-all">{taskUuid}</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* 右侧：结果区域 */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>执行结果</CardTitle>
                <CardDescription>查看动作执行的返回数据</CardDescription>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => queryResult()}
                disabled={isQuerying || !taskUuid}
              >
                {isQuerying ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <RefreshCw className="h-4 w-4" />
                )}
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {result ? (
              <div className="space-y-4">
                {/* 状态 */}
                <div className="flex items-center gap-2">
                  {result.status === 'success' ? (
                    <CheckCircle2 className="h-5 w-5 text-green-500" />
                  ) : result.status === 'fail' || result.status === 'failed' ? (
                    <XCircle className="h-5 w-5 text-red-500" />
                  ) : (
                    <Loader2 className="h-5 w-5 text-yellow-500 animate-spin" />
                  )}
                  <span className="text-lg font-semibold capitalize">
                    {result.status}
                  </span>
                </div>

                {/* 详细信息 */}
                <div className="space-y-2">
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div className="text-neutral-600 dark:text-neutral-400">
                      Job ID:
                    </div>
                    <div className="font-mono text-xs break-all">
                      {result.job_id}
                    </div>

                    <div className="text-neutral-600 dark:text-neutral-400">
                      Device ID:
                    </div>
                    <div className="font-mono text-xs">{result.device_id}</div>

                    <div className="text-neutral-600 dark:text-neutral-400">
                      Action:
                    </div>
                    <div className="font-mono text-xs">
                      {result.action_name}
                    </div>
                  </div>
                </div>

                {/* Feedback Data */}
                {result.feedback_data && (
                  <div className="space-y-2">
                    <Label className="text-sm">Feedback Data</Label>
                    <div className="border rounded-lg overflow-hidden dark:border-neutral-700">
                      <Editor
                        height="200px"
                        defaultLanguage="json"
                        value={JSON.stringify(result.feedback_data, null, 2)}
                        options={{
                          readOnly: true,
                          minimap: { enabled: false },
                          fontSize: 12,
                          lineNumbers: 'on',
                          scrollBeyondLastLine: false,
                          automaticLayout: true,
                          folding: true,
                          foldingStrategy: 'indentation',
                          showFoldingControls: 'mouseover',
                          glyphMargin: true,
                        }}
                        theme={isDarkMode ? 'vs-dark' : 'vs'}
                      />
                    </div>
                  </div>
                )}

                {/* Return Info */}
                {result.return_info && (
                  <div className="space-y-2">
                    <Label className="text-sm">Return Info</Label>
                    <div className="border rounded-lg overflow-hidden dark:border-neutral-700">
                      <Editor
                        height="200px"
                        defaultLanguage="json"
                        value={JSON.stringify(result.return_info, null, 2)}
                        options={{
                          readOnly: true,
                          minimap: { enabled: false },
                          fontSize: 12,
                          lineNumbers: 'on',
                          scrollBeyondLastLine: false,
                          automaticLayout: true,
                          folding: true,
                          foldingStrategy: 'indentation',
                          showFoldingControls: 'mouseover',
                          glyphMargin: true,
                        }}
                        theme={isDarkMode ? 'vs-dark' : 'vs'}
                      />
                    </div>
                  </div>
                )}

                {/* 复制按钮 */}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    copyToClipboard(JSON.stringify(result, null, 2))
                  }
                  className="w-full"
                >
                  <Copy className="mr-2 h-4 w-4" />
                  复制完整结果
                </Button>
              </div>
            ) : (
              <div className="text-center py-12 text-neutral-500">
                执行动作后，结果将显示在这里
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 使用说明 */}
      <Card>
        <CardHeader>
          <CardTitle>使用说明</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <h3 className="font-semibold dark:text-neutral-200">
              JSON 参数格式：
            </h3>
            <div className="border rounded-lg overflow-hidden dark:border-neutral-700">
              <Editor
                height="150px"
                defaultLanguage="json"
                value={`{
  "device_id": "设备ID（必填）",
  "action": "动作名称（必填）",
  "action_type": "动作类型（必填，如：query/setter）",
  "param": {
    "key": "value"
  }
}`}
                options={{
                  readOnly: true,
                  minimap: { enabled: false },
                  fontSize: 12,
                  lineNumbers: 'on',
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  folding: true,
                  foldingStrategy: 'indentation',
                  showFoldingControls: 'mouseover',
                }}
                theme={isDarkMode ? 'vs-dark' : 'vs'}
              />
            </div>
          </div>

          <div className="space-y-2">
            <h3 className="font-semibold dark:text-neutral-200">操作流程：</h3>
            <ol className="list-decimal list-inside space-y-1 text-sm text-neutral-600 dark:text-neutral-400">
              <li>填写或粘贴 JSON 格式的动作参数</li>
              <li>点击"执行动作"按钮</li>
              <li>等待执行完成，查看右侧结果</li>
              <li>可以点击"刷新"按钮手动查询最新结果</li>
            </ol>
          </div>

          <div className="p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
            <p className="text-sm text-blue-800 dark:text-blue-300">
              <strong>提示：</strong>
              此页面用于开发和调试，直接向后端发送动作指令。 请确保 JSON
              参数格式正确，并且实验室终端已连接。
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
