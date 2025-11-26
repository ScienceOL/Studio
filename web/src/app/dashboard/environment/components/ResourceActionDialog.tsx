/**
 * 📄 资源动作执行对话框
 *
 * 职责：展示资源支持的动作并允许用户执行
 *
 * 功能：
 * 1. 显示资源的所有可用动作
 * 2. 选择动作后显示参数表单（基于 schema 自动生成）
 * 3. 填写参数并执行动作
 * 4. 显示执行结果
 */

import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
  LocalDialog as Dialog,
  LocalDialogContent as DialogContent,
  LocalDialogDescription as DialogDescription,
  LocalDialogHeader as DialogHeader,
  LocalDialogTitle as DialogTitle,
} from '@/components/ui/local-dialog';
import type { DeviceActionInfo, ResourceTemplate } from '@/types/material';
import Editor, { loader } from '@monaco-editor/react';
import { ChevronRight, Loader2, Play, Sparkles } from 'lucide-react';
import * as monaco from 'monaco-editor';
import { useState } from 'react';

// 配置 Monaco Editor 使用本地包
loader.config({ monaco });

interface ResourceActionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  resourceTemplate: ResourceTemplate;
  labUuid: string;
}

export default function ResourceActionDialog({
  open,
  onOpenChange,
  resourceTemplate,
}: ResourceActionDialogProps) {
  const [selectedAction, setSelectedAction] = useState<DeviceActionInfo | null>(
    null
  );
  const [paramsJson, setParamsJson] = useState<string>('');

  const [searchQuery, setSearchQuery] = useState('');

  // 过滤动作列表
  const filteredActions = (resourceTemplate.actions || []).filter((action) =>
    action.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // 重置状态
  const resetState = () => {
    setSelectedAction(null);
    setParamsJson('');
  };

  // 选择动作
  const handleSelectAction = (action: DeviceActionInfo) => {
    setSelectedAction(action);

    // 根据 goal_default 生成默认参数
    if (action.goal_default) {
      setParamsJson(JSON.stringify(action.goal_default, null, 2));
    } else if (action.schema) {
      // 从 schema 生成示例参数
      const example = generateExampleFromSchema(action.schema);
      setParamsJson(JSON.stringify(example, null, 2));
    } else {
      setParamsJson('{}');
    }
  };

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

      if (propObj.default !== undefined) {
        example[key] = propObj.default;
      } else if (type === 'number') {
        example[key] = 0;
      } else if (type === 'string') {
        example[key] = '';
      } else if (type === 'boolean') {
        example[key] = false;
      } else if (type === 'array') {
        example[key] = [];
      } else if (type === 'object') {
        example[key] = {};
      }
    });

    return example;
  };

  // 格式化 JSON
  const formatJson = () => {
    try {
      const parsed = JSON.parse(paramsJson);
      setParamsJson(JSON.stringify(parsed, null, 2));
    } catch {
      // 忽略格式化错误
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => {
        if (!open) resetState();
        onOpenChange(open);
      }}
    >
      <DialogContent className="max-w-5xl w-[95vw] h-[90vh] overflow-hidden flex flex-col p-0">
        <DialogHeader className="px-4 sm:px-6 pt-4 sm:pt-6 pb-3 sm:pb-4 border-b border-neutral-200 dark:border-neutral-800 flex-shrink-0">
          <DialogTitle className="flex items-center gap-3 text-xl">
            {resourceTemplate.icon && (
              <span className="text-3xl">{resourceTemplate.icon}</span>
            )}
            <div className="flex flex-col gap-1">
              <span className="text-neutral-900 dark:text-neutral-100">
                {resourceTemplate.name}
              </span>
              <span className="text-sm font-normal text-neutral-500 dark:text-neutral-400">
                执行动作
              </span>
            </div>
          </DialogTitle>
          <DialogDescription className="text-base mt-2">
            选择要执行的动作，修改参数后点击执行按钮
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-0 flex-1 min-h-0 overflow-hidden">
          {/* 左侧：动作列表 */}
          <div className="border-r-0 lg:border-r border-b lg:border-b-0 border-neutral-200 dark:border-neutral-800 bg-neutral-50 dark:bg-neutral-900/50 flex flex-col overflow-hidden">
            <div className="p-4 sm:p-6 flex-1 flex flex-col overflow-hidden">
              <div className="flex items-center justify-between mb-3 flex-shrink-0">
                <h3 className="font-semibold text-lg text-neutral-900 dark:text-neutral-100">
                  可用动作
                </h3>
                <span className="text-sm text-neutral-500 dark:text-neutral-400 bg-neutral-200 dark:bg-neutral-800 px-2 py-1 rounded">
                  {filteredActions.length}/
                  {resourceTemplate.actions?.length || 0}
                </span>
              </div>

              {/* 搜索框 */}
              {resourceTemplate.actions &&
                resourceTemplate.actions.length > 0 && (
                  <div className="mb-3 flex-shrink-0">
                    <input
                      type="text"
                      placeholder="搜索动作..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      className="w-full px-3 py-2 text-sm border border-neutral-300 dark:border-neutral-700 rounded-lg bg-white dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 placeholder-neutral-400 dark:placeholder-neutral-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
                    />
                  </div>
                )}

              {!resourceTemplate.actions ||
              resourceTemplate.actions.length === 0 ? (
                <div className="flex-1 flex items-center justify-center text-neutral-500 dark:text-neutral-400">
                  <div className="text-center">
                    <Sparkles className="h-12 w-12 mx-auto mb-3 opacity-30" />
                    <p>该资源暂无可用动作</p>
                  </div>
                </div>
              ) : filteredActions.length === 0 ? (
                <div className="flex-1 flex items-center justify-center text-neutral-500 dark:text-neutral-400">
                  <div className="text-center">
                    <Sparkles className="h-12 w-12 mx-auto mb-3 opacity-30" />
                    <p>未找到匹配的动作</p>
                    <p className="text-sm mt-1 opacity-70">试试其他关键词</p>
                  </div>
                </div>
              ) : (
                <div className="flex-1 overflow-y-auto space-y-2 pr-2 custom-scrollbar">
                  {filteredActions.map((action, idx) => (
                    <button
                      key={idx}
                      onClick={() => handleSelectAction(action)}
                      className={`w-full text-left p-4 rounded-lg border-2 transition-all ${
                        selectedAction?.name === action.name
                          ? 'border-indigo-500 bg-white dark:bg-indigo-950 shadow-sm'
                          : 'border-transparent bg-white dark:bg-neutral-800 hover:border-neutral-300 dark:hover:border-neutral-700 hover:shadow-sm'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1 min-w-0">
                          <div className="font-medium text-neutral-900 dark:text-neutral-100 flex items-center gap-2 mb-1">
                            <Sparkles
                              className={`h-4 w-4 flex-shrink-0 ${
                                selectedAction?.name === action.name
                                  ? 'text-indigo-500'
                                  : 'text-neutral-400'
                              }`}
                            />
                            <span className="truncate">{action.name}</span>
                          </div>
                          <div className="text-xs truncate text-neutral-500 dark:text-neutral-400">
                            {action.type || 'action'}
                          </div>
                        </div>
                        <ChevronRight
                          className={`h-5 w-5 flex-shrink-0 ml-2 transition-transform ${
                            selectedAction?.name === action.name
                              ? 'text-indigo-500 transform rotate-90'
                              : 'text-neutral-400'
                          }`}
                        />
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* 右侧：参数配置和执行 */}
          <div className="bg-white dark:bg-neutral-950 flex flex-col overflow-hidden">
            <div className="p-4 sm:p-6 flex-1 flex flex-col overflow-hidden">
              {!selectedAction ? (
                <div className="flex-1 flex items-center justify-center text-neutral-500 dark:text-neutral-400">
                  <div className="text-center">
                    <Play className="h-12 w-12 mx-auto mb-3 opacity-30" />
                    <p className="text-base">请先在左侧选择一个动作</p>
                    <p className="text-sm mt-1 opacity-70">
                      点击动作卡片查看参数配置
                    </p>
                  </div>
                </div>
              ) : (
                <>
                  {/* 动作信息 - 固定在顶部 */}
                  <div className="mb-4 pb-4 border-b border-neutral-200 dark:border-neutral-800 flex-shrink-0">
                    <h4 className="font-semibold text-lg text-neutral-900 dark:text-neutral-100 mb-2 flex items-center gap-2">
                      <Sparkles className="h-5 w-5 text-indigo-500" />
                      {selectedAction.name}
                    </h4>
                    <div className="flex items-center gap-4 text-sm text-neutral-600 dark:text-neutral-400">
                      <span className="bg-neutral-100 max-w-full truncate dark:bg-neutral-800 px-2 py-1 rounded">
                        类型: {selectedAction.type || 'action'}
                      </span>
                    </div>
                  </div>

                  {/* 参数编辑 - 可滚动区域 */}
                  <div className="flex-1 flex flex-col mb-4 min-h-0">
                    <div className="flex items-center justify-between mb-2 flex-shrink-0">
                      <Label
                        htmlFor="params"
                        className="text-base font-medium text-neutral-900 dark:text-neutral-100"
                      >
                        参数配置
                      </Label>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={formatJson}
                        className="h-8 text-xs"
                      >
                        <span className="mr-1">✨</span>
                        格式化
                      </Button>
                    </div>
                    <div className="flex-1 min-h-0 border-2 border-neutral-200 dark:border-neutral-800 rounded-lg overflow-hidden focus-within:border-indigo-500 dark:focus-within:border-indigo-400">
                      <Editor
                        height="100%"
                        defaultLanguage="json"
                        value={JSON.stringify(selectedAction.schema, null, 2)}
                        theme={'vs-dark'}
                        options={{
                          readOnly: true,
                          minimap: { enabled: false },
                          fontSize: 13,
                          lineNumbers: 'on',
                          scrollBeyondLastLine: false,
                          wordWrap: 'on',
                          wrappingStrategy: 'advanced',
                          automaticLayout: true,
                          tabSize: 2,
                          folding: true,
                          foldingStrategy: 'indentation',
                          showFoldingControls: 'mouseover',
                          renderLineHighlight: 'none',
                          glyphMargin: true,
                        }}
                        loading={
                          <div className="flex items-center justify-center h-full">
                            <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
                          </div>
                        }
                      />
                    </div>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
