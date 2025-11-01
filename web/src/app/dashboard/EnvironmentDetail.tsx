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
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useLabInfo, useLabMembersList } from '@/hooks/useEnvironment';
import { materialService } from '@/service/materialService';
import type { LabMember } from '@/types/environment';
import type { Material, ResourceInfo } from '@/types/material';
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/react';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, Box, Info, Layers, Users } from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';

export default function EnvironmentDetail() {
  const { labUuid } = useParams<{ labUuid: string }>();
  const navigate = useNavigate();

  const { lab, isLoading: isLoadingLab } = useLabInfo(labUuid || '');
  const { members, isLoading: isLoadingMembers } = useLabMembersList(
    labUuid || ''
  );

  // 查询 Resources
  const { data: resourcesData, isLoading: isLoadingResources } = useQuery({
    queryKey: ['resources', labUuid],
    queryFn: () => materialService.getResourceList({ lab_uuid: labUuid || '' }),
    enabled: !!labUuid,
  });

  // 查询 Materials
  const { data: materialsData, isLoading: isLoadingMaterials } = useQuery({
    queryKey: ['materials', labUuid],
    queryFn: () => materialService.downloadMaterial(labUuid || ''),
    enabled: !!labUuid,
  });

  if (!labUuid) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center text-neutral-900 dark:text-neutral-100">
          Invalid lab UUID
        </div>
      </div>
    );
  }

  // 注意：后端返回的是 { code: 0, data: { resource_name_list: [...] } }
  const resources = (resourcesData?.data?.resource_name_list ||
    []) as ResourceInfo[];
  const materials = (materialsData?.data?.nodes || []) as Material[];

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
          <Tab
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
              <Info className="h-4 w-4" />
              <span>详细信息</span>
            </div>
          </Tab>
          <Tab
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
              <Box className="h-4 w-4" />
              <span>Resources</span>
            </div>
          </Tab>
          <Tab
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
              <Layers className="h-4 w-4" />
              <span>Materials</span>
            </div>
          </Tab>
        </TabList>

        <TabPanels className="mt-6">
          {/* 详细信息面板 */}
          <TabPanel>
            <div className="space-y-6">
              {/* 基本信息 */}
              <Card className="border-neutral-200 dark:border-neutral-800">
                <CardHeader className="space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="space-y-2">
                      <CardTitle className="flex items-center gap-2 text-neutral-900 dark:text-neutral-100">
                        <Info className="h-5 w-5" />
                        实验室信息
                      </CardTitle>
                      <CardDescription className="text-neutral-600 dark:text-neutral-400">
                        查看实验室的详细信息
                      </CardDescription>
                    </div>
                    <Badge className="bg-indigo-100 text-indigo-900 dark:bg-indigo-900 dark:text-indigo-100">
                      {labUuid.slice(0, 8)}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="pt-6">
                  {isLoadingLab ? (
                    <div className="space-y-4 p-4">
                      <Skeleton className="h-6 w-1/2 bg-neutral-200 dark:bg-neutral-700" />
                      <Skeleton className="h-4 w-3/4 bg-neutral-200 dark:bg-neutral-700" />
                      <Skeleton className="h-4 w-1/4 bg-neutral-200 dark:bg-neutral-700" />
                    </div>
                  ) : lab ? (
                    <Table>
                      <TableBody>
                        <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                          <TableCell className="font-medium text-neutral-900 dark:text-neutral-100 py-4">
                            名称
                          </TableCell>
                          <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                            {lab.name}
                          </TableCell>
                        </TableRow>
                        <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                          <TableCell className="font-medium text-neutral-900 dark:text-neutral-100 py-4">
                            描述
                          </TableCell>
                          <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                            {lab.description || '-'}
                          </TableCell>
                        </TableRow>
                        <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                          <TableCell className="font-medium text-neutral-900 dark:text-neutral-100 py-4">
                            UUID
                          </TableCell>
                          <TableCell className="py-4">
                            <code className="bg-neutral-100 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 px-3 py-1.5 rounded text-sm">
                              {lab.uuid}
                            </code>
                          </TableCell>
                        </TableRow>
                        <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                          <TableCell className="font-medium text-neutral-900 dark:text-neutral-100 py-4">
                            所有者
                          </TableCell>
                          <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                            {lab.owner_uuid}
                          </TableCell>
                        </TableRow>
                        <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                          <TableCell className="font-medium text-neutral-900 dark:text-neutral-100 py-4">
                            创建时间
                          </TableCell>
                          <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                            {new Date(lab.created_at).toLocaleString('zh-CN')}
                          </TableCell>
                        </TableRow>
                        <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                          <TableCell className="font-medium text-neutral-900 dark:text-neutral-100 py-4">
                            更新时间
                          </TableCell>
                          <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                            {new Date(lab.updated_at).toLocaleString('zh-CN')}
                          </TableCell>
                        </TableRow>
                        {/* 展示所有其他数据 */}
                        {Object.entries(lab)
                          .filter(
                            ([key]) =>
                              ![
                                'name',
                                'description',
                                'uuid',
                                'owner_uuid',
                                'created_at',
                                'updated_at',
                                'code',
                                'message',
                              ].includes(key)
                          )
                          .map(([key, value]) => (
                            <TableRow
                              key={key}
                              className="border-b border-neutral-200 dark:border-neutral-800"
                            >
                              <TableCell className="font-medium text-neutral-900 dark:text-neutral-100 py-4">
                                {key}
                              </TableCell>
                              <TableCell className="py-4">
                                <code className="bg-neutral-100 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 px-3 py-1.5 rounded text-sm block break-all">
                                  {typeof value === 'object'
                                    ? JSON.stringify(value, null, 2)
                                    : String(value)}
                                </code>
                              </TableCell>
                            </TableRow>
                          ))}
                      </TableBody>
                    </Table>
                  ) : (
                    <div className="text-center py-12 text-neutral-500 dark:text-neutral-400">
                      加载失败或实验室不存在
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* 成员列表 */}
              <Card className="border-neutral-200 dark:border-neutral-800">
                <CardHeader className="space-y-2">
                  <CardTitle className="flex items-center gap-2 text-neutral-900 dark:text-neutral-100">
                    <Users className="h-5 w-5" />
                    成员列表
                  </CardTitle>
                  <CardDescription className="text-neutral-600 dark:text-neutral-400">
                    实验室成员及其角色
                  </CardDescription>
                </CardHeader>
                <CardContent className="pt-6">
                  {isLoadingMembers ? (
                    <div className="space-y-3 p-4">
                      {[1, 2, 3].map((i) => (
                        <Skeleton
                          key={i}
                          className="h-12 w-full bg-neutral-200 dark:bg-neutral-700"
                        />
                      ))}
                    </div>
                  ) : members.length === 0 ? (
                    <div className="text-center py-12 text-neutral-500 dark:text-neutral-400">
                      暂无成员
                    </div>
                  ) : (
                    <Table>
                      <TableHeader>
                        <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                          <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                            UUID
                          </TableHead>
                          <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                            用户
                          </TableHead>
                          <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                            角色
                          </TableHead>
                          <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                            加入时间
                          </TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {members.map((member: LabMember) => (
                          <TableRow
                            key={member.uuid}
                            className="border-b border-neutral-200 dark:border-neutral-800"
                          >
                            <TableCell className="py-4">
                              <Badge
                                variant="outline"
                                className="border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300"
                              >
                                {member.uuid.slice(0, 8)}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-neutral-900 dark:text-neutral-100 py-4">
                              {member.username ||
                                member.email ||
                                member.user_uuid}
                            </TableCell>
                            <TableCell className="py-4">
                              <Badge className="bg-indigo-100 text-indigo-900 dark:bg-indigo-900 dark:text-indigo-100">
                                {member.role}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                              {new Date(member.created_at).toLocaleString(
                                'zh-CN'
                              )}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  )}
                </CardContent>
              </Card>

              {/* 原始数据展示 */}
              <Card className="border-neutral-200 dark:border-neutral-800">
                <CardHeader className="space-y-2">
                  <CardTitle className="text-neutral-900 dark:text-neutral-100">
                    原始数据
                  </CardTitle>
                  <CardDescription className="text-neutral-600 dark:text-neutral-400">
                    实验室的完整 JSON 数据
                  </CardDescription>
                </CardHeader>
                <CardContent className="pt-6">
                  <pre className="bg-neutral-100 dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 p-6 rounded-lg overflow-auto max-h-96">
                    <code className="text-sm text-neutral-900 dark:text-neutral-100">
                      {JSON.stringify({ lab, members }, null, 2)}
                    </code>
                  </pre>
                </CardContent>
              </Card>
            </div>
          </TabPanel>

          {/* Resources 面板 */}
          <TabPanel>
            <Card className="border-neutral-200 dark:border-neutral-800">
              <CardHeader className="space-y-2">
                <CardTitle className="flex items-center gap-2 text-neutral-900 dark:text-neutral-100">
                  <Box className="h-5 w-5" />
                  Resource 资源列表
                </CardTitle>
                <CardDescription className="text-neutral-600 dark:text-neutral-400">
                  实验室中已注册的资源模板
                </CardDescription>
              </CardHeader>
              <CardContent className="pt-6">
                {isLoadingResources ? (
                  <div className="space-y-3 p-4">
                    {[1, 2, 3].map((i) => (
                      <Skeleton
                        key={i}
                        className="h-12 w-full bg-neutral-200 dark:bg-neutral-700"
                      />
                    ))}
                  </div>
                ) : resources.length === 0 ? (
                  <div className="text-center py-12 text-neutral-500 dark:text-neutral-400">
                    暂无资源
                  </div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                        <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                          UUID
                        </TableHead>
                        <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                          名称
                        </TableHead>
                        <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                          父节点 UUID
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {resources.map((resource) => (
                        <TableRow
                          key={resource.uuid}
                          className="border-b border-neutral-200 dark:border-neutral-800"
                        >
                          <TableCell className="py-4">
                            <Badge
                              variant="outline"
                              className="border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300"
                            >
                              {String(resource.uuid).slice(0, 8) || 'N/A'}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-neutral-900 dark:text-neutral-100 py-4">
                            {resource.name || 'Unnamed'}
                          </TableCell>
                          <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                            {resource.parent_uuid ? (
                              <code className="text-xs bg-neutral-100 dark:bg-neutral-800 px-2 py-1 rounded">
                                {String(resource.parent_uuid).slice(0, 8)}
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
          </TabPanel>

          {/* Materials 面板 */}
          <TabPanel>
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
                {isLoadingMaterials ? (
                  <div className="space-y-3 p-4">
                    {[1, 2, 3].map((i) => (
                      <Skeleton
                        key={i}
                        className="h-12 w-full bg-neutral-200 dark:bg-neutral-700"
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
                          UUID
                        </TableHead>
                        <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                          名称
                        </TableHead>
                        <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                          类型
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
                              className="border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300"
                            >
                              {material.uuid?.slice(0, 8) || 'N/A'}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-neutral-900 dark:text-neutral-100 py-4">
                            {String(material.name || material.id || 'Unnamed')}
                          </TableCell>
                          <TableCell className="py-4">
                            <Badge className="bg-green-100 text-green-900 dark:bg-green-900 dark:text-green-100">
                              {String(
                                material.type || material.class || 'Unknown'
                              )}
                            </Badge>
                          </TableCell>
                          <TableCell className="py-4">
                            {material.status ? (
                              <Badge className="bg-yellow-100 text-yellow-900 dark:bg-yellow-900 dark:text-yellow-100">
                                {String(material.status)}
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
          </TabPanel>
        </TabPanels>
      </TabGroup>
    </div>
  );
}
