import { config } from '@/configs';

// 认证相关服务（纯 HTTP 调用，不含业务逻辑）
export const authService = {
  // 刷新 token（无需携带 Authorization）
  async refreshToken(refreshToken: string) {
    const res = await fetch(`${config.apiBaseUrl}/api/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (!res.ok) throw new Error('Failed to refresh token');
    return res.json();
  },

  // 可扩展：交换 code 获取 token
  async exchangeCode(params: { code: string; state?: string }) {
    const res = await fetch(`${config.apiBaseUrl}/api/auth/code`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(params),
    });
    if (!res.ok) throw new Error('Failed to exchange code');
    return res.json();
  },

  // 可扩展：获取当前用户信息（需要鉴权，按需在外部保证 token 有效）
  async getProfile() {
    const res = await fetch(`${config.apiBaseUrl}/api/auth/me`, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
    });
    if (!res.ok) throw new Error('Failed to get profile');
    return res.json();
  },
};
