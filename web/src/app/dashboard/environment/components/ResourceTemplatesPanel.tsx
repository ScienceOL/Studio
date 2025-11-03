import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import type { ResourceTemplate } from '@/types/material';
import { Box, Play, Zap } from 'lucide-react';
import { useMemo, useState } from 'react';

interface ResourceTemplatesPanelProps {
  resourceTemplates: ResourceTemplate[];
  isLoading: boolean;
  onSelectResource: (template: ResourceTemplate) => void;
}

export default function ResourceTemplatesPanel({
  resourceTemplates,
  isLoading,
  onSelectResource,
}: ResourceTemplatesPanelProps) {
  const [query, setQuery] = useState('');

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return resourceTemplates;
    return resourceTemplates.filter(
      (template) =>
        template.name.toLowerCase().includes(q) ||
        (template.description || '').toLowerCase().includes(q)
    );
  }, [query, resourceTemplates]);

  return (
    <Card className="border-neutral-200 dark:border-neutral-800">
      <CardHeader className="space-y-2">
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <CardTitle className="flex items-center gap-2 text-neutral-900 dark:text-neutral-100">
              <Box className="h-5 w-5" />
              Resource 资源列表
            </CardTitle>
            <CardDescription className="text-neutral-600 dark:text-neutral-400 mt-1">
              实验室中已注册的资源模板
            </CardDescription>
          </div>
          {resourceTemplates.length > 0 && (
            <div className="text-sm text-neutral-500 dark:text-neutral-400 bg-neutral-100 dark:bg-neutral-800 px-3 py-1 rounded-full">
              {filtered.length} / {resourceTemplates.length}
            </div>
          )}
        </div>

        {/* 搜索框 */}
        {resourceTemplates.length > 0 && (
          <div className="pt-2">
            <input
              type="text"
              placeholder="搜索资源名称或描述..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="w-full px-4 py-2.5 text-sm border border-neutral-300 dark:border-neutral-700 rounded-lg bg-white dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 placeholder-neutral-400 dark:placeholder-neutral-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
            />
          </div>
        )}
      </CardHeader>

      <CardContent className="pt-6">
        {isLoading ? (
          <div className="space-y-3 p-4">
            {[1, 2, 3].map((i) => (
              <Skeleton
                key={i}
                className="h-12 w-full bg-neutral-200 dark:bg-neutral-700"
              />
            ))}
          </div>
        ) : resourceTemplates.length === 0 ? (
          <div className="text-center py-12 text-neutral-500 dark:text-neutral-400">
            暂无资源模板
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-12 text-neutral-500 dark:text-neutral-400">
            <Box className="h-12 w-12 mx-auto mb-3 opacity-30" />
            <p>未找到匹配的资源</p>
            <p className="text-sm mt-1 opacity-70">试试其他关键词</p>
          </div>
        ) : (
          <div className="space-y-4">
            {filtered.map((template) => (
              <Card
                key={template.uuid}
                className="border-neutral-200 dark:border-neutral-800"
              >
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex items-start gap-3">
                      {template.icon ? (
                        <div className="text-2xl mt-1">{template.icon}</div>
                      ) : (
                        <Box className="h-6 w-6 text-neutral-400 mt-1" />
                      )}
                      <div>
                        <CardTitle className="text-lg text-neutral-900 dark:text-neutral-100">
                          {template.name}
                        </CardTitle>
                        <CardDescription className="text-sm text-neutral-600 dark:text-neutral-400 mt-1">
                          {template.description || '无描述'}
                        </CardDescription>
                      </div>
                    </div>
                    <div className="flex flex-col gap-2 items-end">
                      <Badge className="bg-indigo-100 text-indigo-900 dark:bg-indigo-900 dark:text-indigo-100">
                        {template.uuid.slice(0, 8)}
                      </Badge>
                      <Badge
                        variant="outline"
                        className="border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300"
                      >
                        实例数: {template.material_count || 0}
                      </Badge>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="text-neutral-600 dark:text-neutral-400">
                        资源类型:
                      </span>
                      <span className="ml-2 text-neutral-900 dark:text-neutral-100">
                        {template.resource_type || '-'}
                      </span>
                    </div>
                    <div>
                      <span className="text-neutral-600 dark:text-neutral-400">
                        语言:
                      </span>
                      <span className="ml-2 text-neutral-900 dark:text-neutral-100">
                        {template.language || '-'}
                      </span>
                    </div>
                    <div>
                      <span className="text-neutral-600 dark:text-neutral-400">
                        模块:
                      </span>
                      <span className="ml-2 text-neutral-900 dark:text-neutral-100">
                        {template.module || '-'}
                      </span>
                    </div>
                    <div>
                      <span className="text-neutral-600 dark:text-neutral-400">
                        版本:
                      </span>
                      <span className="ml-2 text-neutral-900 dark:text-neutral-100">
                        {template.version || '-'}
                      </span>
                    </div>
                  </div>

                  {template.tags && template.tags.length > 0 && (
                    <div className="mt-3 flex flex-wrap gap-2">
                      {template.tags.map((tag, idx) => (
                        <Badge
                          key={idx}
                          variant="outline"
                          className="border-green-300 dark:border-green-600 text-green-700 dark:text-green-300"
                        >
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  )}

                  {template.actions && template.actions.length > 0 && (
                    <button
                      onClick={() => onSelectResource(template)}
                      className="w-full hover:cursor-pointer mt-4 py-4 border-t border-neutral-200 dark:border-neutral-800 text-left transition-all hover:bg-neutral-50 dark:hover:bg-neutral-800/50 rounded-lg -mx-2 px-2 group"
                    >
                      <div className="flex items-center justify-between mb-3">
                        <div className="font-medium text-sm text-neutral-700 dark:text-neutral-300 flex items-center gap-2">
                          <Play className="h-4 w-4 text-indigo-500 group-hover:text-indigo-600 transition-colors" />
                          支持的动作 ({template.actions.length})
                        </div>
                        <Zap className="h-4 w-4 text-neutral-400 group-hover:text-indigo-500 transition-colors" />
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {template.actions.slice(0, 6).map((action, idx) => (
                          <Badge
                            key={idx}
                            className="bg-blue-100 text-blue-900 dark:bg-blue-900 dark:text-blue-100 group-hover:bg-blue-200 dark:group-hover:bg-blue-800 transition-colors"
                          >
                            {action.name}
                          </Badge>
                        ))}
                        {template.actions.length > 6 && (
                          <Badge
                            variant="outline"
                            className="border-neutral-300 dark:border-neutral-600 text-neutral-600 dark:text-neutral-400"
                          >
                            +{template.actions.length - 6} 更多
                          </Badge>
                        )}
                      </div>
                    </button>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
