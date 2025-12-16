import { config } from "@/configs";
import { AuthUtils } from "@/utils/auth";
import type {
  AxiosError,
  AxiosHeaderValue,
  AxiosInstance,
  AxiosRequestConfig,
} from "axios";
import axios, { AxiosHeaders } from "axios";

// 由 Core 注入的依赖，避免 Service → Core 反向依赖导致循环
export interface ApiClientDeps {
  getAccessToken?: () => string | null;
  refreshToken?: () => Promise<boolean>; // Core 内部已做并发去重
  onAuthFailure?: () => void; // e.g. 登出/跳转
}

type RetriableRequestConfig = AxiosRequestConfig & { _retry?: boolean };

let deps: ApiClientDeps = {};

export const apiClient: AxiosInstance = axios.create({
  baseURL: config.apiBaseUrl,
  // withCredentials: false, // 使用 Authorization Bearer，不依赖 Cookie
  timeout: 30_000,
  headers: {
    "Content-Type": "application/json",
  },
});

/**
 * 配置客户端的依赖（在应用启动时由 Core 调用）
 */
export function configureApiClient(injected: ApiClientDeps) {
  deps = injected;
}

// 工具：将任意 headers 规范化为 AxiosHeaders
function toAxiosHeaders(h: unknown): AxiosHeaders {
  if (h instanceof AxiosHeaders) return h;
  const ax = new AxiosHeaders();
  if (h && typeof h === "object") {
    for (const [k, v] of Object.entries(h as Record<string, unknown>)) {
      const vv = v as AxiosHeaderValue | undefined;
      if (typeof vv !== "undefined") ax.set(k, vv);
    }
  }
  return ax;
}

// 请求拦截：附加 Authorization 头
apiClient.interceptors.request.use((request) => {
  console.group(
    `🚀 API Request: ${request.method?.toUpperCase()} ${request.url}`,
  );
  console.log("📍 Full URL:", `${request.baseURL}${request.url}`);
  console.log("📋 Headers:", request.headers);
  console.log("📦 Data:", request.data);
  console.log("🔍 Params:", request.params);

  try {
    const token = deps.getAccessToken?.() || AuthUtils.getAccessToken();
    if (token) {
      console.log("🔑 Token found:", token.substring(0, 20) + "...");
      const axHeaders = toAxiosHeaders(request.headers);
      axHeaders.set("Authorization", `Bearer ${token}`);
      request.headers = axHeaders;
      console.log("✅ Authorization header added");
    } else {
      console.log("⚠️ No token available");
    }
  } catch (err) {
    console.error("❌ Error adding token:", err);
    // 静默失败，交由服务端处理未鉴权
  }

  console.groupEnd();
  return request;
});

// 响应拦截：统一处理 401/403，尝试刷新并重放原请求
apiClient.interceptors.response.use(
  (response) => {
    console.group(
      `✅ API Response: ${response.config.method?.toUpperCase()} ${
        response.config.url
      }`,
    );
    console.log("📊 Status:", response.status, response.statusText);
    console.log("📋 Headers:", response.headers);
    console.log("📦 Data:", response.data);
    console.groupEnd();
    return response;
  },
  async (error: AxiosError) => {
    console.group(
      `❌ API Error: ${error.config?.method?.toUpperCase()} ${
        error.config?.url
      }`,
    );
    console.log("📊 Status:", error.response?.status);
    console.log("📋 Response Headers:", error.response?.headers);
    console.log("📦 Response Data:", error.response?.data);
    console.log("🔍 Error Message:", error.message);
    console.log("🌐 Network Error:", !error.response);
    console.groupEnd();

    const status = error.response?.status;
    const original = error.config as RetriableRequestConfig | undefined;

    // 不具备必要信息或未配置依赖，直接透传错误
    if (!original || !(status === 401 || status === 403)) {
      return Promise.reject(error);
    }

    // 防止在刷新接口自身或登录接口上循环重试
    const url = (original.url || "").toString();
    if (url.includes("/api/auth/refresh") || url.includes("/api/auth/login")) {
      return Promise.reject(error);
    }

    // 已经重试过一次则失败并触发 onAuthFailure
    if (original._retry) {
      deps.onAuthFailure?.();
      return Promise.reject(error);
    }

    original._retry = true;

    try {
      const ok = await deps.refreshToken?.();
      if (!ok) {
        deps.onAuthFailure?.();
        return Promise.reject(error);
      }

      // 刷新成功，读取最新 token 并重放原请求
      const newToken = deps.getAccessToken?.() || AuthUtils.getAccessToken();
      if (newToken) {
        const axHeaders = toAxiosHeaders(original.headers);
        axHeaders.set("Authorization", `Bearer ${newToken}`);
        original.headers = axHeaders;
      } else {
        // 无新 token 也视为失败
        deps.onAuthFailure?.();
        return Promise.reject(error);
      }

      return apiClient(original);
    } catch (e) {
      deps.onAuthFailure?.();
      return Promise.reject(e);
    }
  },
);

export default apiClient;
