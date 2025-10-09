// 应用配置

const isProduction = import.meta.env.PROD;

console.log('Current Mode (NODE_ENV):', import.meta.env.MODE);
console.log('Is Production:', isProduction);

function getApiBaseUrl() {
  // 1. 环境变量优先级最高
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL;
  }
  // 2. 生产环境根据当前域名
  if (isProduction && typeof window !== 'undefined') {
    return `${window.location.protocol}//${window.location.host}`;
  }
  // 3. 开发环境或非浏览器环境回退到 localhost
  return 'http://localhost:48197';
}

function getFrontendBaseUrl() {
  // 1. 环境变量优先级最高
  if (import.meta.env.VITE_FRONTEND_BASE_URL) {
    return import.meta.env.VITE_FRONTEND_BASE_URL;
  }
  // 2. 生产环境根据当前域名
  if (isProduction && typeof window !== 'undefined') {
    return `${window.location.protocol}//${window.location.host}`;
  }
  // 3. 开发环境或非浏览器环境回退到 localhost
  return 'http://localhost:32234';
}

export const config = {
  // API 基础地址
  apiBaseUrl: getApiBaseUrl(),

  // 前端基础地址
  frontendBaseUrl: getFrontendBaseUrl(),
} as const;

console.log('Config:', config);

export type Config = typeof config;
