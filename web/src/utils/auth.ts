import { config } from '@/configs';

// 本地存储 key
const storage = {
  accessToken: 'access_token',
  refreshToken: 'refresh_token',
  tokenExpiry: 'token_expiry',
  userInfo: 'user_info',
} as const;

// 用户信息类型（与后端返回结构对齐即可）
export interface UserInfo {
  id: string;
  name: string;
  owner: string;
  email: string;
  displayName: string;
  avatar: string;
  type: string;
  signupApplication: string;
  accessToken?: string;
  accessKey?: string;
  accessSecret?: string;
  phone?: string;
  status?: number;
  user_no?: string;
}

export interface TokenInfo {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  tokenType?: string;
}

export const AuthUtils = {
  // 检查是否已登录（仅根据本地 token 与过期时间）
  isAuthenticated(): boolean {
    if (typeof window === 'undefined') return false;
    const token = localStorage.getItem(storage.accessToken);
    const expiry = localStorage.getItem(storage.tokenExpiry);
    if (!token || !expiry) return false;
    const now = Date.now();
    const expiryTime = parseInt(expiry, 10);
    return now < expiryTime;
  },

  // 保存认证信息
  saveAuthInfo(tokenInfo: TokenInfo, userInfo?: UserInfo): void {
    if (typeof window === 'undefined') return;
    localStorage.setItem(storage.accessToken, tokenInfo.accessToken);
    localStorage.setItem(storage.refreshToken, tokenInfo.refreshToken);
    const expiryTime = Date.now() + tokenInfo.expiresIn * 1000 - 5 * 60 * 1000;
    localStorage.setItem(storage.tokenExpiry, expiryTime.toString());
    if (userInfo) {
      localStorage.setItem(storage.userInfo, JSON.stringify(userInfo));
    }
  },

  getAccessToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(storage.accessToken);
  },

  getRefreshToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(storage.refreshToken);
  },

  getUserInfo(): UserInfo | null {
    if (typeof window === 'undefined') return null;
    const userInfoStr = localStorage.getItem(storage.userInfo);
    if (!userInfoStr) return null;
    try {
      return JSON.parse(userInfoStr) as UserInfo;
    } catch (e) {
      console.error('Failed to parse user info:', e);
      return null;
    }
  },

  clearAuthInfo(): void {
    if (typeof window === 'undefined') return;
    localStorage.removeItem(storage.accessToken);
    localStorage.removeItem(storage.refreshToken);
    localStorage.removeItem(storage.tokenExpiry);
    localStorage.removeItem(storage.userInfo);
  },

  // 重定向到登录页（带前端回调）
  redirectToLogin(returnUrl?: string): void {
    if (typeof window === 'undefined') return;
    if (returnUrl) {
      sessionStorage.setItem('login_return_url', returnUrl);
    }
    const frontendCallbackURL = `${config.frontendBaseUrl}/login/callback`;
    const loginUrlWithCallback = `${
      config.apiBaseUrl
    }/api/auth/login?frontend_callback_url=${encodeURIComponent(
      frontendCallbackURL
    )}`;
    window.location.href = loginUrlWithCallback;
  },
};
