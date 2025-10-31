/**
 * 🎯 Core Layer - 认证核心业务逻辑
 *
 * 职责：
 * 1. 编排业务流程
 * 2. 调用 Store 管理状态
 * 3. 调用 Utils 处理认证逻辑
 * 4. 处理副作用（日志、通知等）
 */

import { authService } from '@/service/authService';
import { configureApiClient } from '@/service/http/client';
import { useAuthStore } from '@/store/authStore';
import { AuthUtils, type UserInfo } from '@/utils/auth';

export class AuthCore {
  // 避免并发刷新导致的重复请求与竞态
  private static refreshInFlight: Promise<boolean> | null = null;
  /**
   * 初始化认证状态
   * 业务流程：检查本地存储 → 验证 token → 刷新 token → 更新状态
   */
  static async initialize(): Promise<void> {
    console.log('🔐 [AuthCore] Starting initialization...');

    const store = useAuthStore.getState();
    store.setLoading(true);

    // 注入 apiClient 依赖（只需要执行一次，重复调用也安全）
    configureApiClient({
      getAccessToken: () => AuthUtils.getAccessToken(),
      refreshToken: () => AuthCore.refreshToken(),
      onAuthFailure: () => AuthCore.logout(),
    });

    try {
      // 1. 检查是否已认证
      const isAuthenticated = AuthUtils.isAuthenticated();
      console.log('🔑 [AuthCore] Is authenticated:', isAuthenticated);

      if (isAuthenticated) {
        // 2. 获取用户信息
        const userInfo = AuthUtils.getUserInfo();
        console.log('👤 [AuthCore] User info:', userInfo?.name || 'no user');

        store.setUser(userInfo);
        store.setAuthenticated(true);
      } else {
        // 3. 尝试刷新 token
        const hasRefreshToken = AuthUtils.getRefreshToken();

        if (hasRefreshToken) {
          console.log('🔄 [AuthCore] Trying to refresh token...');
          const refreshed = await this.refreshToken();

          if (!refreshed) {
            console.log('❌ [AuthCore] Refresh failed, clearing state');
            this.clearAuthState();
          }
        } else {
          console.log('ℹ️ [AuthCore] No refresh token, skipping refresh');
          this.clearAuthState();
        }
      }
    } catch (error) {
      console.error('❌ [AuthCore] Initialization failed:', error);
      this.clearAuthState();
    } finally {
      store.setLoading(false);
    }
  }

  /**
   * 登录
   * 业务流程：重定向到登录页
   */
  static login(returnUrl?: string): void {
    console.log('🚀 [AuthCore] Redirecting to login, returnUrl:', returnUrl);
    AuthUtils.redirectToLogin(returnUrl);
  }

  /**
   * 登出
   * 业务流程：清除状态 → 清除本地存储 → 重定向
   */
  static logout(): void {
    console.log('👋 [AuthCore] Logging out...');

    // 1. 清除 Store 状态
    this.clearAuthState();

    // 2. 清除本地存储
    AuthUtils.clearAuthInfo();

    // 3. 重定向到首页
    if (typeof window !== 'undefined') {
      window.location.href = '/';
    }
  }

  /**
   * 刷新 Token
   * 业务流程：调用刷新接口 → 保存新 token → 更新用户信息
   */
  static async refreshToken(): Promise<boolean> {
    console.log('🔄 [AuthCore] Refreshing token...');

    // 并发保护：复用进行中的刷新请求
    if (this.refreshInFlight) {
      return this.refreshInFlight;
    }

    this.refreshInFlight = (async () => {
      try {
        // 1. 从本地获取 refresh token
        const refreshToken = AuthUtils.getRefreshToken();
        if (!refreshToken) {
          // 没有 refresh token，清理状态并中止
          this.clearAuthState();
          AuthUtils.clearAuthInfo();
          return false;
        }

        // 2. 调用刷新 API（Service 层）
        const data = await authService.refreshToken(refreshToken);

        // 3. 保存 token
        if (data?.code === 0 && data?.data) {
          AuthUtils.saveAuthInfo({
            accessToken: data.data.access_token,
            refreshToken: data.data.refresh_token,
            expiresIn: data.data.expires_in,
            tokenType: data.data.token_type,
          });

          // 4. 同步 Store
          const userInfo = AuthUtils.getUserInfo();
          const store = useAuthStore.getState();
          store.setUser(userInfo);
          store.setAuthenticated(true);
          store.setLoading(false);
          return true;
        }

        // 刷新失败：清理本地令牌以打断后续尝试
        this.clearAuthState();
        AuthUtils.clearAuthInfo();
        return false;
      } catch (error) {
        console.error('❌ [AuthCore] Refresh failed:', error);
        // 异常也做彻底清理，避免无限刷新
        this.clearAuthState();
        AuthUtils.clearAuthInfo();
        return false;
      } finally {
        this.refreshInFlight = null;
      }
    })();

    return this.refreshInFlight;
  }

  /**
   * 检查认证状态
   * 业务流程：检查 token 有效性 → 如需要则刷新
   */
  static async checkAuthStatus(): Promise<boolean> {
    console.log('🔍 [AuthCore] Checking auth status...');

    const isAuthenticated = AuthUtils.isAuthenticated();

    if (isAuthenticated) {
      return true;
    }

    // Token 过期，尝试刷新
    const hasRefreshToken = AuthUtils.getRefreshToken();
    if (hasRefreshToken) {
      return await this.refreshToken();
    }

    return false;
  }

  /**
   * 处理登录回调
   * 业务流程：解析 URL 参数 → 保存 token → 更新状态 → 重定向
   */
  static async handleLoginCallback(params: {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
    userInfo?: UserInfo;
  }): Promise<void> {
    console.log('🎉 [AuthCore] Handling login callback...');

    const store = useAuthStore.getState();

    try {
      // 1. 保存认证信息到本地存储
      AuthUtils.saveAuthInfo(
        {
          accessToken: params.accessToken,
          refreshToken: params.refreshToken,
          expiresIn: params.expiresIn,
        },
        params.userInfo
      );

      // 2. 更新 Store 状态
      store.setUser(params.userInfo || null);
      store.setAuthenticated(true);
      store.setLoading(false);

      console.log('✅ [AuthCore] Login callback handled successfully');
    } catch (error) {
      console.error('❌ [AuthCore] Failed to handle login callback:', error);
      this.clearAuthState();
      throw error;
    }
  }

  /**
   * 获取当前用户信息
   */
  static getCurrentUser(): UserInfo | null {
    const store = useAuthStore.getState();
    return store.user;
  }

  /**
   * 判断是否是管理员
   */
  static isAdmin(): boolean {
    const user = this.getCurrentUser();
    return user?.type === 'admin';
  }

  /**
   * 私有方法：清除认证状态
   */
  private static clearAuthState(): void {
    const store = useAuthStore.getState();
    store.setUser(null);
    store.setAuthenticated(false);
    store.setLoading(false);
  }
}
