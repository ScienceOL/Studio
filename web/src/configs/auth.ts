// 应用配置
export const config = {
  // API 基础地址
  apiBaseUrl: import.meta.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:48197',

  // 前端基础地址
  frontendBaseUrl:
    import.meta.env.NEXT_PUBLIC_FRONTEND_BASE_URL || 'http://localhost:32234',

  // OAuth2 相关配置
  oauth2: {
    // 后端登录地址
    loginUrl: 'http://localhost:48197/api/auth/login',
    // 后端回调地址
    callbackUrl: 'http://localhost:48197/api/auth/callback/casdoor',
    // 后端刷新token地址
    refreshUrl: 'http://localhost:48197/api/auth/refresh',
  },

  // 本地存储 key
  storage: {
    accessToken: 'access_token',
    refreshToken: 'refresh_token',
    tokenExpiry: 'token_expiry',
    userInfo: 'user_info',
  },
} as const;

export type Config = typeof config;
