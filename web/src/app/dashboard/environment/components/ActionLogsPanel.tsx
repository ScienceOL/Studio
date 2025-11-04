import { Button } from '@/components/ui/button';
/**
 * 📄 ActionLogsPanel 组件
 *
 * 职责：在 Environment 详情页中显示动作执行日志
 *
 * 功能：
 * 1. 显示当前 lab 的所有动作执行历史记录
 * 2. 过滤和搜索日志
 * 3. 查看详细执行过程
 * 4. 导出和删除日志
 */

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { useActionLogStore, type ActionLogEntry } from '@/store/actionLogStore';
import Editor, { loader } from '@monaco-editor/react';
import {
  CheckCircle2,
  Clock,
  Download,
  Loader2,
  Trash2,
  XCircle,
} from 'lucide-react';
import * as monaco from 'monaco-editor';
import { useEffect, useState } from 'react';

loader.config({ monaco });

interface ActionLogsPanelProps {
  labUuid: string;
}

export default function ActionLogsPanel({ labUuid }: ActionLogsPanelProps) {
  const { logs, deleteLog } = useActionLogStore();
  // 状态筛选：all | success | failed | running | pending
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [selectedLog, setSelectedLog] = useState<ActionLogEntry | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);

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

  // 合并同一 taskUuid 的日志为一条，按 statusHistory 展示阶段
  const mergedLogs = (() => {
    // 只取当前 labUuid
    const labLogs = logs.filter((log) => log.labUuid === labUuid);
    // 按 taskUuid 分组
    const map = new Map<string, ActionLogEntry>();
    for (const log of labLogs) {
      if (!map.has(log.taskUuid)) {
        map.set(log.taskUuid, { ...log });
      } else {
        // 合并 statusHistory
        const exist = map.get(log.taskUuid)!;
        exist.statusHistory = [
          ...exist.statusHistory,
          ...log.statusHistory.filter(
            (h) =>
              !exist.statusHistory.some(
                (eh) => eh.status === h.status && eh.timestamp === h.timestamp
              )
          ),
        ];
        // 取最新状态
        if (log.endTime && (!exist.endTime || log.endTime > exist.endTime)) {
          exist.endTime = log.endTime;
        }
        if (log.status && log.status !== exist.status) {
          exist.status = log.status;
        }
        if (log.finalResult) {
          exist.finalResult = log.finalResult;
        }
        if (log.error) {
          exist.error = log.error;
        }
      }
    }
    // 排序 statusHistory 按时间
    for (const entry of map.values()) {
      entry.statusHistory = entry.statusHistory.sort(
        (a, b) =>
          new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
      );
    }
    // 状态过滤
    let arr = Array.from(map.values());
    if (statusFilter !== 'all') {
      arr = arr.filter((log) => log.status === statusFilter);
    }
    // 按开始时间倒序
    arr.sort(
      (a, b) =>
        new Date(b.startTime).getTime() - new Date(a.startTime).getTime()
    );
    return arr;
  })();

  // 格式化持续时间（精确到毫秒）
  const formatDuration = (ms?: number): string => {
    if (typeof ms !== 'number' || isNaN(ms)) return '-';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`;
    return `${(ms / 60000).toFixed(2)}min`;
  };

  // 格式化时间（精确到毫秒）
  const formatTime = (iso: string): string => {
    const d = new Date(iso);
    if (isNaN(d.getTime())) return '-';
    const pad = (n: number, l = 2) => n.toString().padStart(l, '0');
    return (
      `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
      `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}.${pad(
        d.getMilliseconds(),
        3
      )}`
    );
  };

  // 获取状态图标
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      case 'failed':
      case 'fail':
        return <XCircle className="h-4 w-4 text-red-500" />;
      case 'running':
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />;
      default:
        return <Clock className="h-4 w-4 text-yellow-500" />;
    }
  };

  // 导出日志为 JSON
  const exportLogs = () => {
    const dataStr = JSON.stringify(mergedLogs, null, 2);
    const dataBlob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `action-logs-${labUuid}-${new Date().toISOString()}.json`;
    link.click();
    URL.revokeObjectURL(url);
  };

  const labLogs = logs.filter((l) => l.labUuid === labUuid);

  return (
    <div className="space-y-6">
      {/* 标题栏 */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-neutral-900 dark:text-neutral-100">
            动作执行日志
          </h2>
          <p className="text-neutral-500 dark:text-neutral-400 mt-1">
            查看当前环境的所有设备动作执行历史记录
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={exportLogs}
            disabled={mergedLogs.length === 0}
            size="sm"
          >
            <Download className="mr-2 h-4 w-4" />
            导出日志
          </Button>
        </div>
      </div>

      {/* 状态筛选按钮组 */}
      <div className="flex gap-4 my-2">
        <Button
          variant={statusFilter === 'all' ? 'default' : 'outline'}
          onClick={() => setStatusFilter('all')}
        >
          总计 <span className="ml-1">{labLogs.length}</span>
        </Button>
        <Button
          variant={statusFilter === 'success' ? 'default' : 'outline'}
          onClick={() => setStatusFilter('success')}
        >
          成功{' '}
          <span className="ml-1">
            {labLogs.filter((l) => l.status === 'success').length}
          </span>
        </Button>
        <Button
          variant={statusFilter === 'failed' ? 'default' : 'outline'}
          onClick={() => setStatusFilter('failed')}
        >
          失败{' '}
          <span className="ml-1">
            {
              labLogs.filter(
                (l) => l.status === 'failed' || l.status === 'fail'
              ).length
            }
          </span>
        </Button>
        <Button
          variant={statusFilter === 'running' ? 'default' : 'outline'}
          onClick={() => setStatusFilter('running')}
        >
          执行中{' '}
          <span className="ml-1">
            {labLogs.filter((l) => l.status === 'running').length}
          </span>
        </Button>
      </div>

      {/* 统计信息 */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-neutral-50 dark:bg-neutral-900 p-4 rounded-lg border border-neutral-200 dark:border-neutral-800">
          <p className="text-sm text-neutral-500 dark:text-neutral-400">总计</p>
          <p className="text-2xl font-bold text-neutral-900 dark:text-neutral-100">
            {labLogs.length}
          </p>
        </div>
        <div className="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg border border-green-200 dark:border-green-800/50">
          <p className="text-sm text-green-600 dark:text-green-400">成功</p>
          <p className="text-2xl font-bold text-green-600 dark:text-green-400">
            {labLogs.filter((l) => l.status === 'success').length}
          </p>
        </div>
        <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-lg border border-red-200 dark:border-red-800/50">
          <p className="text-sm text-red-600 dark:text-red-400">失败</p>
          <p className="text-2xl font-bold text-red-600 dark:text-red-400">
            {
              labLogs.filter(
                (l) => l.status === 'failed' || l.status === 'fail'
              ).length
            }
          </p>
        </div>
        <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg border border-blue-200 dark:border-blue-800/50">
          <p className="text-sm text-blue-600 dark:text-blue-400">执行中</p>
          <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
            {labLogs.filter((l) => l.status === 'running').length}
          </p>
        </div>
      </div>

      {/* 日志时间条列表 */}
      <div className="border rounded-lg overflow-x-hidden border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-950">
        {mergedLogs.length === 0 ? (
          <div className="py-12 text-center text-neutral-500 dark:text-neutral-400">
            {labLogs.length === 0 ? '暂无日志记录' : '没有符合条件的日志'}
          </div>
        ) : (
          <div className="divide-y divide-neutral-200 dark:divide-neutral-800">
            {mergedLogs.map((log, logIndex) => {
              // 计算每个阶段的持续时间，并合并连续相同状态
              const rawStages = log.statusHistory.map((h, idx, arr) => {
                const start = new Date(h.timestamp).getTime();
                const end =
                  idx < arr.length - 1
                    ? new Date(arr[idx + 1].timestamp).getTime()
                    : log.endTime
                    ? new Date(log.endTime).getTime()
                    : Date.now();
                return {
                  status: h.status,
                  start,
                  end,
                  duration: end - start,
                  timestamp: h.timestamp,
                };
              });

              // 合并连续相同状态的阶段
              const stages = rawStages.reduce((acc, stage) => {
                if (acc.length === 0) {
                  return [stage];
                }
                const last = acc[acc.length - 1];
                // 如果当前阶段状态与上一个相同，合并它们
                if (last.status === stage.status) {
                  last.end = stage.end;
                  last.duration = last.end - last.start;
                  return acc;
                }
                // 否则添加新阶段
                return [...acc, stage];
              }, [] as typeof rawStages);

              // 总持续时间
              const totalDuration =
                stages.length > 0
                  ? stages[stages.length - 1].end - stages[0].start
                  : 0;
              return (
                <div
                  key={log.taskUuid}
                  className="flex flex-col md:flex-row items-stretch md:items-center gap-2 px-4 py-4 hover:bg-neutral-50 dark:hover:bg-neutral-900/50 cursor-pointer transition-colors"
                  onClick={() => {
                    setSelectedLog(log);
                    setDetailOpen(true);
                  }}
                >
                  {/* 设备/动作信息 */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-neutral-900 dark:text-neutral-100">
                        {log.deviceName || log.deviceId}
                      </span>
                      <span className="text-xs text-neutral-500 dark:text-neutral-400">
                        {log.deviceId}
                      </span>
                    </div>
                    <div className="text-sm text-neutral-700 dark:text-neutral-300 truncate">
                      {log.actionName}
                    </div>
                  </div>
                  {/* 时间条 */}
                  <div className="flex-1 flex flex-col gap-1 min-w-0">
                    <div className="flex items-center w-full overflow-visible rounded h-4 relative">
                      {stages.map((stage, idx) => {
                        let color = '';
                        let statusLabel = '';
                        if (stage.status === 'success') {
                          color = 'bg-green-500';
                          statusLabel = '成功';
                        } else if (
                          stage.status === 'failed' ||
                          stage.status === 'fail'
                        ) {
                          color = 'bg-red-500';
                          statusLabel = '失败';
                        } else if (stage.status === 'running') {
                          color = 'bg-yellow-400';
                          statusLabel = '执行中';
                        } else {
                          color = 'bg-gray-400';
                          statusLabel = stage.status;
                        }
                        const percent =
                          totalDuration > 0
                            ? (stage.duration / totalDuration) * 100
                            : 0;
                        // 前两条记录的 tooltip 显示在下方，其他显示在上方
                        const showBelow = logIndex < 2;
                        return (
                          <div
                            key={idx}
                            className={`h-4 ${color} relative group transition-all duration-200 hover:brightness-110`}
                            style={{ width: `${percent}%` }}
                          >
                            {/* Tooltip */}
                            <div
                              className={`absolute left-1/2 -translate-x-1/2 ${
                                showBelow ? 'top-full mt-2' : 'bottom-full mb-2'
                              } opacity-0 group-hover:opacity-100 transition-all duration-200 pointer-events-none z-50 whitespace-nowrap`}
                            >
                              <div className="bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-900 text-xs px-3 py-2 rounded-lg shadow-lg">
                                <div className="font-semibold">
                                  {statusLabel}
                                </div>
                                <div className="text-neutral-300 dark:text-neutral-600">
                                  {formatTime(stage.timestamp)}
                                </div>
                                <div className="text-neutral-400 dark:text-neutral-500">
                                  持续: {formatDuration(stage.duration)}
                                </div>
                              </div>
                              {/* Arrow */}
                              <div
                                className={`absolute left-1/2 -translate-x-1/2 ${
                                  showBelow
                                    ? 'bottom-full -mb-1'
                                    : 'top-full -mt-1'
                                }`}
                              >
                                <div
                                  className={`border-4 border-transparent ${
                                    showBelow
                                      ? 'border-b-neutral-900 dark:border-b-neutral-100'
                                      : 'border-t-neutral-900 dark:border-t-neutral-100'
                                  }`}
                                ></div>
                              </div>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                    <div className="flex justify-between text-xs text-neutral-500 dark:text-neutral-400">
                      <span>{formatTime(log.startTime)}</span>
                      <span>{log.endTime ? formatTime(log.endTime) : '-'}</span>
                    </div>
                  </div>
                  {/* 总持续时间/状态/操作 */}
                  <div className="flex flex-col items-end gap-2 min-w-[120px]">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(log.status)}
                      <span className="capitalize text-neutral-900 dark:text-neutral-100">
                        {log.status}
                      </span>
                    </div>
                    <div className="text-xs text-neutral-700 dark:text-neutral-300">
                      {formatDuration(totalDuration)}
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={async (e) => {
                        e.stopPropagation();
                        if (confirm('确定要删除这条日志吗？')) {
                          await deleteLog(log.taskUuid);
                        }
                      }}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* 详情对话框 */}
      <Dialog open={detailOpen} onOpenChange={setDetailOpen} size="5xl">
        <DialogContent className="max-w-5xl m-4 max-h-[90vh] custom-scrollbar overflow-auto">
          <DialogHeader>
            <DialogTitle>日志详情</DialogTitle>
            <DialogDescription>
              任务 UUID: {selectedLog?.taskUuid}
            </DialogDescription>
          </DialogHeader>

          {selectedLog && (
            <div className="space-y-4 overflow-y-auto">
              {/* 基本信息 */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-neutral-700 dark:text-neutral-300">
                    设备名称
                  </Label>
                  <p className="text-sm mt-1 text-neutral-900 dark:text-neutral-100">
                    {selectedLog.deviceName || '-'}
                  </p>
                </div>
                <div>
                  <Label className="text-neutral-700 dark:text-neutral-300">
                    设备 ID
                  </Label>
                  <p className="text-sm mt-1 font-mono text-neutral-900 dark:text-neutral-100">
                    {selectedLog.deviceId}
                  </p>
                </div>
                <div>
                  <Label className="text-neutral-700 dark:text-neutral-300">
                    动作名称
                  </Label>
                  <p className="text-sm mt-1 text-neutral-900 dark:text-neutral-100">
                    {selectedLog.actionName}
                  </p>
                </div>
                <div>
                  <Label className="text-neutral-700 dark:text-neutral-300">
                    状态
                  </Label>
                  <div className="flex items-center gap-2 mt-1">
                    {getStatusIcon(selectedLog.status)}
                    <span className="capitalize text-neutral-900 dark:text-neutral-100">
                      {selectedLog.status}
                    </span>
                  </div>
                </div>
                <div>
                  <Label className="text-neutral-700 dark:text-neutral-300">
                    开始时间
                  </Label>
                  <p className="text-sm mt-1 text-neutral-900 dark:text-neutral-100">
                    {formatTime(selectedLog.startTime)}
                  </p>
                </div>
                <div>
                  <Label className="text-neutral-700 dark:text-neutral-300">
                    结束时间
                  </Label>
                  <p className="text-sm mt-1 text-neutral-900 dark:text-neutral-100">
                    {selectedLog.endTime
                      ? formatTime(selectedLog.endTime)
                      : '-'}
                  </p>
                </div>
              </div>

              {/* 状态历史 */}
              <div>
                <Label className="text-neutral-700 dark:text-neutral-300">
                  状态变化历史
                </Label>
                <div className="mt-2 space-y-2">
                  {selectedLog.statusHistory.map((history, index) => (
                    <div
                      key={index}
                      className="p-3 bg-neutral-50 dark:bg-neutral-900 rounded-lg border border-neutral-200 dark:border-neutral-800 flex items-start justify-between"
                    >
                      <div className="flex items-center gap-2">
                        {getStatusIcon(history.status)}
                        <span className="font-medium capitalize text-neutral-900 dark:text-neutral-100">
                          {history.status}
                        </span>
                      </div>
                      <span className="text-sm text-neutral-500 dark:text-neutral-400">
                        {formatTime(history.timestamp)}
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              {/* 最终结果 */}
              {selectedLog.finalResult && (
                <div>
                  <Label className="text-neutral-700 dark:text-neutral-300">
                    返回结果
                  </Label>
                  <div className="mt-2 border rounded overflow-hidden border-neutral-200 dark:border-neutral-800">
                    <Editor
                      height="300px"
                      defaultLanguage="json"
                      value={JSON.stringify(selectedLog.finalResult, null, 2)}
                      options={{
                        readOnly: true,
                        minimap: { enabled: false },
                        fontSize: 12,
                        lineNumbers: 'on',
                        scrollBeyondLastLine: false,
                        automaticLayout: true,
                        wordWrap: 'on',
                      }}
                      theme={isDarkMode ? 'vs-dark' : 'vs'}
                    />
                  </div>
                </div>
              )}

              {/* 错误信息 */}
              {selectedLog.error && (
                <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
                  <Label className="text-red-600 dark:text-red-400">
                    错误信息
                  </Label>
                  <p className="text-sm mt-1 text-red-600 dark:text-red-400">
                    {selectedLog.error}
                  </p>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
