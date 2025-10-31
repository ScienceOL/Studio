import { config } from '@/configs';
import apiClient from '@/service/http/client';

// 认证相关服务（纯 HTTP 调用，不含业务逻辑）
export const authService = {
  // 刷新 token（无需携带 Authorization）
  async refreshToken(refreshToken: string) {
    const res = await apiClient.post(`${config.apiBaseUrl}/api/auth/refresh`, {
      refresh_token: refreshToken,
    });
    return res.data;
  },

  // 可扩展：交换 code 获取 token
  async exchangeCode(params: { code: string; state?: string }) {
    const res = await apiClient.post(
      `${config.apiBaseUrl}/api/auth/code`,
      params
    );
    return res.data;
  },

  // 可扩展：获取当前用户信息（需要鉴权，按需在外部保证 token 有效）
  async getProfile() {
    const res = await apiClient.get(`${config.apiBaseUrl}/api/auth/me`);
    return res.data;
  },
};
