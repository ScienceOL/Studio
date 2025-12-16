/**
 * 📄 Environment 页面
 *
 * 职责：UI 渲染和事件绑定
 *
 * 功能：
 * 1. 展示实验室列表
 * 2. 创建新实验室
 * 3. 查看 AK/SK
 * 4. 点击进入实验室详情
 */

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  LocalDialog as Dialog,
  LocalDialogContent as DialogContent,
  LocalDialogDescription as DialogDescription,
  LocalDialogFooter as DialogFooter,
  LocalDialogHeader as DialogHeader,
  LocalDialogTitle as DialogTitle,
} from '@/components/ui/local-dialog';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Textarea } from '@/components/ui/textarea';
import { useEnvironment } from '@/hooks/useEnvironment';
import { useLabStatus } from '@/hooks/useLabStatus';
import type { Lab } from '@/types/environment';
import {
  ArrowRight,
  CheckCheck,
  Copy,
  Grid,
  Key,
  List,
  Plus,
} from 'lucide-react';
import { useState } from 'react';
import EnvironmentDetail from './EnvironmentDetail';
import { LabStatusIndicator } from './components';

export default function EnvironmentPage() {
  const [selectedLabUuid, setSelectedLabUuid] = useState<string | null>(null);
  const environment = useEnvironment();

  // 实验室在线状态监控（自动查询列表）
  const { labStatuses } = useLabStatus({
    autoQueryList: true, // 自动查询所有实验室状态
    onStatusUpdate: (_statuses) => {
      console.log('Lab statuses updated:', _statuses);
    },
  });

  // 本地表单状态
  const [labName, setLabName] = useState('');
  const [labDescription, setLabDescription] = useState('');
  const [credentials, setCredentials] = useState<{
    accessKey: string;
    secretKey: string;
  } | null>(null);
  const [copiedField, setCopiedField] = useState<string | null>(null);

  // ========== 事件处理 ==========

  // 创建实验室
  const handleCreateLab = async () => {
    if (!labName.trim()) return;

    try {
      await environment.createAndEnterLab({
        name: labName,
        description: labDescription || undefined,
      });

      // 重置表单并关闭对话框
      setLabName('');
      setLabDescription('');
      environment.setCreateDialogOpen(false);
    } catch (error) {
      console.error('Failed to create lab:', error);
      // TODO: 显示错误通知
    }
  };

  // 查看凭证
  const handleViewCredentials = async (labUuid: string) => {
    try {
      const creds = await environment.getLabCredentials(labUuid);
      setCredentials(creds);
      environment.setCredentialsDialogOpen(true);
    } catch (error) {
      console.error('Failed to get credentials:', error);
    }
  };

  // 复制到剪贴板
  const handleCopy = async (text: string, field: string) => {
    try {
      await environment.copyToClipboard(text, field);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch (error) {
      console.error('Failed to copy:', error);
    }
  };

  // 进入实验室
  const handleEnterLab = async (labUuid: string) => {
    try {
      await environment.enterLab(labUuid);
      // 在桌面模式下，切换到详情视图而不是跳转路由
      setSelectedLabUuid(labUuid);
    } catch (error) {
      console.error('Failed to enter lab:', error);
    }
  };

  // ========== 渲染 ==========

  // 如果选中了实验室，显示详情页
  if (selectedLabUuid) {
    return (
      <EnvironmentDetail
        labUuid={selectedLabUuid}
        onBack={() => setSelectedLabUuid(null)}
      />
    );
  }

  return (
    <div className="relative flex flex-1 min-h-0 w-full flex-col overflow-hidden">
      <div className="flex-1 overflow-y-auto">
        <div className="container mx-auto space-y-6 py-8 px-4">
          {/* 头部 */}
          <div className="flex items-center justify-between mb-8">
            <div className="space-y-2">
              <h1 className="text-3xl font-bold text-neutral-900 dark:text-neutral-100">
                实验室环境
              </h1>
              <p className="text-neutral-600 dark:text-neutral-400 mt-2">
                管理你的实验室环境和访问凭证
              </p>
            </div>
            <div className="flex items-center gap-3">
              {/* 视图切换 */}
              <Button
                variant={
                  environment.viewMode === 'grid' ? 'default' : 'outline'
                }
                size="icon"
                onClick={() => environment.setViewMode('grid')}
                className="hover:bg-neutral-100 dark:hover:bg-neutral-800"
              >
                <Grid className="h-4 w-4" />
              </Button>
              <Button
                variant={
                  environment.viewMode === 'list' ? 'default' : 'outline'
                }
                size="icon"
                onClick={() => environment.setViewMode('list')}
                className="hover:bg-neutral-100 dark:hover:bg-neutral-800"
              >
                <List className="h-4 w-4" />
              </Button>
              {/* 创建按钮 */}
              <Button
                onClick={() => environment.setCreateDialogOpen(true)}
                className="ml-2"
              >
                <Plus className="mr-2 h-4 w-4" />
                创建实验室
              </Button>
            </div>
          </div>

          {/* 实验室列表 */}
          {environment.isLoadingLabs ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6">
              {[1, 2, 3].map((i) => (
                <Card
                  key={i}
                  className="border-neutral-200 dark:border-neutral-800"
                >
                  <CardHeader className="space-y-3">
                    <Skeleton className="h-6 w-3/4 bg-neutral-200 dark:bg-neutral-700" />
                    <Skeleton className="h-4 w-full mt-2 bg-neutral-200 dark:bg-neutral-700" />
                  </CardHeader>
                  <CardContent>
                    <Skeleton className="h-4 w-1/2 bg-neutral-200 dark:bg-neutral-700" />
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : environment.labs.length === 0 ? (
            <Card className="border-neutral-200 dark:border-neutral-800 mt-6">
              <CardContent className="flex mt-12 flex-col items-center justify-center py-16">
                <p className="text-neutral-600 dark:text-neutral-400 mb-6 text-lg">
                  暂无实验室
                </p>
                <Button onClick={() => environment.setCreateDialogOpen(true)}>
                  <Plus className="mr-2 h-4 w-4" />
                  创建第一个实验室
                </Button>
              </CardContent>
            </Card>
          ) : environment.viewMode === 'grid' ? (
            // 网格视图
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6">
              {environment.labs.map((lab: Lab) => {
                // 获取实时状态
                const realtimeStatus = labStatuses.get(lab.uuid);
                const isOnline =
                  realtimeStatus?.is_online ?? lab.is_online ?? false;
                const lastConnectedAt =
                  realtimeStatus?.last_connected_at ?? lab.last_connected_at;

                // Debug log
                // console.log(`Lab ${lab.name} (${lab.uuid}): realtime=${realtimeStatus?.is_online}, static=${lab.is_online}, final=${isOnline}`);

                return (
                  <Card
                    key={lab.uuid}
                    className="hover:shadow-lg dark:hover:shadow-neutral-900/50 transition-all duration-200 cursor-pointer border-neutral-200 dark:border-neutral-800 hover:border-neutral-300 dark:hover:border-neutral-700"
                  >
                    <CardHeader className="space-y-3">
                      <div className="flex items-start justify-between gap-2">
                        <CardTitle className="text-neutral-900 dark:text-neutral-100 flex-1">
                          {lab.name}
                        </CardTitle>
                        <LabStatusIndicator
                          isOnline={isOnline}
                          lastConnectedAt={lastConnectedAt}
                          showText={false}
                          size="sm"
                        />
                      </div>
                      <CardDescription className="text-neutral-600 dark:text-neutral-400 line-clamp-2">
                        {lab.description || '暂无描述'}
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-3">
                      {/* 状态信息 */}
                      <LabStatusIndicator
                        isOnline={isOnline}
                        lastConnectedAt={lastConnectedAt}
                        showText={true}
                        showTime={true}
                        size="sm"
                      />
                      {/* 其他信息 */}
                      <div className="flex items-center gap-2 text-sm text-neutral-600 dark:text-neutral-400">
                        <Badge
                          variant="outline"
                          className="border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300"
                        >
                          {lab.uuid.slice(0, 8)}
                        </Badge>
                        <span>•</span>
                        <span>
                          {new Date(lab.created_at).toLocaleDateString('zh-CN')}
                        </span>
                      </div>
                    </CardContent>
                    <CardFooter className="flex gap-3 pt-4">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleViewCredentials(lab.uuid)}
                        className="hover:bg-neutral-100 dark:hover:bg-neutral-800"
                      >
                        <Key className="mr-2 h-4 w-4" />
                        查看凭证
                      </Button>
                      <Button
                        size="sm"
                        onClick={() => handleEnterLab(lab.uuid)}
                        className="ml-auto"
                      >
                        进入
                        <ArrowRight className="ml-2 h-4 w-4" />
                      </Button>
                    </CardFooter>
                  </Card>
                );
              })}
            </div>
          ) : (
            // 列表视图
            <Card className="border-neutral-200 dark:border-neutral-800 mt-6">
              <Table>
                <TableHeader>
                  <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                    <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                      名称
                    </TableHead>
                    <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                      描述
                    </TableHead>
                    <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                      UUID
                    </TableHead>
                    <TableHead className="text-neutral-700 dark:text-neutral-300 py-4">
                      创建时间 / 状态
                    </TableHead>
                    <TableHead className="text-right text-neutral-700 dark:text-neutral-300 py-4">
                      操作
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {environment.labs.map((lab: Lab) => {
                    // 获取实时状态
                    const realtimeStatus = labStatuses.get(lab.uuid);
                    const isOnline =
                      realtimeStatus?.is_online ?? lab.is_online ?? false;
                    const lastConnectedAt =
                      realtimeStatus?.last_connected_at ??
                      lab.last_connected_at;

                    return (
                      <TableRow
                        key={lab.uuid}
                        className="border-b border-neutral-200 dark:border-neutral-800 hover:bg-neutral-50 dark:hover:bg-neutral-800/50 transition-colors"
                      >
                        <TableCell className="font-medium text-neutral-900 dark:text-neutral-100 py-4">
                          <div className="flex items-center gap-3">
                            <LabStatusIndicator
                              isOnline={isOnline}
                              showText={false}
                              size="sm"
                            />
                            {lab.name}
                          </div>
                        </TableCell>
                        <TableCell className="text-neutral-700 dark:text-neutral-300 py-4 max-w-xs truncate">
                          {lab.description || '-'}
                        </TableCell>
                        <TableCell className="py-4">
                          <Badge
                            variant="outline"
                            className="border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300"
                          >
                            {lab.uuid.slice(0, 8)}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                          <div className="space-y-1">
                            <div>
                              {new Date(lab.created_at).toLocaleDateString(
                                'zh-CN'
                              )}
                            </div>
                            <LabStatusIndicator
                              isOnline={isOnline}
                              lastConnectedAt={lastConnectedAt}
                              showText={true}
                              showTime={!isOnline}
                              size="sm"
                            />
                          </div>
                        </TableCell>
                        <TableCell className="text-right py-4">
                          <div className="flex justify-end gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleViewCredentials(lab.uuid)}
                              className="hover:bg-neutral-100 dark:hover:bg-neutral-800"
                            >
                              <Key className="h-4 w-4" />
                            </Button>
                            <Button
                              size="sm"
                              onClick={() => handleEnterLab(lab.uuid)}
                            >
                              进入
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </Card>
          )}

          {/* 创建实验室对话框 */}
          <Dialog
            open={environment.isCreateDialogOpen}
            onOpenChange={environment.setCreateDialogOpen}
            size="md"
          >
            <DialogContent className="bg-white m-4 dark:bg-neutral-900 border-neutral-200 dark:border-neutral-800">
              <DialogHeader className="space-y-3">
                <DialogTitle className="text-xl text-neutral-900 dark:text-neutral-100">
                  创建实验室
                </DialogTitle>
                <DialogDescription className="text-neutral-600 dark:text-neutral-400">
                  创建一个新的实验室环境来管理你的资源
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-6 py-4">
                <div className="space-y-3">
                  <Label
                    htmlFor="name"
                    className="text-sm font-medium text-neutral-900 dark:text-neutral-100"
                  >
                    名称 *
                  </Label>
                  <Input
                    id="name"
                    placeholder="输入实验室名称"
                    value={labName}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                      setLabName(e.target.value)
                    }
                    className="bg-white dark:bg-neutral-800 border-neutral-300 dark:border-neutral-700 text-neutral-900 dark:text-neutral-100 placeholder:text-neutral-500 dark:placeholder:text-neutral-500"
                  />
                </div>
                <div className="space-y-3">
                  <Label
                    htmlFor="description"
                    className="text-sm font-medium text-neutral-900 dark:text-neutral-100"
                  >
                    描述
                  </Label>
                  <Textarea
                    id="description"
                    placeholder="输入实验室描述（可选）"
                    value={labDescription}
                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                      setLabDescription(e.target.value)
                    }
                    rows={3}
                    className="bg-white dark:bg-neutral-800 border-neutral-300 dark:border-neutral-700 text-neutral-900 dark:text-neutral-100 placeholder:text-neutral-500 dark:placeholder:text-neutral-500"
                  />
                </div>
              </div>
              <DialogFooter className="gap-2">
                <Button
                  variant="outline"
                  onClick={() => environment.setCreateDialogOpen(false)}
                  className="hover:bg-neutral-100 dark:hover:bg-neutral-800"
                >
                  取消
                </Button>
                <Button
                  onClick={handleCreateLab}
                  disabled={!labName.trim() || environment.isCreating}
                >
                  {environment.isCreating ? '创建中...' : '创建'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* AK/SK 凭证对话框 */}
          <Dialog
            open={environment.isCredentialsDialogOpen}
            onOpenChange={environment.setCredentialsDialogOpen}
            size="md"
          >
            <DialogContent className="bg-white mx-4 dark:bg-neutral-900 border-neutral-200 dark:border-neutral-800">
              <DialogHeader className="space-y-3">
                <DialogTitle className="text-xl text-neutral-900 dark:text-neutral-100 flex items-center gap-2">
                  <Key className="h-5 w-5" />
                  访问凭证
                </DialogTitle>
                <DialogDescription className="text-neutral-600 dark:text-neutral-400">
                  请妥善保管你的访问凭证，不要泄露给他人
                </DialogDescription>
              </DialogHeader>
              {credentials && (
                <div className="space-y-6 py-4">
                  <div className="space-y-3">
                    <Label className="text-sm font-medium text-neutral-900 dark:text-neutral-100">
                      Access Key (AK)
                    </Label>
                    <div className="flex gap-2">
                      <Input
                        value={credentials.accessKey}
                        readOnly
                        className="bg-neutral-50 dark:bg-neutral-800 border-neutral-300 dark:border-neutral-700 text-neutral-900 dark:text-neutral-100 font-mono text-sm"
                      />
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={() =>
                          handleCopy(credentials.accessKey, 'accessKey')
                        }
                        className="hover:bg-neutral-100 dark:hover:bg-neutral-800 shrink-0"
                      >
                        {copiedField === 'accessKey' ? (
                          <CheckCheck className="h-4 w-4 text-green-500" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>
                  <div className="space-y-3">
                    <Label className="text-sm font-medium text-neutral-900 dark:text-neutral-100">
                      Secret Key (SK)
                    </Label>
                    <div className="flex gap-2">
                      <Input
                        value={credentials.secretKey}
                        readOnly
                        type="password"
                        className="bg-neutral-50 dark:bg-neutral-800 border-neutral-300 dark:border-neutral-700 text-neutral-900 dark:text-neutral-100 font-mono text-sm"
                      />
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={() =>
                          handleCopy(credentials.secretKey, 'secretKey')
                        }
                        className="hover:bg-neutral-100 dark:hover:bg-neutral-800 shrink-0"
                      >
                        {copiedField === 'secretKey' ? (
                          <CheckCheck className="h-4 w-4 text-green-500" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>
                </div>
              )}
              <DialogFooter>
                <Button
                  onClick={() => environment.setCredentialsDialogOpen(false)}
                  className="w-full sm:w-auto"
                >
                  关闭
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>
    </div>
  );
}
