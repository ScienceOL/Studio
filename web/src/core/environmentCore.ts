/**
 * 🎯 Core Layer - Environment 核心业务逻辑
 *
 * 职责：
 * 1. 编排复杂的业务流程
 * 2. 调用 Service 层进行数据操作
 * 3. 更新 Store 状态
 * 4. 处理副作用（通知、日志等）
 *
 * 注意：Core 直接调用 Service，不调用 Query Hook
 */

import { environmentService } from '@/service';
import { useEnvironmentStore } from '@/store/environmentStore';

export class EnvironmentCore {
  /**
   * 进入实验室（复杂流程：验证 → 设置状态 → 可能的副作用）
   */
  static async enterLab(labUuid: string): Promise<void> {
    console.log('🚪 [EnvironmentCore] Entering lab:', labUuid);

    try {
      // 1. 验证实验室是否存在（调用 Service）
      const labInfo = await environmentService.getLabInfo(labUuid);

      if (!labInfo || labInfo.code !== 0) {
        throw new Error('实验室不存在或无权访问');
      }

      // 2. 更新 Store 状态
      const store = useEnvironmentStore.getState();
      store.setCurrentLabUuid(labUuid);
      store.setSelectedLabUuid(labUuid);

      console.log('✅ [EnvironmentCore] Entered lab successfully');

      // 3. 可以添加其他副作用，比如记录日志、发送分析事件等
      // Analytics.track('lab_entered', { labUuid });
    } catch (error) {
      console.error('❌ [EnvironmentCore] Failed to enter lab:', error);
      throw error;
    }
  }

  /**
   * 退出实验室
   */
  static exitLab(): void {
    console.log('🚪 [EnvironmentCore] Exiting lab');

    const store = useEnvironmentStore.getState();
    store.setCurrentLabUuid(null);
    store.setSelectedLabUuid(null);
  }

  /**
   * 创建实验室并进入（复杂流程编排）
   */
  static async createAndEnterLab(data: {
    name: string;
    description?: string;
  }): Promise<string> {
    console.log('🏗️ [EnvironmentCore] Creating and entering lab:', data.name);

    try {
      // 1. 创建实验室
      const result = await environmentService.createLab(data);

      if (!result || result.code !== 0 || !result.data?.uuid) {
        throw new Error('创建实验室失败');
      }

      const labUuid = result.data.uuid;

      // 2. 自动进入新创建的实验室
      await this.enterLab(labUuid);

      console.log('✅ [EnvironmentCore] Lab created and entered:', labUuid);

      return labUuid;
    } catch (error) {
      console.error('❌ [EnvironmentCore] Failed to create lab:', error);
      throw error;
    }
  }

  /**
   * 获取实验室凭证（从后端获取已存在的 AK/SK）
   */
  static async getLabCredentials(
    labUuid: string
  ): Promise<{ accessKey: string; secretKey: string }> {
    console.log('🔑 [EnvironmentCore] Getting credentials for lab:', labUuid);

    try {
      // 调用 getLabInfo 获取实验室详情（包含 AK/SK）
      const result = await environmentService.getLabInfo(labUuid);

      if (!result || result.code !== 0 || !result.data) {
        throw new Error('无法获取实验室信息');
      }

      const { access_key, access_secret } = result.data;

      if (!access_key || !access_secret) {
        throw new Error('实验室凭证不存在');
      }

      console.log('✅ [EnvironmentCore] Credentials retrieved');
      return {
        accessKey: access_key,
        secretKey: access_secret,
      };
    } catch (error) {
      console.error('❌ [EnvironmentCore] Failed to get credentials:', error);
      throw error;
    }
  }

  /**
   * 复制文本到剪贴板（带通知）
   */
  static async copyToClipboard(text: string, label = '内容'): Promise<void> {
    try {
      await navigator.clipboard.writeText(text);
      console.log(`📋 [EnvironmentCore] Copied ${label} to clipboard`);

      // 可以触发通知
      // toast.success(`${label}已复制到剪贴板`);
    } catch (error) {
      console.error('❌ [EnvironmentCore] Failed to copy:', error);
      throw new Error(`复制${label}失败`);
    }
  }

  /**
   * 删除实验室成员（带确认流程）
   */
  static async removeMember(
    labUuid: string,
    memberUuid: string,
    memberName?: string
  ): Promise<void> {
    console.log(
      '🗑️ [EnvironmentCore] Removing member:',
      memberName || memberUuid
    );

    try {
      // 1. 调用删除 API
      await environmentService.deleteLabMember(labUuid, memberUuid);

      console.log('✅ [EnvironmentCore] Member removed successfully');

      // 2. 可以添加通知
      // toast.success(`已移除成员 ${memberName || memberUuid}`);
    } catch (error) {
      console.error('❌ [EnvironmentCore] Failed to remove member:', error);
      throw error;
    }
  }

  /**
   * 接受邀请并进入实验室（复杂流程）
   */
  static async acceptInviteAndEnter(inviteUuid: string): Promise<string> {
    console.log('📨 [EnvironmentCore] Accepting invite:', inviteUuid);

    try {
      // 1. 接受邀请
      const result = await environmentService.acceptInvite(inviteUuid);

      if (!result || result.code !== 0 || !result.data?.lab_uuid) {
        throw new Error('接受邀请失败');
      }

      const labUuid = result.data.lab_uuid;

      // 2. 进入实验室
      await this.enterLab(labUuid);

      console.log('✅ [EnvironmentCore] Invite accepted and lab entered');

      return labUuid;
    } catch (error) {
      console.error('❌ [EnvironmentCore] Failed to accept invite:', error);
      throw error;
    }
  }
}
