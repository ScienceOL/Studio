/**
 * 📄 实验室详情页面
 *
 * 职责：展示实验室的详细信息和数据
 *
 * 功能：
 * 1. 展示实验室基本信息（详细信息标签页）
 * 2. 展示 Resources 资源列表（Resources 标签页）
 * 3. 展示 Materials 物料信息（Materials 标签页）
 */

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  useLabDetail,
  useLabMembers,
  useMaterials,
  useResourceTemplates,
} from '@/hooks/queries/useEnvironmentQueries';
import type { ResourceTemplate } from '@/types/material';
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/react';
import {
  ArrowLeft,
  Box,
  Bug,
  Info,
  Layers,
  Zap,
  type LucideIcon,
} from 'lucide-react';
import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  ActionDebugPanel,
  ActionPanel,
  DetailsPanel,
  MaterialsPanel,
  ResourceActionDialog,
  ResourceTemplatesPanel,
} from './components';

// 标签页配置
const TABS_CONFIG: Array<{
  icon: LucideIcon;
  label: string;
}> = [
  { icon: Info, label: '详细信息' },
  { icon: Box, label: 'Templates' },
  { icon: Layers, label: 'Materials' },
  { icon: Zap, label: 'Actions' },
  { icon: Bug, label: 'Debug' },
];

export default function EnvironmentDetail() {
  const { labUuid } = useParams<{ labUuid: string }>();
  const navigate = useNavigate();

  // 使用统一的 query hooks
  const { data: lab, isLoading: isLoadingLab } = useLabDetail(labUuid || '');
  const { data: members = [], isLoading: isLoadingMembers } = useLabMembers(
    labUuid || ''
  );

  // 查询 Resource Templates 和 Materials
  const { data: resourceTemplates = [], isLoading: isLoadingResources } =
    useResourceTemplates(labUuid || '');
  const { data: materials = [], isLoading: isLoadingMaterials } = useMaterials(
    labUuid || ''
  );

  // Resource Action Dialog 状态
  const [actionDialogOpen, setActionDialogOpen] = useState(false);
  const [selectedResource, setSelectedResource] =
    useState<ResourceTemplate | null>(null);

  if (!labUuid) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center text-neutral-900 dark:text-neutral-100">
          Invalid lab UUID
        </div>
      </div>
    );
  }

  const handleOpenResourceActions = (template: ResourceTemplate) => {
    setSelectedResource(template);
    setActionDialogOpen(true);
  };

  return (
    <div className="container mx-auto py-8 px-4 space-y-6">
      {/* 返回按钮 */}
      <Button
        variant="ghost"
        onClick={() => navigate('/dashboard/environment')}
        className="mb-6 hover:bg-neutral-100 dark:hover:bg-neutral-800"
      >
        <ArrowLeft className="mr-2 h-4 w-4" />
        返回列表
      </Button>

      {/* 标题区域 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-neutral-900 dark:text-neutral-100">
            {lab?.name || '加载中...'}
          </h1>
          <p className="text-neutral-600 dark:text-neutral-400 mt-1">
            {lab?.description || '暂无描述'}
          </p>
        </div>
        <Badge className="bg-indigo-100 text-indigo-900 dark:bg-indigo-900 dark:text-indigo-100">
          {labUuid.slice(0, 8)}
        </Badge>
      </div>

      {/* Tabs 标签页 */}
      <TabGroup>
        <TabList className="flex space-x-1 rounded-xl bg-neutral-100 dark:bg-neutral-800 p-1">
          {TABS_CONFIG.map((tab) => {
            const Icon = tab.icon;
            return (
              <Tab
                key={tab.label}
                className={({ selected }) =>
                  `w-full rounded-lg py-2.5 text-sm font-medium leading-5 transition-all
                  ${
                    selected
                      ? 'bg-white dark:bg-neutral-700 text-indigo-700 dark:text-indigo-400 shadow'
                      : 'text-neutral-700 dark:text-neutral-300 hover:bg-white/[0.12] hover:text-neutral-900 dark:hover:text-white'
                  }`
                }
              >
                <div className="flex items-center justify-center gap-2">
                  <Icon className="h-4 w-4" />
                  <span>{tab.label}</span>
                </div>
              </Tab>
            );
          })}
        </TabList>

        <TabPanels className="mt-6">
          {/* 详细信息面板 */}
          <TabPanel>
            <DetailsPanel
              labUuid={labUuid}
              lab={lab}
              isLoadingLab={isLoadingLab}
              members={members}
              isLoadingMembers={isLoadingMembers}
            />
          </TabPanel>

          {/* Resources 面板 */}
          <TabPanel>
            <ResourceTemplatesPanel
              resourceTemplates={resourceTemplates}
              isLoading={isLoadingResources}
              onSelectResource={handleOpenResourceActions}
            />
          </TabPanel>

          {/* Materials 面板 */}
          <TabPanel>
            <MaterialsPanel
              materials={materials}
              isLoading={isLoadingMaterials}
              resourceTemplates={resourceTemplates}
              onOpenResourceActions={handleOpenResourceActions}
            />
          </TabPanel>

          {/* Actions 面板 */}
          <TabPanel>
            <ActionPanel
              labUuid={labUuid}
              materials={materials}
              resourceTemplates={resourceTemplates}
              isLoadingMaterials={isLoadingMaterials}
              isLoadingResources={isLoadingResources}
            />
          </TabPanel>

          {/* Debug 面板 */}
          <TabPanel>
            <ActionDebugPanel labUuid={labUuid} />
          </TabPanel>
        </TabPanels>
      </TabGroup>

      {/* Resource Action Dialog */}
      {selectedResource && (
        <ResourceActionDialog
          open={actionDialogOpen}
          onOpenChange={setActionDialogOpen}
          resourceTemplate={selectedResource}
          labUuid={labUuid}
        />
      )}
    </div>
  );
}
