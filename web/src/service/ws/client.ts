import { config } from "@/configs";
import { AuthUtils } from "@/utils/auth";
import type { Options } from "react-use-websocket";

/**
 * 获取 WebSocket URL（自动处理 http/https 到 ws/wss 的转换）
 */
export function getWebSocketUrl(path: string): string {
  const wsProtocol = config.apiBaseUrl.startsWith("https") ? "wss" : "ws";
  const baseUrl = config.apiBaseUrl.replace(/^https?:/, "");
  return `${wsProtocol}:${baseUrl}${path}`;
}

/**
 * 由 Core 注入的依赖（与 http/client.ts 保持一致）
 */
export interface WsClientDeps {
  getAccessToken?: () => string | null;
}

let deps: WsClientDeps = {};

/**
 * 配置 WebSocket 客户端的依赖（在应用启动时由 Core 调用）
 */
export function configureWsClient(injected: WsClientDeps) {
  deps = injected;
}

/**
 * 获取带认证的 WebSocket 配置选项
 *
 * 注意：原生 WebSocket API 不支持自定义 headers，
 * 所以我们需要通过 URL query 参数传递 token
 */
export function getAuthenticatedWsOptions(
  baseOptions?: Partial<Options>,
): Options {
  return {
    ...baseOptions,
    // 在连接时添加 token 到 URL
    queryParams: {
      ...baseOptions?.queryParams,
    },
    // 或者使用 protocols 传递 token（某些服务端支持）
    protocols: baseOptions?.protocols,
    shouldReconnect: (closeEvent) => {
      // 401/403 不重连，其他情况可以重连
      if (closeEvent.code === 401 || closeEvent.code === 403) {
        return false;
      }
      return baseOptions?.shouldReconnect?.(closeEvent) ?? true;
    },
    reconnectAttempts: baseOptions?.reconnectAttempts ?? 3,
    reconnectInterval: baseOptions?.reconnectInterval ?? 3000,
  };
}

/**
 * 获取带 token 的 WebSocket URL（通过 query 参数）
 * 注意：使用 access_token_v2 作为参数名，与后端 auth 中间件保持一致
 */
export function getAuthenticatedWsUrl(path: string): string {
  const baseUrl = getWebSocketUrl(path);
  const token = deps.getAccessToken?.() || AuthUtils.getAccessToken();

  if (!token) {
    console.warn("⚠️ No token available for WebSocket connection");
    return baseUrl;
  }

  // 将 token 添加到 URL query 参数（使用 access_token_v2，与后端保持一致）
  const url = new URL(baseUrl);
  url.searchParams.set("access_token_v2", `Bearer ${token}`);

  return url.toString();
}
