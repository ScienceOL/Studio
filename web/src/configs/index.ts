// 应用配置
export const config = {
  // API 基础地址
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:48197',

  // 前端基础地址
  frontendBaseUrl:
    import.meta.env.VITE_FRONTEND_BASE_URL || 'http://localhost:32234',
} as const;

export type Config = typeof config;
