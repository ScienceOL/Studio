/**
 * 📄 ActionPanel 组件
 *
 * 职责：在实验室详情页面中执行设备动作
 *
 * 功能：
 * 1. 左侧：选择设备实例（Material）
 * 2. 中间：选择可用动作（点击打开 Dialog）
 * 3. 右侧：查看执行历史记录
 */

import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import type {
  DeviceActionInfo,
  Material,
  ResourceTemplate,
} from '@/types/material';
import {
  Box,
  ChevronRight,
  Clock,
  History,
  Loader2,
  Search,
  Zap,
} from 'lucide-react';
import { useState } from 'react';
import ActionRunnerDialog from './ActionRunnerDialog';

// 执行历史记录接口
interface ExecutionHistory {
  id: string;
  timestamp: number;
  device_id: string;
  device_name: string;
  action_name: string;
  action_type: string;
  task_uuid?: string;
  status?: 'success' | 'fail' | 'pending';
  params?: Record<string, unknown>;
}

interface ActionPanelProps {
  labUuid: string;
  materials: Material[];
  resourceTemplates: ResourceTemplate[];
  isLoadingMaterials?: boolean;
  isLoadingResources?: boolean;
}

export default function ActionPanel({
  labUuid,
  materials,
  resourceTemplates,
  isLoadingMaterials = false,
  isLoadingResources = false,
}: ActionPanelProps) {
  // 选择状态
  const [selectedMaterial, setSelectedMaterial] = useState<Material | null>(
    null
  );
  const [materialSearchQuery, setMaterialSearchQuery] = useState('');
  const [actionSearchQuery, setActionSearchQuery] = useState('');

  // Dialog 状态
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedActionForDialog, setSelectedActionForDialog] =
    useState<DeviceActionInfo | null>(null);

  // 执行历史
  const [executionHistory, setExecutionHistory] = useState<ExecutionHistory[]>(
    []
  );
  const [selectedHistoryId, setSelectedHistoryId] = useState<string | null>(
    null
  );

  // 过滤 Materials
  const filteredMaterials = materials.filter(
    (material: Material) =>
      material.name.toLowerCase().includes(materialSearchQuery.toLowerCase()) ||
      material.type.toLowerCase().includes(materialSearchQuery.toLowerCase())
  );

  // 获取选中 Material 对应的 Resource Template
  // 使用 material.class 匹配 resourceTemplate.name
  const matchedResourceTemplate = selectedMaterial
    ? resourceTemplates.find(
        (rt: ResourceTemplate) => rt.name === selectedMaterial.class
      )
    : null;

  // 获取可用的 Actions
  const availableActions = matchedResourceTemplate?.actions || [];

  // 过滤 Actions
  const filteredActions = availableActions.filter(
    (action: DeviceActionInfo) =>
      action.name.toLowerCase().includes(actionSearchQuery.toLowerCase()) ||
      action.type.toLowerCase().includes(actionSearchQuery.toLowerCase())
  );

  // 点击动作：打开 Dialog
  const handleActionClick = (action: DeviceActionInfo) => {
    setSelectedActionForDialog(action);
    setDialogOpen(true);
  };

  // 处理执行完成，添加到历史记录
  const handleExecutionComplete = (executionData: {
    task_uuid: string;
    status: string;
    result?: unknown;
  }) => {
    if (!selectedMaterial || !selectedActionForDialog) return;

    const historyItem: ExecutionHistory = {
      id: `${Date.now()}-${Math.random()}`,
      timestamp: Date.now(),
      device_id: selectedMaterial.name,
      device_name: selectedMaterial.name,
      action_name: selectedActionForDialog.name,
      action_type: selectedActionForDialog.type,
      task_uuid: executionData.task_uuid,
      status: executionData.status as 'success' | 'fail' | 'pending',
    };

    setExecutionHistory((prev) => {
      // 如果是更新状态，查找并更新现有记录
      const existingIndex = prev.findIndex(
        (h) => h.task_uuid === executionData.task_uuid
      );
      if (existingIndex !== -1) {
        const updated = [...prev];
        updated[existingIndex] = { ...updated[existingIndex], ...historyItem };
        return updated;
      }
      // 否则添加新记录
      return [historyItem, ...prev];
    });
  };

  // 格式化时间
  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const selectedHistory = executionHistory.find(
    (h) => h.id === selectedHistoryId
  );

  return (
    <>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 左侧：设备列表 */}
        <Card className="border-neutral-200 dark:border-neutral-800">
          <CardHeader>
            <CardTitle className="text-neutral-900 dark:text-neutral-100 flex items-center gap-2">
              <Box className="h-5 w-5" />
              设备列表
            </CardTitle>
            <CardDescription className="text-neutral-600 dark:text-neutral-400">
              选择要操作的设备实例
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* 搜索框 */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-neutral-400" />
              <Input
                placeholder="搜索设备名称或类型..."
                value={materialSearchQuery}
                onChange={(e) => setMaterialSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* 设备列表 */}
            <div className="space-y-2 max-h-[600px] overflow-y-auto">
              {isLoadingMaterials ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin text-neutral-400" />
                </div>
              ) : filteredMaterials.length === 0 ? (
                <div className="text-center py-8 text-neutral-500 text-sm">
                  没有找到设备
                </div>
              ) : (
                filteredMaterials.map((material: Material) => (
                  <button
                    key={material.uuid}
                    onClick={() => {
                      setSelectedMaterial(material);
                      setActionSearchQuery('');
                    }}
                    className={`w-full text-left p-3 rounded-lg border transition-all duration-200 ${
                      selectedMaterial?.uuid === material.uuid
                        ? 'bg-indigo-50 dark:bg-indigo-900/20 border-indigo-200 dark:border-indigo-800 shadow-sm'
                        : 'bg-white dark:bg-neutral-800 border-neutral-200 dark:border-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-800/80 hover:border-neutral-300 dark:hover:border-neutral-600 hover:shadow-sm'
                    }`}
                  >
                    <div className="font-medium text-sm text-neutral-900 dark:text-neutral-100">
                      {material.name}
                    </div>
                    <div className="text-xs text-neutral-500 dark:text-neutral-400 mt-1">
                      类型: {material.type}
                    </div>
                    {material.status && (
                      <Badge
                        variant={
                          material.status === 'active' ? 'default' : 'secondary'
                        }
                        className="mt-2"
                      >
                        {material.status}
                      </Badge>
                    )}
                  </button>
                ))
              )}
            </div>
          </CardContent>
        </Card>

        {/* 中间：动作列表 */}
        <Card className="border-neutral-200 dark:border-neutral-800">
          <CardHeader>
            <CardTitle className="text-neutral-900 dark:text-neutral-100 flex items-center gap-2">
              <Zap className="h-5 w-5" />
              可用动作
            </CardTitle>
            <CardDescription className="text-neutral-600 dark:text-neutral-400">
              {selectedMaterial
                ? `点击执行设备 "${selectedMaterial.name}" 的动作`
                : '请先在左侧选择设备'}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {selectedMaterial ? (
              <>
                {/* 搜索框 */}
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-neutral-400" />
                  <Input
                    placeholder="搜索动作名称或类型..."
                    value={actionSearchQuery}
                    onChange={(e) => setActionSearchQuery(e.target.value)}
                    className="pl-9"
                  />
                </div>

                {/* 动作列表 */}
                <div className="space-y-2 max-h-[600px] overflow-y-auto">
                  {isLoadingResources ? (
                    <div className="flex items-center justify-center py-8">
                      <Loader2 className="h-6 w-6 animate-spin text-neutral-400" />
                    </div>
                  ) : !matchedResourceTemplate ? (
                    <div className="text-center py-8 text-neutral-500 text-sm">
                      未找到该设备类型对应的资源模板
                    </div>
                  ) : filteredActions.length === 0 ? (
                    <div className="text-center py-8 text-neutral-500 text-sm">
                      没有找到可用动作
                    </div>
                  ) : (
                    filteredActions.map(
                      (action: DeviceActionInfo, idx: number) => (
                        <button
                          key={idx}
                          onClick={() => handleActionClick(action)}
                          className="w-full text-left p-3 rounded-lg border truncate transition-colors bg-white dark:bg-neutral-800 border-neutral-200 dark:border-neutral-700 hover:bg-indigo-50 dark:hover:bg-indigo-900/10 hover:border-indigo-200 dark:hover:border-indigo-800 group"
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex-1">
                              <div className="font-medium text-sm text-neutral-900 dark:text-neutral-100">
                                {action.name}
                              </div>
                              <Badge
                                variant="secondary"
                                className="mt-1 text-xs"
                              >
                                <span className=" truncate w-64">
                                  {action.type}
                                </span>
                              </Badge>
                            </div>
                            <ChevronRight className="h-4 w-4 text-neutral-400 group-hover:text-indigo-600 dark:group-hover:text-indigo-400" />
                          </div>
                        </button>
                      )
                    )
                  )}
                </div>
              </>
            ) : (
              <div className="text-center py-12 text-neutral-500">
                请先在左侧选择一个设备
              </div>
            )}
          </CardContent>
        </Card>

        {/* 右侧：执行历史 */}
        <Card className="border-neutral-200 dark:border-neutral-800">
          <CardHeader>
            <CardTitle className="text-neutral-900 dark:text-neutral-100 flex items-center gap-2">
              <History className="h-5 w-5" />
              执行历史
            </CardTitle>
            <CardDescription className="text-neutral-600 dark:text-neutral-400">
              查看最近的动作执行记录
            </CardDescription>
          </CardHeader>
          <CardContent>
            {executionHistory.length === 0 ? (
              <div className="text-center py-12 text-neutral-500 text-sm">
                暂无执行历史
              </div>
            ) : (
              <div className="space-y-2 max-h-[600px] overflow-y-auto">
                {executionHistory.map((item) => (
                  <button
                    key={item.id}
                    onClick={() => setSelectedHistoryId(item.id)}
                    className={`w-full text-left p-3 rounded-lg border transition-all duration-200 ${
                      selectedHistoryId === item.id
                        ? 'bg-indigo-50 dark:bg-indigo-900/20 border-indigo-200 dark:border-indigo-800 shadow-sm'
                        : 'bg-white dark:bg-neutral-800 border-neutral-200 dark:border-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-800/80 hover:border-neutral-300 dark:hover:border-neutral-600 hover:shadow-sm'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex-1 min-w-0">
                        <div className="font-medium text-sm text-neutral-900 dark:text-neutral-100 truncate">
                          {item.action_name}
                        </div>
                        <div className="text-xs text-neutral-500 dark:text-neutral-400 mt-1">
                          设备: {item.device_name}
                        </div>
                        <div className="flex items-center gap-1 mt-1 text-xs text-neutral-400">
                          <Clock className="h-3 w-3" />
                          {formatTime(item.timestamp)}
                        </div>
                      </div>
                      <Badge
                        variant={
                          item.status === 'success'
                            ? 'default'
                            : item.status === 'fail'
                            ? 'destructive'
                            : 'secondary'
                        }
                        className="text-xs"
                      >
                        {item.status === 'success'
                          ? '成功'
                          : item.status === 'fail'
                          ? '失败'
                          : '执行中'}
                      </Badge>
                    </div>
                  </button>
                ))}
              </div>
            )}

            {/* 历史详情 */}
            {selectedHistory && (
              <div className="mt-4 p-4 bg-neutral-50 dark:bg-neutral-900/50 rounded-lg border border-neutral-200 dark:border-neutral-800 space-y-3">
                <div className="font-medium text-sm text-neutral-900 dark:text-neutral-100">
                  执行详情
                </div>
                <div className="space-y-2 text-xs">
                  <div className="flex justify-between">
                    <span className="text-neutral-500 dark:text-neutral-400">
                      动作:
                    </span>
                    <span className="text-neutral-900 dark:text-neutral-100 font-mono">
                      {selectedHistory.action_name}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-neutral-500 dark:text-neutral-400">
                      类型:
                    </span>
                    <span className="text-neutral-900 dark:text-neutral-100">
                      {selectedHistory.action_type}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-neutral-500 dark:text-neutral-400">
                      设备:
                    </span>
                    <span className="text-neutral-900 dark:text-neutral-100 font-mono">
                      {selectedHistory.device_id}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-neutral-500 dark:text-neutral-400">
                      时间:
                    </span>
                    <span className="text-neutral-900 dark:text-neutral-100">
                      {formatTime(selectedHistory.timestamp)}
                    </span>
                  </div>
                  {selectedHistory.task_uuid && (
                    <div className="pt-2 border-t border-neutral-200 dark:border-neutral-800">
                      <Label className="text-xs text-neutral-500 dark:text-neutral-400">
                        任务 UUID:
                      </Label>
                      <div className="mt-1 p-2 bg-white dark:bg-neutral-800 rounded border border-neutral-200 dark:border-neutral-700 font-mono text-xs break-all text-neutral-900 dark:text-neutral-100">
                        {selectedHistory.task_uuid}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* ActionRunnerDialog */}
      {selectedActionForDialog && selectedMaterial && (
        <ActionRunnerDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          material={selectedMaterial}
          action={selectedActionForDialog}
          labUuid={labUuid}
          onExecutionComplete={handleExecutionComplete}
        />
      )}
    </>
  );
}
