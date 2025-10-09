import { config } from '@/configs';

// OAuth2 相关配置
const oauth2 = {
  // 后端登录地址
  loginUrl: `${config.apiBaseUrl}/api/auth/login`,
  // 后端刷新token地址
  refreshUrl: `${config.apiBaseUrl}/api/auth/refresh`,
};

// 本地存储 key
const storage = {
  accessToken: 'access_token',
  refreshToken: 'refresh_token',
  tokenExpiry: 'token_expiry',
  userInfo: 'user_info',
};

// 用户信息类型（匹配后端 UserData 结构）
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

// Token 信息类型
export interface TokenInfo {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  tokenType?: string;
}

// 认证工具类
export class AuthUtils {
  // 检查是否已登录
  static isAuthenticated(): boolean {
    if (typeof window === 'undefined') return false;

    const token = localStorage.getItem(storage.accessToken);
    const expiry = localStorage.getItem(storage.tokenExpiry);

    if (!token || !expiry) return false;

    // 检查 token 是否过期
    const now = Date.now();
    const expiryTime = parseInt(expiry, 10);

    return now < expiryTime;
  }

  // 保存认证信息
  static saveAuthInfo(tokenInfo: TokenInfo, userInfo?: UserInfo): void {
    if (typeof window === 'undefined') return;

    localStorage.setItem(storage.accessToken, tokenInfo.accessToken);
    localStorage.setItem(storage.refreshToken, tokenInfo.refreshToken);

    // 计算过期时间（提前5分钟过期，避免边界情况）
    const expiryTime = Date.now() + tokenInfo.expiresIn * 1000 - 5 * 60 * 1000;
    localStorage.setItem(storage.tokenExpiry, expiryTime.toString());

    if (userInfo) {
      localStorage.setItem(storage.userInfo, JSON.stringify(userInfo));
    }
  }

  // 获取存储的访问令牌
  static getAccessToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(storage.accessToken);
  }

  // 获取存储的刷新令牌
  static getRefreshToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(storage.refreshToken);
  }

  // 获取用户信息
  static getUserInfo(): UserInfo | null {
    if (typeof window === 'undefined') return null;

    const userInfoStr = localStorage.getItem(storage.userInfo);
    if (!userInfoStr) return null;

    try {
      return JSON.parse(userInfoStr);
    } catch (error) {
      console.error('Failed to parse user info:', error);
      return null;
    }
  }

  // 清除认证信息
  static clearAuthInfo(): void {
    if (typeof window === 'undefined') return;

    localStorage.removeItem(storage.accessToken);
    localStorage.removeItem(storage.refreshToken);
    localStorage.removeItem(storage.tokenExpiry);
    localStorage.removeItem(storage.userInfo);
  }

  // 重定向到登录页
  static redirectToLogin(returnUrl?: string): void {
    if (typeof window === 'undefined') return;

    // 保存登录前的页面路径到 sessionStorage
    if (returnUrl) {
      console.log('💾 Saving to sessionStorage - login_return_url:', returnUrl);
      sessionStorage.setItem('login_return_url', returnUrl);
      console.log(
        '✅ Saved. Verifying:',
        sessionStorage.getItem('login_return_url')
      );
    }

    window.location.href = oauth2.loginUrl;
  }

  // 刷新 token
  static async refreshToken(): Promise<boolean> {
    if (typeof window === 'undefined') return false;

    const refreshToken = this.getRefreshToken();
    if (!refreshToken) {
      this.clearAuthInfo();
      return false;
    }

    try {
      const response = await fetch(oauth2.refreshUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refresh_token: refreshToken,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to refresh token');
      }

      const data = await response.json();

      if (data.code === 0 && data.data) {
        this.saveAuthInfo({
          accessToken: data.data.access_token,
          refreshToken: data.data.refresh_token,
          expiresIn: data.data.expires_in,
          tokenType: data.data.token_type,
        });

        return true;
      } else {
        throw new Error(data.error?.msg || 'Token refresh failed');
      }
    } catch (error) {
      console.error('Token refresh failed:', error);
      this.clearAuthInfo();
      return false;
    }
  }

  // 检查并刷新token（如果需要）
  static async ensureValidToken(): Promise<boolean> {
    if (this.isAuthenticated()) {
      return true;
    }

    // 尝试刷新 token
    const refreshed = await this.refreshToken();
    if (refreshed) {
      return true;
    }

    // 刷新失败，重定向到登录
    this.redirectToLogin();
    return false;
  }

  // 登出
  static logout(): void {
    this.clearAuthInfo();
    // 重定向到首页或登录页
    if (typeof window !== 'undefined') {
      window.location.href = '/';
    }
  }
}

// HTTP 请求工具（带认证）
export class ApiClient {
  private static async makeRequest(
    url: string,
    options: RequestInit = {}
  ): Promise<Response> {
    // 确保有有效的 token
    await AuthUtils.ensureValidToken();

    const token = AuthUtils.getAccessToken();
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    return fetch(url, {
      ...options,
      headers,
    });
  }

  static async get(url: string, options: RequestInit = {}): Promise<Response> {
    return this.makeRequest(url, { ...options, method: 'GET' });
  }

  static async post(
    url: string,
    data?: unknown,
    options: RequestInit = {}
  ): Promise<Response> {
    return this.makeRequest(url, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  static async put(
    url: string,
    data?: unknown,
    options: RequestInit = {}
  ): Promise<Response> {
    return this.makeRequest(url, {
      ...options,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  static async delete(
    url: string,
    options: RequestInit = {}
  ): Promise<Response> {
    return this.makeRequest(url, { ...options, method: 'DELETE' });
  }
}
