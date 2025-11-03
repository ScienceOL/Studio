import { Badge } from '@/components/ui/badge';
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
import type { Lab, LabMember } from '@/types/environment';
import { Info, Users } from 'lucide-react';

interface DetailsPanelProps {
  labUuid: string;
  lab?: Lab;
  isLoadingLab: boolean;
  members: LabMember[];
  isLoadingMembers: boolean;
}

export default function DetailsPanel({
  labUuid,
  lab,
  isLoadingLab,
  members,
  isLoadingMembers,
}: DetailsPanelProps) {
  return (
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
                  <TableCell className="py-4">
                    <code className="bg-neutral-100 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 px-3 py-1.5 rounded text-sm">
                      {lab.user_id}
                    </code>
                  </TableCell>
                </TableRow>
                <TableRow className="border-b border-neutral-200 dark:border-neutral-800">
                  <TableCell className="font-medium text-neutral-900 dark:text-neutral-100 py-4">
                    状态
                  </TableCell>
                  <TableCell className="py-4">
                    <Badge
                      className={
                        lab.is_admin
                          ? 'bg-purple-100 text-purple-900 dark:bg-purple-900 dark:text-purple-100'
                          : 'bg-blue-100 text-blue-900 dark:bg-blue-900 dark:text-blue-100'
                      }
                    >
                      {lab.is_admin ? '管理员' : '成员'}
                    </Badge>
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
                        'user_id',
                        'owner_uuid',
                        'created_at',
                        'updated_at',
                        'is_admin',
                        'status',
                        'access_key',
                        'access_secret',
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
                      {member.user_id}
                    </TableCell>
                    <TableCell className="py-4">
                      <Badge
                        className={
                          member.is_admin
                            ? 'bg-purple-100 text-purple-900 dark:bg-purple-900 dark:text-purple-100'
                            : 'bg-indigo-100 text-indigo-900 dark:bg-indigo-900 dark:text-indigo-100'
                        }
                      >
                        {member.is_admin ? 'admin' : member.role}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-neutral-700 dark:text-neutral-300 py-4">
                      {member.created_at
                        ? new Date(member.created_at).toLocaleString('zh-CN')
                        : '-'}
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
  );
}
