import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import type { Material, ResourceTemplate } from '@/types/material';
import { Layers, Sparkles } from 'lucide-react';

interface MaterialsPanelProps {
  materials: Material[];
  isLoading: boolean;
  resourceTemplates: ResourceTemplate[];
  onOpenResourceActions: (template: ResourceTemplate) => void;
}

export default function MaterialsPanel({
  materials,
  isLoading,
  resourceTemplates,
  onOpenResourceActions,
}: MaterialsPanelProps) {
  return (
    <Card className="border-neutral-200 dark:border-neutral-800">
      <CardHeader className="space-y-2">
        <CardTitle className="flex items-center gap-2 text-neutral-900 dark:text-neutral-100">
          <Layers className="h-5 w-5" />
          Material 物料列表
        </CardTitle>
        <CardDescription className="text-neutral-600 dark:text-neutral-400">
          实验室中的物料节点数据
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-6">
        {isLoading ? (
          <div className="space-y-3 p-4">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-12 w-full rounded bg-neutral-200 dark:bg-neutral-700"
              />
            ))}
          </div>
        ) : materials.length === 0 ? (
          <div className="text-center py-12 text-neutral-500 dark:text-neutral-400">
            暂无物料数据
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                  ID
                </TableHead>
                <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                  名称
                </TableHead>
                <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                  类型
                </TableHead>
                <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                  Class
                </TableHead>
                <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                  状态
                </TableHead>
                <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                  父节点
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {materials.map((material) => (
                <TableRow
                  key={material.uuid}
                  className="border-b border-neutral-200 dark:border-neutral-800"
                >
                  <TableCell className="py-4">
                    <Badge
                      variant="outline"
                      className="border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300 font-mono"
                    >
                      {material.id || material.name || 'N/A'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-neutral-900 dark:text-neutral-100 py-4 font-medium">
                    {material.name || '-'}
                  </TableCell>
                  <TableCell className="py-4">
                    <Badge className="bg-indigo-100 text-indigo-900 dark:bg-indigo-900 dark:text-indigo-100">
                      {material.type || 'Unknown'}
                    </Badge>
                  </TableCell>
                  <TableCell className="py-4">
                    {material.class ? (
                      <button
                        onClick={() => {
                          const template = resourceTemplates.find(
                            (rt) => rt.name === material.class
                          );
                          if (template) {
                            onOpenResourceActions(template);
                          }
                        }}
                        className="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-green-100 text-green-900 dark:bg-green-900 dark:text-green-100 hover:bg-green-200 dark:hover:bg-green-800 transition-colors font-mono text-xs cursor-pointer"
                      >
                        <Sparkles className="h-3 w-3" />
                        {material.class}
                      </button>
                    ) : (
                      <span className="text-neutral-500">-</span>
                    )}
                  </TableCell>
                  <TableCell className="py-4">
                    {material.status ? (
                      <Badge
                        variant={
                          material.status === 'active' ? 'default' : 'secondary'
                        }
                      >
                        {material.status}
                      </Badge>
                    ) : (
                      <span className="text-neutral-500">-</span>
                    )}
                  </TableCell>
                  <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                    {material.parent_uuid ? (
                      <code className="text-xs bg-neutral-100 dark:bg-neutral-800 px-2 py-1 rounded">
                        {String(material.parent_uuid).slice(0, 8)}
                      </code>
                    ) : (
                      '-'
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
